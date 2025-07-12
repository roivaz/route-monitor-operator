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
	CommonMonitorOptions `json:",inline"`
	// +kubebuilder:validation:Enum=infra;hcp
	DomainRef ClusterDomainRef `json:"domainRef,omitempty"`
	// ServiceMonitorType dictates the type of ServiceMonitor the RouteMonitor should create
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=monitoring.coreos.com;monitoring.rhobs
	ServiceMonitorType string `json:"serviceMonitorType,omitempty"`
	// DomainExtractPattern is a regex pattern used to extract the valid domain from the cluster's base domain.
	// The pattern should contain one capture group that matches the desired domain part.
	// If not specified, defaults to extracting everything after the first subdomain (e.g., "rosa" or "api").
	// +kubebuilder:default:="^[^.]+\\.(.+)$"
	// +kubebuilder:validation:Optional
	DomainExtractPattern string `json:"domainExtractPattern,omitempty"`
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

func (c *ClusterUrlMonitorSpec) GetDomainExtractPattern() string {
	if c.DomainExtractPattern != "" {
		return c.DomainExtractPattern
	}
	return "^[^.]+\\.(.+)$"
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

// GetResolvedDomain returns the resolved domain based on the DomainRef field
// following the logic:
// - If DomainRef is set, use that value
// - If DomainRef is unset, always use Infra (regardless of ServiceMonitorType)
func (cum *ClusterUrlMonitor) GetResolvedDomain() ClusterDomainRef {
	// If DomainRef is explicitly set, use that value
	if cum.Spec.DomainRef != "" {
		return cum.Spec.DomainRef
	}

	// Default case: DomainRef is unset, always use Infra
	return ClusterDomainRefInfra
}

// GetResolvedServiceMonitorType returns the resolved service monitor type based on the ServiceMonitorType and DomainRef fields
// following the logic:
// - If ServiceMonitorType is set, use that value
// - If ServiceMonitorType is unset and DomainRef is HCP, use RHOBS
// - If ServiceMonitorType is unset and DomainRef is Infra or unset, use CoreOS
func (cum *ClusterUrlMonitor) GetResolvedServiceMonitorType() string {
	// If ServiceMonitorType is explicitly set, use that value
	if cum.Spec.ServiceMonitorType != "" {
		return cum.Spec.ServiceMonitorType
	}

	// ServiceMonitorType is unset, check DomainRef
	if cum.Spec.DomainRef == ClusterDomainRefHCP {
		return ServiceMonitorTypeRHOBS
	}

	// Default case: ServiceMonitorType is unset and DomainRef is Infra or unset
	return ServiceMonitorTypeCoreOS
}
