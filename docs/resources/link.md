# Link

Connectivity between two inventory assets — a physical cable, a logical
overlay, or transit over the public internet.

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Link`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `endpoints` | []AssetReference | yes | Exactly 2, and distinct. Each: `kind` (`Site`, `Cluster`, or `NetworkDevice`) + `name`. |
| `type` | enum | yes | One of `Physical`, `Logical`, `Internet`. |
| `capacityMbps` | int64 | no | Nominal capacity in Mbps (≥1). |
| `latencyMs` | Quantity | no | Nominal one-way latency in ms as a dimensionless Quantity (e.g. `5`, `250m` for 0.25 ms). |

## Status

- Condition `Ready`: `Accepted`; `EndpointNotFound` if an endpoint is missing.
- Condition `EndpointsResolved`: set once both endpoint objects are verified to
  exist (reported alongside `Ready` for observability).

## Validation & deletion

- Endpoints must be distinct and exactly two (CEL); endpoint `kind` must be one
  of the allowed kinds.
- On create/update the `vlink` webhook **rejects** the request if either
  endpoint object does not exist.
- No deletion guard — a Link is itself the relationship.

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Link
metadata:
  name: use1-to-usw1
spec:
  endpoints:
    - kind: Site
      name: us-east-1a
    - kind: Site
      name: us-west-1a
  type: Logical
  capacityMbps: 100000
  latencyMs: "60"
```
