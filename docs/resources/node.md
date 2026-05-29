# Node

A physical or virtual machine that physically lives in a [Site](site.md) and
may optionally be assigned to a [Cluster](cluster.md).

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Node`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `siteRef` | LocalObjectReference | yes | [Site](site.md) where the Node lives. **Immutable.** |
| `hardware` | object | yes | Physical capabilities — see below. |
| `addresses` | []object | no | Reachable addresses: `type` (`Internal`/`External`/`Hostname`) + `address`. |
| `assignment` | object | no | Cluster membership — see below. Unset ⇒ unassigned. |

### `hardware`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cpuCores` | int32 | yes | Logical core count (≥1). |
| `cpuArchitecture` | enum | yes | `amd64` or `arm64`. |
| `memoryBytes` | int64 | yes | Total RAM in bytes (≥1). |
| `disks` | []object | no | Each: `name`, `sizeBytes` (≥1), `type` (`SSD`/`HDD`/`NVMe`). |

### `assignment`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clusterRef` | LocalObjectReference | yes | [Cluster](cluster.md) the Node joins. |
| `role` | enum | yes | `ControlPlane` or `Worker`. |

## Status

- Condition `Ready`: `Accepted`; `SiteNotFound` / `ClusterNotFound` while a
  reference is dangling.
- `phase`: coarse lifecycle — `Unassigned`, `Assigned`, or `Unavailable`.

## Validation & deletion

- `siteRef` is immutable (CEL).
- No deletion guard — Nodes are leaf assets.

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Node
metadata:
  name: use1-node-01
spec:
  siteRef:
    name: us-east-1a
  hardware:
    cpuCores: 64
    cpuArchitecture: amd64
    memoryBytes: 274877906944
    disks:
      - name: nvme0n1
        sizeBytes: 1920383410176
        type: NVMe
  addresses:
    - type: Internal
      address: 10.1.0.21
  assignment:
    clusterRef:
      name: use1-compute-1
    role: Worker
```
