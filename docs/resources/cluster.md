# Cluster

A Kubernetes cluster. Its control plane lives at exactly one
[Site](site.md) (`controlPlaneSiteRef`); its worker [Nodes](node.md) and
[NetworkDevices](networkdevice.md) may live at other Sites, so the cluster's
full geographic footprint is derived from those child assets rather than this
field.

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Cluster`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `displayName` | string | yes | Human-readable name (min length 1). |
| `controlPlaneSiteRef` | LocalObjectReference | yes | [Site](site.md) hosting the API server. **Immutable.** |
| `role` | enum | yes | One of `Compute`, `Management`, `Edge`, `Gateway`. |
| `provider` | string | yes | Free-form provisioner identifier (e.g. `sidero-omni`, `gke`). Not interpreted. |
| `endpoint` | string | no | API server URL; must match `^https?://.+$`. |

## Status

- Condition `Ready`: `Accepted`; `SiteNotFound` while `controlPlaneSiteRef` is
  dangling.

## Validation & deletion

- `controlPlaneSiteRef` is immutable (CEL).
- Deletion is **rejected** while any [Node](node.md) (via its assignment) or
  [NetworkDevice](networkdevice.md) references the Cluster (`vcluster` webhook).

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Cluster
metadata:
  name: use1-compute-1
spec:
  displayName: US East 1 Compute
  controlPlaneSiteRef:
    name: us-east-1a
  role: Compute
  provider: sidero-omni
  endpoint: https://use1-compute-1.example.com:6443
```
