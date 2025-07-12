package v1alpha1

import (
	"testing"
)

func TestClusterUrlMonitor_GetResolvedDomain(t *testing.T) {
	tests := []struct {
		name               string
		serviceMonitorType string
		domainRef          ClusterDomainRef
		expectedDomain     ClusterDomainRef
	}{
		{
			name:               "RHOBS + unset DomainRef -> Infra",
			serviceMonitorType: ServiceMonitorTypeRHOBS,
			domainRef:          "",
			expectedDomain:     ClusterDomainRefInfra,
		},
		{
			name:               "CoreOS + unset DomainRef -> Infra",
			serviceMonitorType: ServiceMonitorTypeCoreOS,
			domainRef:          "",
			expectedDomain:     ClusterDomainRefInfra,
		},
		{
			name:               "unset ServiceMonitorType + unset DomainRef -> Infra",
			serviceMonitorType: "",
			domainRef:          "",
			expectedDomain:     ClusterDomainRefInfra,
		},
		{
			name:               "RHOBS + HCP DomainRef -> HCP",
			serviceMonitorType: ServiceMonitorTypeRHOBS,
			domainRef:          ClusterDomainRefHCP,
			expectedDomain:     ClusterDomainRefHCP,
		},
		{
			name:               "CoreOS + HCP DomainRef -> HCP",
			serviceMonitorType: ServiceMonitorTypeCoreOS,
			domainRef:          ClusterDomainRefHCP,
			expectedDomain:     ClusterDomainRefHCP,
		},
		{
			name:               "unset ServiceMonitorType + HCP DomainRef -> HCP",
			serviceMonitorType: "",
			domainRef:          ClusterDomainRefHCP,
			expectedDomain:     ClusterDomainRefHCP,
		},
		{
			name:               "RHOBS + Infra DomainRef -> Infra",
			serviceMonitorType: ServiceMonitorTypeRHOBS,
			domainRef:          ClusterDomainRefInfra,
			expectedDomain:     ClusterDomainRefInfra,
		},
		{
			name:               "CoreOS + Infra DomainRef -> Infra",
			serviceMonitorType: ServiceMonitorTypeCoreOS,
			domainRef:          ClusterDomainRefInfra,
			expectedDomain:     ClusterDomainRefInfra,
		},
		{
			name:               "unset ServiceMonitorType + Infra DomainRef -> Infra",
			serviceMonitorType: "",
			domainRef:          ClusterDomainRefInfra,
			expectedDomain:     ClusterDomainRefInfra,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cum := &ClusterUrlMonitor{
				Spec: ClusterUrlMonitorSpec{
					ServiceMonitorType: tt.serviceMonitorType,
					DomainRef:          tt.domainRef,
				},
			}

			result := cum.GetResolvedDomain()
			if result != tt.expectedDomain {
				t.Errorf("GetResolvedDomain() = %v, expected %v", result, tt.expectedDomain)
			}
		})
	}
}

func TestClusterUrlMonitor_GetResolvedServiceMonitorType(t *testing.T) {
	tests := []struct {
		name                       string
		serviceMonitorType         string
		domainRef                  ClusterDomainRef
		expectedServiceMonitorType string
	}{
		{
			name:                       "RHOBS + unset DomainRef -> RHOBS",
			serviceMonitorType:         ServiceMonitorTypeRHOBS,
			domainRef:                  "",
			expectedServiceMonitorType: ServiceMonitorTypeRHOBS,
		},
		{
			name:                       "CoreOS + unset DomainRef -> CoreOS",
			serviceMonitorType:         ServiceMonitorTypeCoreOS,
			domainRef:                  "",
			expectedServiceMonitorType: ServiceMonitorTypeCoreOS,
		},
		{
			name:                       "unset ServiceMonitorType + unset DomainRef -> CoreOS",
			serviceMonitorType:         "",
			domainRef:                  "",
			expectedServiceMonitorType: ServiceMonitorTypeCoreOS,
		},
		{
			name:                       "RHOBS + HCP DomainRef -> RHOBS",
			serviceMonitorType:         ServiceMonitorTypeRHOBS,
			domainRef:                  ClusterDomainRefHCP,
			expectedServiceMonitorType: ServiceMonitorTypeRHOBS,
		},
		{
			name:                       "CoreOS + HCP DomainRef -> CoreOS",
			serviceMonitorType:         ServiceMonitorTypeCoreOS,
			domainRef:                  ClusterDomainRefHCP,
			expectedServiceMonitorType: ServiceMonitorTypeCoreOS,
		},
		{
			name:                       "unset ServiceMonitorType + HCP DomainRef -> RHOBS",
			serviceMonitorType:         "",
			domainRef:                  ClusterDomainRefHCP,
			expectedServiceMonitorType: ServiceMonitorTypeRHOBS,
		},
		{
			name:                       "RHOBS + Infra DomainRef -> RHOBS",
			serviceMonitorType:         ServiceMonitorTypeRHOBS,
			domainRef:                  ClusterDomainRefInfra,
			expectedServiceMonitorType: ServiceMonitorTypeRHOBS,
		},
		{
			name:                       "CoreOS + Infra DomainRef -> CoreOS",
			serviceMonitorType:         ServiceMonitorTypeCoreOS,
			domainRef:                  ClusterDomainRefInfra,
			expectedServiceMonitorType: ServiceMonitorTypeCoreOS,
		},
		{
			name:                       "unset ServiceMonitorType + Infra DomainRef -> CoreOS",
			serviceMonitorType:         "",
			domainRef:                  ClusterDomainRefInfra,
			expectedServiceMonitorType: ServiceMonitorTypeCoreOS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cum := &ClusterUrlMonitor{
				Spec: ClusterUrlMonitorSpec{
					ServiceMonitorType: tt.serviceMonitorType,
					DomainRef:          tt.domainRef,
				},
			}

			result := cum.GetResolvedServiceMonitorType()
			if result != tt.expectedServiceMonitorType {
				t.Errorf("GetResolvedServiceMonitorType() = %v, expected %v", result, tt.expectedServiceMonitorType)
			}
		})
	}
}
