# Custom Relabel Configurations

The Route Monitor Operator (RMO) now supports custom relabel configurations for ServiceMonitor resources. This feature allows users to customize how metrics are labeled and processed by Prometheus, providing greater flexibility for monitoring setups.

## Overview

By default, RMO creates ServiceMonitor resources with standard relabel configurations that add:
- `probe_url`: The target URL being monitored
- `_id`: The cluster ID

With the new `relabelConfigs` field, users can override these defaults entirely with their own relabel configurations.

## Usage

### Basic Example

```yaml
apiVersion: monitoring.openshift.io/v1alpha1
kind: RouteMonitor
metadata:
  name: my-route-monitor
spec:
  route:
    namespace: my-app
    name: api-server
  slo:
    targetAvailabilityPercent: "99.9"
  relabelConfigs:
    - targetLabel: "environment"
      replacement: "production"
      action: "replace"
    - targetLabel: "probe_url"
      replacement: "https://api-server-my-app.apps.cluster.com"
      action: "replace"
    - targetLabel: "_id"
      replacement: "my-cluster-id"
      action: "replace"
```

## Important Notes

### Complete Override Behavior

When you specify `relabelConfigs`, they **completely override** the default configurations. This means:

- ✅ **You have full control** over metric labeling
- ⚠️ **You must include all labels you need**, including `probe_url` and `_id` if required
- ⚠️ **Default labels are not automatically added**

### Compatibility

The `relabelConfigs` field uses the `monitoring.coreos.com/v1.RelabelConfig` structure. When creating RHOBS ServiceMonitors, RMO automatically converts these configurations to the appropriate format.

## Common Use Cases

### 1. Adding Custom Labels

Add environment, team, or application-specific labels:

```yaml
relabelConfigs:
  - targetLabel: "environment"
    replacement: "production"
    action: "replace"
  - targetLabel: "team"
    replacement: "platform"
    action: "replace"
  - targetLabel: "application"
    replacement: "api-gateway"
    action: "replace"
  # Don't forget the standard labels
  - targetLabel: "probe_url"
    replacement: "https://api.example.com"
    action: "replace"
  - targetLabel: "_id"
    replacement: "cluster-123"
    action: "replace"
```

### 2. Metric Filtering

Keep only specific metrics:

```yaml
relabelConfigs:
  # Only keep success and duration metrics
  - sourceLabels: ["__name__"]
    regex: "probe_(success|duration_seconds)"
    action: "keep"
  # Standard labels
  - targetLabel: "probe_url"
    replacement: "https://api.example.com"
    action: "replace"
  - targetLabel: "_id"
    replacement: "cluster-123"
    action: "replace"
```

### 3. Metric Renaming

Rename metrics for better organization:

```yaml
relabelConfigs:
  - sourceLabels: ["__name__"]
    targetLabel: "__name__"
    regex: "probe_success"
    replacement: "my_service_availability"
    action: "replace"
  - sourceLabels: ["__name__"]
    targetLabel: "__name__"
    regex: "probe_duration_seconds"
    replacement: "my_service_response_time"
    action: "replace"
  # Standard labels
  - targetLabel: "probe_url"
    replacement: "https://api.example.com"
    action: "replace"
  - targetLabel: "_id"
    replacement: "cluster-123"
    action: "replace"
```

### 4. Conditional Labels

Add labels based on metric values:

```yaml
relabelConfigs:
  # Add severity based on probe success
  - sourceLabels: ["probe_success"]
    targetLabel: "severity"
    regex: "0"
    replacement: "critical"
    action: "replace"
  - sourceLabels: ["probe_success"]
    targetLabel: "severity"
    regex: "1"
    replacement: "info"
    action: "replace"
  # Standard labels
  - targetLabel: "probe_url"
    replacement: "https://api.example.com"
    action: "replace"
  - targetLabel: "_id"
    replacement: "cluster-123"
    action: "replace"
```

## ClusterUrlMonitor Examples

The same feature works for ClusterUrlMonitor resources:

```yaml
apiVersion: monitoring.openshift.io/v1alpha1
kind: ClusterUrlMonitor
metadata:
  name: api-monitor-with-custom-labels
spec:
  domainRef: infra
  serviceMonitorType: monitoring.coreos.com
  prefix: "https://api."
  port: "6443"
  suffix: "/healthz"
  slo:
    targetAvailabilityPercent: "99.5"
  relabelConfigs:
    - targetLabel: "component"
      replacement: "kube-apiserver"
      action: "replace"
    - targetLabel: "service_type"
      replacement: "infrastructure"
      action: "replace"
    # Standard labels
    - targetLabel: "probe_url"
      replacement: "https://api.cluster.example.com:6443/healthz"
      action: "replace"
    - targetLabel: "_id"
      replacement: "cluster-123"
      action: "replace"
```

## ServiceMonitor Type Support

### CoreOS ServiceMonitors (`monitoring.coreos.com`)

Uses the relabel configs directly as specified.

### RHOBS ServiceMonitors (`monitoring.rhobs`)

RMO automatically converts the configurations to the RHOBS format using JSON marshalling/unmarshalling. The field structures are identical, so this conversion is transparent.

## Best Practices

### 1. Always Include Essential Labels

Make sure to include the essential labels that your monitoring system expects:

```yaml
relabelConfigs:
  # Your custom configs...
  - targetLabel: "custom_label"
    replacement: "custom_value"
    action: "replace"
  
  # Essential labels
  - targetLabel: "probe_url"
    replacement: "{{ your-target-url }}"
    action: "replace"
  - targetLabel: "_id"
    replacement: "{{ your-cluster-id }}"
    action: "replace"
```

### 2. Test Configurations

Test your relabel configurations in a development environment before applying to production.

### 3. Document Custom Labels

Document any custom labels you add for your team and future maintenance.

### 4. Consider Performance

Complex relabel configurations can impact Prometheus performance. Keep them as simple as possible while meeting your requirements.

## Migration from Default Labels

If you're migrating from the default RMO labels to custom ones:

1. **Start with the defaults**: Copy the current default behavior
2. **Add your custom labels**: Incrementally add new labels
3. **Test thoroughly**: Verify your monitoring and alerting still work
4. **Update dashboards**: Update any dashboards or alerts that depend on the labels

## Troubleshooting

### Problem: Metrics not appearing in Prometheus

**Possible causes:**
- Missing essential labels like `probe_url`
- Incorrect relabel configuration syntax
- Metrics being filtered out by relabel rules

**Solution:** Check the ServiceMonitor resource and verify your relabel configurations.

### Problem: RHOBS ServiceMonitor conversion errors

**Possible causes:**
- Using field names not supported by RHOBS
- Complex relabel configurations that don't convert cleanly

**Solution:** Simplify the relabel configuration or check RMO logs for conversion errors.

## Examples Repository

For more examples, check the `config/samples/` directory in the RMO repository:
- `monitoring_v1alpha1_routemonitor_with_relabelconfigs.yaml`
- `monitoring_v1alpha1_clusterurlmonitor_with_relabelconfigs.yaml`
