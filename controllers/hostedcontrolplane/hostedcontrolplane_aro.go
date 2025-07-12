/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hostedcontrolplane

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"github.com/openshift/route-monitor-operator/api/v1alpha1"
	"github.com/openshift/route-monitor-operator/pkg/util/finalizer"
	utilreconcile "github.com/openshift/route-monitor-operator/pkg/util/reconcile"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// aroHcpFinalizer defines the finalizer used by ARO-HCP controller's objects
	aroHcpFinalizer = "hostedcontrolplane.routemonitoroperator.aro-hcp.openshift.io/finalizer"

	// aroHcpWatchResourceLabel is a label key indicating which objects this controller should reconcile against
	aroHcpWatchResourceLabel = "hostedcontrolplane.routemonitoroperator.aro-hcp.openshift.io/managed"

	// aroHcpRouteMonitorName is the name for the RouteMonitor resource created for ARO-HCP
	aroHcpRouteMonitorName = "kas-monitor"

	// aroHcpKubeApiServerRouteName is the name of the kube-apiserver route to monitor
	aroHcpKubeApiServerRouteName = "kube-apiserver"

	// aroHcpKubeApiServerPort is the port number for the kube-apiserver route
	aroHcpKubeApiServerPort = 443

	// aroHcpHealthCheckSuffix is the health check endpoint suffix
	aroHcpHealthCheckSuffix = "/livez"

	// aroHcpTargetAvailabilityPercent is the SLO target availability percentage
	aroHcpTargetAvailabilityPercent = "99.95"
)

var aroLogger logr.Logger = ctrl.Log.WithName("controllers").WithName("HostedControlPlane-ARO")

// HostedControlPlaneAROReconciler reconciles a HostedControlPlane object for ARO-HCP environment
type HostedControlPlaneAROReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewHostedControlPlaneAROReconciler creates a HostedControlPlaneAROReconciler
func NewHostedControlPlaneAROReconciler(mgr manager.Manager) *HostedControlPlaneAROReconciler {
	return &HostedControlPlaneAROReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

//+kubebuilder:rbac:groups=hypershift.openshift.io,resources=hostedcontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hypershift.openshift.io,resources=hostedcontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hypershift.openshift.io,resources=hostedcontrolplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.openshift.io,resources=routemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile responds to events against watched HostedControlPlane objects for ARO-HCP environment
func (r *HostedControlPlaneAROReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := aroLogger.WithName("Reconcile").WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling HostedControlPlane for ARO-HCP")
	defer log.Info("Finished reconciling HostedControlPlane for ARO-HCP")

	// Fetch the HostedControlPlane instance
	hostedcontrolplane := &hypershiftv1beta1.HostedControlPlane{}
	err := r.Get(ctx, req.NamespacedName, hostedcontrolplane)
	if err != nil {
		if kerr.IsNotFound(err) {
			log.Info("HostedControlPlane not found, assumed deleted")
			return utilreconcile.Stop()
		}
		log.Error(err, "unable to fetch HostedControlPlane")
		return utilreconcile.RequeueWith(err)
	}

	// If the HostedControlPlane is marked for deletion, clean up
	shouldDelete := finalizer.WasDeleteRequested(hostedcontrolplane)
	if shouldDelete {
		err := r.finalizeHostedControlPlane(ctx, log, hostedcontrolplane)
		if err != nil {
			log.Error(err, "failed to finalize HostedControlPlane")
			return utilreconcile.RequeueWith(err)
		}

		finalizer.Remove(hostedcontrolplane, aroHcpFinalizer)
		err = r.Update(ctx, hostedcontrolplane)
		if err != nil {
			return utilreconcile.RequeueWith(err)
		}
		return utilreconcile.Stop()
	}

	// Add finalizer if not present
	if !finalizer.Contains(hostedcontrolplane.Finalizers, aroHcpFinalizer) {
		finalizer.Add(hostedcontrolplane, aroHcpFinalizer)
		err := r.Update(ctx, hostedcontrolplane)
		if err != nil {
			return utilreconcile.RequeueWith(err)
		}
	}

	// Check if HCP is ready using the existing health check logic
	hcpReconciler := &HostedControlPlaneReconciler{
		Client: r.Client,
		Scheme: r.Scheme,
	}

	err = hcpReconciler.hcpReady(ctx, hostedcontrolplane)
	if err != nil {
		log.Info(fmt.Sprintf("skipped deploying RouteMonitor, HostedControlPlane not ready: %v", err))
		return utilreconcile.RequeueAfter(healthcheckIntervalSeconds * time.Second), nil
	}

	log.Info("Deploying RouteMonitor for ARO-HCP")
	err = r.deployRouteMonitor(ctx, log, hostedcontrolplane)
	if err != nil {
		log.Error(err, "failed to deploy RouteMonitor")
		return utilreconcile.RequeueWith(err)
	}

	return ctrl.Result{}, nil
}

// deployRouteMonitor creates or updates the RouteMonitor needed to monitor the HCP's kube-apiserver
func (r *HostedControlPlaneAROReconciler) deployRouteMonitor(ctx context.Context, log logr.Logger, hostedcontrolplane *hypershiftv1beta1.HostedControlPlane) error {
	// Create or update RouteMonitor object
	expectedRouteMonitor := r.buildRouteMonitor(hostedcontrolplane)

	err := r.Create(ctx, &expectedRouteMonitor)
	if err != nil {
		if !kerr.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create RouteMonitor: %w", err)
		}
		// Object already exists: update it
		actualRouteMonitor := v1alpha1.RouteMonitor{}
		err := r.Get(ctx, types.NamespacedName{Name: expectedRouteMonitor.Name, Namespace: expectedRouteMonitor.Namespace}, &actualRouteMonitor)
		if err != nil {
			return fmt.Errorf("failed to retrieve RouteMonitor: %w", err)
		}

		// Update the RouteMonitor with expected values
		expectedRouteMonitor.ObjectMeta = buildMetadataForUpdate(expectedRouteMonitor.ObjectMeta, actualRouteMonitor.ObjectMeta)
		err = r.Update(ctx, &expectedRouteMonitor)
		if err != nil {
			return fmt.Errorf("failed to update RouteMonitor: %w", err)
		}
	}

	return nil
}

// buildRouteMonitor constructs the RouteMonitor object needed to monitor a HostedControlPlane's kube-apiserver
func (r *HostedControlPlaneAROReconciler) buildRouteMonitor(hostedcontrolplane *hypershiftv1beta1.HostedControlPlane) v1alpha1.RouteMonitor {
	routemonitor := v1alpha1.RouteMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            aroHcpRouteMonitorName,
			Namespace:       hostedcontrolplane.Namespace,
			OwnerReferences: buildOwnerReferences(hostedcontrolplane),
			Labels: map[string]string{
				aroHcpWatchResourceLabel: "true",
			},
		},
		Spec: v1alpha1.RouteMonitorSpec{
			Route: v1alpha1.RouteMonitorRouteSpec{
				NamespacedName: v1alpha1.NamespacedName{
					Name:      aroHcpKubeApiServerRouteName,
					Namespace: hostedcontrolplane.Namespace,
				},
				Port:   aroHcpKubeApiServerPort,
				Suffix: aroHcpHealthCheckSuffix,
			},
			CommonMonitorOptions: v1alpha1.CommonMonitorOptions{
				SkipPrometheusRule: false,
				Slo: v1alpha1.SloSpec{
					TargetAvailabilityPercent: aroHcpTargetAvailabilityPercent,
				},
				InsecureSkipTLSVerify: true,
			},
			ServiceMonitorType: v1alpha1.ServiceMonitorTypeCoreOS,
		},
	}
	return routemonitor
}

// finalizeHostedControlPlane cleans up HostedControlPlane-related objects managed by the ARO-HCP controller
func (r *HostedControlPlaneAROReconciler) finalizeHostedControlPlane(ctx context.Context, log logr.Logger, hostedcontrolplane *hypershiftv1beta1.HostedControlPlane) error {
	err := r.deleteRouteMonitor(ctx, log, hostedcontrolplane)
	if err != nil {
		return fmt.Errorf("failed to cleanup RouteMonitor: %w", err)
	}
	return nil
}

// deleteRouteMonitor removes the RouteMonitor object for the provided HostedControlPlane
func (r *HostedControlPlaneAROReconciler) deleteRouteMonitor(ctx context.Context, log logr.Logger, hostedcontrolplane *hypershiftv1beta1.HostedControlPlane) error {
	// Delete RouteMonitor
	expectedRouteMonitor := r.buildRouteMonitor(hostedcontrolplane)
	err := r.Delete(ctx, &expectedRouteMonitor)
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
		log.Info(fmt.Sprintf("Skipped deleting RouteMonitor %s/%s: already deleted", expectedRouteMonitor.Namespace, expectedRouteMonitor.Name))
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HostedControlPlaneAROReconciler) SetupWithManager(mgr ctrl.Manager) error {
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      aroHcpWatchResourceLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	selectorPredicate, err := predicate.LabelSelectorPredicate(selector)
	if err != nil {
		return fmt.Errorf("failed to build label selector predicate for RouteMonitors: %w", err)
	}

	// The following:
	// - Reconciles against all HostedControlPlane objects
	// - Additionally watches against RouteMonitor objects with the 'aroHcpWatchResourceLabel' present.
	//   When these objects are modified, the HCP specified in the objects' .metadata.OwnerReferences is
	//   reconciled
	return ctrl.NewControllerManagedBy(mgr).
		For(&hypershiftv1beta1.HostedControlPlane{}).
		Watches(
			&v1alpha1.RouteMonitor{},
			handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &hypershiftv1beta1.HostedControlPlane{}, handler.OnlyControllerOwner()),
			builder.WithPredicates(selectorPredicate),
		).
		Complete(r)
}
