package servicemonitor

import (
	"context"
	"fmt"

	"github.com/openshift/route-monitor-operator/api/v1alpha1"
	"github.com/openshift/route-monitor-operator/controllers"
	"github.com/openshift/route-monitor-operator/pkg/consts/blackboxexporter"
	util "github.com/openshift/route-monitor-operator/pkg/reconcile"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	rhobsv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceMonitor struct {
	Client   client.Client
	Ctx      context.Context
	Comparer util.ResourceComparerInterface
}

func NewServiceMonitor(ctx context.Context, c client.Client) *ServiceMonitor {
	return &ServiceMonitor{
		Client:   c,
		Ctx:      ctx,
		Comparer: &util.ResourceComparer{},
	}
}

const (
	ServiceMonitorPeriod string = "30s"
	UrlLabelName         string = "probe_url"
)

func (u *ServiceMonitor) TemplateAndUpdateServiceMonitorDeployment(routeURL, blackBoxExporterNamespace string, namespacedName types.NamespacedName, clusterID string, smType controllers.ServiceMonitorType, useInsecure bool, owner *metav1.OwnerReference) error {
	module := "http_2xx"
	if useInsecure {
		module = "insecure_http_2xx"
	}

	params := map[string][]string{
		"module": {module},
		"target": {routeURL},
	}

	template, err := u.createServiceMonitorTemplate(routeURL, blackBoxExporterNamespace, params, namespacedName, clusterID, smType, owner)
	if err != nil {
		return err
	}

	return u.UpdateServiceMonitorDeployment(template)
}

// UpdateServiceMonitorDeployment creates or updates Service Monitor Deployment according to the template
func (u *ServiceMonitor) UpdateServiceMonitorDeployment(template client.Object) error {
	namespacedName := types.NamespacedName{Name: template.GetName(), Namespace: template.GetNamespace()}

	switch t := template.(type) {
	case *monitoringv1.ServiceMonitor:
		deployedServiceMonitor := &monitoringv1.ServiceMonitor{}
		err := u.Client.Get(u.Ctx, namespacedName, deployedServiceMonitor)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
			return u.Client.Create(u.Ctx, template)
		}
		if !u.Comparer.DeepEqual(deployedServiceMonitor.Spec, t.Spec) {
			deployedServiceMonitor.Spec = t.Spec
			return u.Client.Update(u.Ctx, deployedServiceMonitor)
		}
		return nil

	case *rhobsv1.ServiceMonitor:
		deployedServiceMonitor := &rhobsv1.ServiceMonitor{}
		err := u.Client.Get(u.Ctx, namespacedName, deployedServiceMonitor)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
			return u.Client.Create(u.Ctx, template)
		}
		if !u.Comparer.DeepEqual(deployedServiceMonitor.Spec, t.Spec) {
			deployedServiceMonitor.Spec = t.Spec
			return u.Client.Update(u.Ctx, deployedServiceMonitor)
		}
		return nil

	default:
		return fmt.Errorf("unsupported ServiceMonitor type: %T", template)
	}
}

// DeleteServiceMonitorDeployment deletes the ServiceMonitor Deployment
func (u *ServiceMonitor) DeleteServiceMonitorDeployment(serviceMonitorRef v1alpha1.NamespacedName, smType controllers.ServiceMonitorType) error {
	if serviceMonitorRef == (v1alpha1.NamespacedName{}) {
		return nil
	}
	namespacedName := types.NamespacedName{Name: serviceMonitorRef.Name, Namespace: serviceMonitorRef.Namespace}

	switch smType {
	case controllers.RhobsServiceMonitor:
		resource := &rhobsv1.ServiceMonitor{}
		err := u.Client.Get(u.Ctx, namespacedName, resource)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
			return nil
		}
		return u.Client.Delete(u.Ctx, resource)

	case controllers.CoreosServiceMonitor:
		resource := &monitoringv1.ServiceMonitor{}
		err := u.Client.Get(u.Ctx, namespacedName, resource)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
			return nil
		}
		return u.Client.Delete(u.Ctx, resource)

	default:
		return fmt.Errorf("unsupported ServiceMonitor type: %s", smType)
	}
}

// TemplateForServiceMonitorResource returns a ServiceMonitor
func (u *ServiceMonitor) TemplateForServiceMonitorResource(routeURL, blackBoxExporterNamespace string, params map[string][]string, namespacedName types.NamespacedName, clusterID string, owner *metav1.OwnerReference) monitoringv1.ServiceMonitor {
	return monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namespacedName.Name,
			Namespace:       namespacedName.Namespace,
			OwnerReferences: []metav1.OwnerReference{*owner},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port: blackboxexporter.BlackBoxExporterPortName,
					// Probe every 30s
					Interval: monitoringv1.Duration(ServiceMonitorPeriod),
					// Timeout has to be smaller than probe interval
					ScrapeTimeout: "15s",
					Path:          "/probe",
					Scheme:        "http",
					Params:        params,
					MetricRelabelConfigs: []*monitoringv1.RelabelConfig{
						{
							Replacement: routeURL,
							TargetLabel: UrlLabelName,
						},
						{
							Replacement: clusterID,
							TargetLabel: "_id",
						},
					},
				}},
			Selector: metav1.LabelSelector{
				MatchLabels: blackboxexporter.GenerateBlackBoxExporterLables(),
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					blackBoxExporterNamespace,
				},
			},
		},
	}
}

// HyperShiftTemplateForServiceMonitorResource returns a ServiceMonitor for Hypershift
func (u *ServiceMonitor) HyperShiftTemplateForServiceMonitorResource(routeURL, blackBoxExporterNamespace string, params map[string][]string, namespacedName types.NamespacedName, clusterID string, owner *metav1.OwnerReference) rhobsv1.ServiceMonitor {
	return rhobsv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namespacedName.Name,
			Namespace:       namespacedName.Namespace,
			OwnerReferences: []metav1.OwnerReference{*owner},
		},
		Spec: rhobsv1.ServiceMonitorSpec{
			Endpoints: []rhobsv1.Endpoint{
				{
					Port: blackboxexporter.BlackBoxExporterPortName,
					// Probe every 30s
					Interval: rhobsv1.Duration(ServiceMonitorPeriod),
					// Timeout has to be smaller than probe interval
					ScrapeTimeout: "15s",
					Path:          "/probe",
					Scheme:        "http",
					Params:        params,
					MetricRelabelConfigs: []*rhobsv1.RelabelConfig{
						{
							Replacement: routeURL,
							TargetLabel: UrlLabelName,
						},
						{
							Replacement: clusterID,
							TargetLabel: "_id",
						},
					},
				}},
			Selector: metav1.LabelSelector{
				MatchLabels: blackboxexporter.GenerateBlackBoxExporterLables(),
			},
			NamespaceSelector: rhobsv1.NamespaceSelector{
				MatchNames: []string{
					blackBoxExporterNamespace,
				},
			},
		},
	}
}

// createServiceMonitorTemplate creates a ServiceMonitor template based on the specified type
func (u *ServiceMonitor) createServiceMonitorTemplate(routeURL, blackBoxExporterNamespace string, params map[string][]string, namespacedName types.NamespacedName, clusterID string, smType controllers.ServiceMonitorType, owner *metav1.OwnerReference) (client.Object, error) {
	switch smType {
	case controllers.CoreosServiceMonitor:
		template := u.TemplateForServiceMonitorResource(routeURL, blackBoxExporterNamespace, params, namespacedName, clusterID, owner)
		return &template, nil

	case controllers.RhobsServiceMonitor:
		template := u.HyperShiftTemplateForServiceMonitorResource(routeURL, blackBoxExporterNamespace, params, namespacedName, clusterID, owner)
		return &template, nil

	default:
		return nil, fmt.Errorf("unsupported ServiceMonitor type: %s", smType)
	}
}
