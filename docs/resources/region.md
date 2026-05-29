# Region

The top-level geographic grouping in the inventory (e.g. "US East"). A Region
has no parent references.

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Region`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `displayName` | string | yes | Human-readable name (min length 1). |
| `coordinates` | object | no | Representative `latitude` (-90..90) and `longitude` (-180..180), decimal degrees. Descriptive only; not used for routing. |

## Status

- Condition `Ready`: `Accepted` once reconciled (`Pending` before).

## Validation & deletion

- Deletion is **rejected** while any [Site](site.md) references the Region via
  `spec.regionRef` (`vregion` webhook).

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Region
metadata:
  name: us-east
spec:
  displayName: US East
  coordinates:
    latitude: 37.5
    longitude: -77.5
```
