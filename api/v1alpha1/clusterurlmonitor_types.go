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

// ClusterUrlMonitorSpec defines the desired state of ClusterUrlMonitor
type ClusterUrlMonitorSpec struct {
	EnvironmentDefinition `json:",inline"`
	CommonMonitorOptions  `json:",inline"`

	// Prefix is prepended to the cluster domain when constructing the URL to monitor
	// +optional
	Prefix string `json:"prefix,omitempty"`
	// Suffix is appended to the cluster domain when constructing the URL to monitor
	// +optional
	Suffix string `json:"suffix,omitempty"`
	// Port specifies the port to use when constructing the URL to monitor
	// +optional
	Port string `json:"port,omitempty"`
}

// ClusterUrlMonitorStatus defines the observed state of ClusterUrlMonitor
type ClusterUrlMonitorStatus struct {
	// ServiceMonitorRef contains the reference to the ServiceMonitor created for this ClusterUrlMonitor
	// +optional
	ServiceMonitorRef NamespacedName `json:"serviceMonitorRef,omitempty"`
	// PrometheusRuleRef contains the reference to the PrometheusRule created for this ClusterUrlMonitor
	// +optional
	PrometheusRuleRef NamespacedName `json:"prometheusRuleRef,omitempty"`
	// ErrorStatus contains error information if the ClusterUrlMonitor is in an error state
	// +optional
	ErrorStatus string `json:"errorStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ClusterUrlMonitor is the Schema for the clusterurlmonitors API
type ClusterUrlMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterUrlMonitorSpec   `json:"spec,omitempty"`
	Status ClusterUrlMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterUrlMonitorList contains a list of ClusterUrlMonitor
type ClusterUrlMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterUrlMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterUrlMonitor{}, &ClusterUrlMonitorList{})
}
