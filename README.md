# Inventory Service

Kubernetes operator implementing Datum Cloud's Asset Inventory (`inventory.miloapis.com`) -- see the enhancement at [github.com/datum-cloud/infra/docs/enhancements/infrastructure-platform/asset-inventory](https://github.com/datum-cloud/infra/tree/main/docs/enhancements/infrastructure-platform/asset-inventory).

Asset Inventory is a pure-data-model registry of infrastructure assets (regions, sites, clusters, nodes, network devices, and the links between them) and their geographic placement. It answers *"what do we have and where is it?"* and is written by Fleet Operations and read by Fleet Networking, compliance tools, capacity planners, and operators. It is explicitly **not** an operational system: no agents, no provisioning, no configuration delivery, no provider integrations.

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
