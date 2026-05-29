# Site

A physical or logical location within a [Region](region.md) — a datacenter,
availability zone, edge PoP, or virtual site. A Site is the parent of Nodes,
Clusters, and NetworkDevices.

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Site`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `displayName` | string | yes | Human-readable name (min length 1). |
| `regionRef` | LocalObjectReference | yes | The [Region](region.md) this Site belongs to. **Immutable.** |
| `type` | enum | yes | One of `Datacenter`, `AvailabilityZone`, `Edge`, `Virtual`. |
| `address` | string | no | Free-form postal/street address. |
| `providerRef` | LocalObjectReference | no | The [Provider](provider.md) that runs this Site. Mutable (a site can change providers). |

## Status

- Condition `Ready`: `Accepted`; `RegionNotFound` while `regionRef` is dangling,
  or `ProviderNotFound` while `providerRef` is set but dangling (the controller
  requeues until the referent appears).

The controller propagates `topology.inventory.miloapis.com/{region,site,site-type}`
labels onto the Site once its Region resolves.

## Validation & deletion

- `regionRef` is immutable (CEL).
- Deletion is **rejected** while any [Node](node.md), [Cluster](cluster.md), or
  [NetworkDevice](networkdevice.md) references the Site (`vsite` webhook).

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Site
metadata:
  name: us-east-1a
spec:
  displayName: US East 1A
  regionRef:
    name: us-east
  type: Datacenter
  address: 21715 Filigree Ct, Ashburn, VA 20147
  providerRef:
    name: equinix
```
