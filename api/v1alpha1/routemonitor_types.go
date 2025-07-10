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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RouteMonitorSpec defines the desired state of RouteMonitor
type RouteMonitorSpec struct {
	EnvironmentDefinition `json:",inline"`
	CommonMonitorOptions  `json:",inline"`

	// Route specifies the Route resource that should be monitored
	Route RouteMonitorRouteSpec `json:"route"`
}

// RouteMonitorRouteSpec references the observed Route resource
type RouteMonitorRouteSpec struct {
	NamespacedName `json:",inline"`

	// Port optionally defines the port we should use while probing
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum:=1
	// +optional
	Port int64 `json:"port,omitempty"`

	// Suffix optionally defines the path we should probe (/livez /readyz etc)
	// +kubebuilder:validation:Optional
	// +optional
	Suffix string `json:"suffix,omitempty"`
}

// RouteMonitorStatus defines the observed state of RouteMonitor
type RouteMonitorStatus struct {
	// RouteURL is the url extracted from the Route resource
	// +optional
	RouteURL string `json:"routeURL,omitempty"`
	// ServiceMonitorRef contains the reference to the ServiceMonitor created for this RouteMonitor
	// +optional
	ServiceMonitorRef NamespacedName `json:"serviceMonitorRef,omitempty"`
	// PrometheusRuleRef contains the reference to the PrometheusRule created for this RouteMonitor
	// +optional
	PrometheusRuleRef NamespacedName `json:"prometheusRuleRef,omitempty"`
	// ErrorStatus contains error information if the RouteMonitor is in an error state
	// +optional
	ErrorStatus string `json:"errorStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RouteMonitor is the Schema for the routemonitors API
type RouteMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteMonitorSpec   `json:"spec,omitempty"`
	Status RouteMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteMonitorList contains a list of RouteMonitor
type RouteMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouteMonitor{}, &RouteMonitorList{})
}
