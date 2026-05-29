# Inventory Service

Kubernetes operator implementing Datum Cloud's Asset Inventory (`inventory.miloapis.com`) -- see the enhancement at [github.com/datum-cloud/infra/docs/enhancements/infrastructure-platform/asset-inventory](https://github.com/datum-cloud/infra/tree/main/docs/enhancements/infrastructure-platform/asset-inventory).

Asset Inventory is a pure-data-model registry of infrastructure assets (regions, sites, clusters, nodes, network devices, and the links between them) and their geographic placement. It answers *"what do we have and where is it?"* and is written by Fleet Operations and read by Fleet Networking, compliance tools, capacity planners, and operators. It is explicitly **not** an operational system: no agents, no provisioning, no configuration delivery, no provider integrations.

## Usage

Inventory resources are served by the `inventory.miloapis.com` API group on the
Milo control plane. Once the operator is deployed, manage them with
[`datumctl`](https://github.com/datum-cloud/datumctl).

### Point datumctl at the control plane

```bash
datumctl auth login
datumctl auth update-kubeconfig --organization your-org
```

### Demonstrate basic functionality

Record a region and a site within it, then read them back:

```bash
# Create a Region and a Site that references it.
cat <<'EOF' | datumctl apply -f -
apiVersion: inventory.miloapis.com/v1alpha1
kind: Region
metadata:
  name: us-east
spec:
  displayName: US East
  coordinates:
    latitude: 37.5
    longitude: -77.5
---
apiVersion: inventory.miloapis.com/v1alpha1
kind: Site
metadata:
  name: us-east-1a
spec:
  displayName: US East 1A
  regionRef:
    name: us-east
  type: Datacenter
EOF

# List them — the Ready column flips to True once the operator accepts each.
datumctl get regions,sites
```

```text
NAME      DISPLAY   AGE   READY   REASON
us-east   US East   5s    True    Accepted

NAME         REGION    TYPE         AGE   READY   REASON
us-east-1a   us-east   Datacenter   5s    True    Accepted
```

Referential integrity is enforced: deleting a Region while a Site still
references it is rejected.

```bash
datumctl delete region us-east
# Error: admission webhook "vregion.inventory.miloapis.com" denied the request:
# cannot delete Region us-east: 1 Site(s) still reference it: [us-east-1a]
```

### Explore and audit

```bash
datumctl api-resources --api-group=inventory.miloapis.com   # list inventory kinds
datumctl explain site.spec                                  # field documentation
datumctl describe site us-east-1a                           # full object + status
datumctl activity history sites us-east-1a --diff           # who changed what, when
```

## Development

```bash
task build       # Build the binary
task test        # Run tests
task lint        # Run linter
task generate    # Run code generation
task manifests   # Generate CRD, RBAC, and webhook manifests
task dev:setup   # Bring up a kind cluster, build, load, and deploy
task e2e         # Run chainsaw e2e tests against the dev cluster
```
