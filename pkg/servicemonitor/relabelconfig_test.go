package servicemonitor

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	rhobsv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToRHOBSRelabelConfigs(t *testing.T) {
	tests := []struct {
		name     string
		input    []monitoringv1.RelabelConfig
		expected int
		wantErr  bool
	}{
		{
			name:     "empty input",
			input:    []monitoringv1.RelabelConfig{},
			expected: 0,
			wantErr:  false,
		},
		{
			name: "single relabel config",
			input: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "test_label",
					Replacement: "test_value",
					Action:      "replace",
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "multiple relabel configs",
			input: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "probe_url",
					Replacement: "https://example.com",
					Action:      "replace",
				},
				{
					TargetLabel: "_id",
					Replacement: "cluster-123",
					Action:      "replace",
				},
				{
					SourceLabels: []monitoringv1.LabelName{"__name__"},
					TargetLabel:  "__name__",
					Regex:        "probe_success",
					Replacement:  "custom_probe_success",
					Action:       "replace",
				},
			},
			expected: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToRHOBSRelabelConfigs(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, tt.expected)

			// Verify that the conversion preserved the important fields
			for i, original := range tt.input {
				assert.Equal(t, original.TargetLabel, result[i].TargetLabel)
				assert.Equal(t, original.Replacement, result[i].Replacement)
				assert.Equal(t, original.Action, result[i].Action)
				assert.Equal(t, original.Regex, result[i].Regex)
				// Compare SourceLabels lengths and contents instead of direct equality
				assert.Equal(t, len(original.SourceLabels), len(result[i].SourceLabels))
				for j, label := range original.SourceLabels {
					assert.Equal(t, string(label), string(result[i].SourceLabels[j]))
				}
			}
		})
	}
}

func TestGetEffectiveRelabelConfigs(t *testing.T) {
	tests := []struct {
		name           string
		routeURL       string
		clusterID      string
		customConfigs  []monitoringv1.RelabelConfig
		expectedLabels map[string]string
	}{
		{
			name:          "no custom configs - uses defaults",
			routeURL:      "https://example.com",
			clusterID:     "cluster-123",
			customConfigs: []monitoringv1.RelabelConfig{},
			expectedLabels: map[string]string{
				"probe_url": "https://example.com",
				"_id":       "cluster-123",
			},
		},
		{
			name:      "custom configs override defaults",
			routeURL:  "https://example.com",
			clusterID: "cluster-123",
			customConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "custom_label",
					Replacement: "custom_value",
					Action:      "replace",
				},
				{
					TargetLabel: "probe_url",
					Replacement: "https://custom.example.com",
					Action:      "replace",
				},
			},
			expectedLabels: map[string]string{
				"custom_label": "custom_value",
				"probe_url":    "https://custom.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEffectiveRelabelConfigs(tt.routeURL, tt.clusterID, tt.customConfigs, "monitoring.coreos.com")

			// Verify that we get the expected number of configs
			if len(tt.customConfigs) > 0 {
				assert.Len(t, result, len(tt.customConfigs))
			} else {
				assert.Len(t, result, 2) // Default probe_url and _id
			}

			// Verify that the expected labels are present
			labelMap := make(map[string]string)
			for _, config := range result {
				labelMap[config.TargetLabel] = config.Replacement
			}

			for expectedLabel, expectedValue := range tt.expectedLabels {
				assert.Equal(t, expectedValue, labelMap[expectedLabel], "Expected label %s to have value %s", expectedLabel, expectedValue)
			}
		})
	}
}

func TestGetEffectiveRHOBSRelabelConfigs(t *testing.T) {
	tests := []struct {
		name           string
		routeURL       string
		clusterID      string
		customConfigs  []monitoringv1.RelabelConfig
		expectedLabels map[string]string
	}{
		{
			name:          "no custom configs - uses defaults",
			routeURL:      "https://example.com",
			clusterID:     "cluster-123",
			customConfigs: []monitoringv1.RelabelConfig{},
			expectedLabels: map[string]string{
				"probe_url": "https://example.com",
				"_id":       "cluster-123",
			},
		},
		{
			name:      "custom configs converted from CoreOS",
			routeURL:  "https://example.com",
			clusterID: "cluster-123",
			customConfigs: []monitoringv1.RelabelConfig{
				{
					TargetLabel: "environment",
					Replacement: "production",
					Action:      "replace",
				},
			},
			expectedLabels: map[string]string{
				"environment": "production",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEffectiveRHOBSRelabelConfigs(tt.routeURL, tt.clusterID, tt.customConfigs)

			// Verify that the expected labels are present
			labelMap := make(map[string]string)
			for _, config := range result {
				labelMap[config.TargetLabel] = config.Replacement
			}

			for expectedLabel, expectedValue := range tt.expectedLabels {
				assert.Equal(t, expectedValue, labelMap[expectedLabel], "Expected label %s to have value %s", expectedLabel, expectedValue)
			}

			// Verify types are correct (should be rhobsv1.RelabelConfig)
			for _, config := range result {
				assert.IsType(t, &rhobsv1.RelabelConfig{}, config)
			}
		})
	}
}
