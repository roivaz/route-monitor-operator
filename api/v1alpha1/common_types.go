package v1alpha1

import "gopkg.in/inf.v0"

// NamespacedName contains the name of a object and its namespace
type NamespacedName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

const (
	// The following values should match the kubebuilder-enumerated values for serviceMonitorType above
	ServiceMonitorTypeCoreOS = "monitoring.coreos.com"
	ServiceMonitorTypeRHOBS  = "monitoring.rhobs"
)

// ClusterDomainRef defines the object used determine the cluster's domain
// By default, 'infra' is used, which references the 'infrastructures/cluster' object
type ClusterDomainRef string

var (
	// ClusterDomainRefInfra indicates the clusterDomain should be determined from the 'infrastructures/cluster' object
	ClusterDomainRefInfra ClusterDomainRef = "infra"

	// ClusterDomainRefHCP indicates the clusterDomain should be determined from the 'hcp/cluster' object in the same namespace as the ClusterURLMonitor being reconciled
	ClusterDomainRefHCP ClusterDomainRef = "hcp"
)

type EnvironmentDefinition struct {
	// +kubebuilder:validation:Enum=infra;hcp
	// +kubebuilder:default:="infra"
	DomainRef ClusterDomainRef `json:"domainRef,omitempty"`
	// ServiceMonitorType dictates the type of ServiceMonitor the RouteMonitor should create
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=monitoring.coreos.com;monitoring.rhobs
	// +kubebuilder:default=monitoring.coreos.com
	ServiceMonitorType string `json:"serviceMonitorType,omitempty"`
	// DomainExtractPattern is a regex pattern used to extract the valid domain from the cluster's base domain.
	// The pattern should contain one capture group that matches the desired domain part.
	// If not specified, defaults to extracting everything after the first subdomain (e.g., "rosa" or "api").
	// +kubebuilder:default:="^[^.]+\\.(.+)$"
	// +kubebuilder:validation:Optional
	DomainExtractPattern string `json:"domainExtractPattern,omitempty"`
}

func (e *EnvironmentDefinition) GetDomainExtractPattern() string {
	if e.DomainExtractPattern != "" {
		return e.DomainExtractPattern
	}
	return "^[^.]+\\.(.+)$"
}

type CommonMonitorOptions struct {
	// Service level objective for the monitor
	Slo SloSpec `json:"slo,omitempty"`
	// SkipPrometheusRule instructs the controller to skip the creation of PrometheusRule CRs.
	// One common use-case for is for alerts that are defined separately, such as for hosted clusters.
	// +kubebuilder:default:false
	// +kubebuilder:validation:Optional
	SkipPrometheusRule bool `json:"skipPrometheusRule"`
	// InsecureSkipTLSVerify indicates that the blackbox exporter module used to probe this route
	// should *not* use https
	// +kubebuilder:default:false
	// +kubebuilder:validation:Optional
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify"`
}

// SloSpec defines what is the percentage
type SloSpec struct {
	// TargetAvailabilityPercent defines the percent number to be used
	TargetAvailabilityPercent string `json:"targetAvailabilityPercent"`
}

func (s SloSpec) IsValid() (bool, string) {
	if s.TargetAvailabilityPercent == "" {
		return false, ""
	}

	d, success := new(inf.Dec).SetString(s.TargetAvailabilityPercent)
	// value is not parsable
	if !success {
		return false, ""
	}

	// will be 90
	ninety := inf.NewDec(9, -1)
	// is lower than lower bound
	if d.Cmp(ninety) <= 0 {
		return false, ""
	}

	// will be 100
	hundred := inf.NewDec(1, -2)
	// is higher than upper bound
	if d.Cmp(hundred) >= 0 {
		return false, ""
	}

	// will be 1/100
	oneHundredth := inf.NewDec(1, 2)

	res := d.Mul(d, oneHundredth).String()

	return true, res
}
