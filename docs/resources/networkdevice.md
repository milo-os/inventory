# NetworkDevice

A switch, router, or firewall that is part of a [Cluster](cluster.md) and
physically lives in a [Site](site.md).

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `NetworkDevice`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clusterRef` | LocalObjectReference | yes | [Cluster](cluster.md) the device is part of. **Immutable.** |
| `siteRef` | LocalObjectReference | yes | [Site](site.md) where the device lives. **Immutable.** |
| `role` | enum | yes | One of `BorderRouter`, `Spine`, `Leaf`, `Firewall`. |
| `managementAddress` | string | no | Address of the device's management plane. The inventory does not connect to it. |

## Status

- Condition `Ready`: `Accepted`; `ClusterNotFound` / `SiteNotFound` while a
  reference is dangling.

## Validation & deletion

- `clusterRef` and `siteRef` are immutable (CEL).
- On create/update the `vnetworkdevice` webhook **rejects** the request if the
  referenced Cluster does not exist. It does **not** require the device and
  cluster to share a Site — clusters can span Sites.
- No deletion guard — NetworkDevices are leaf assets.

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: NetworkDevice
metadata:
  name: use1-leaf-01
spec:
  clusterRef:
    name: use1-compute-1
  siteRef:
    name: us-east-1a
  role: Leaf
  managementAddress: 10.0.0.21
```
