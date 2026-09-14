# Inventory Service

Kubernetes operator implementing Datum Cloud's Asset Inventory (`inventory.miloapis.com`) -- see the enhancement at [github.com/datum-cloud/infra/docs/enhancements/infrastructure-platform/asset-inventory](https://github.com/datum-cloud/infra/tree/main/docs/enhancements/infrastructure-platform/asset-inventory).

Asset Inventory is a pure-data-model registry of infrastructure assets (regions, sites, clusters, nodes, network devices, and the links between them) and their geographic placement. It answers *"what do we have and where is it?"* and is written by Fleet Operations and read by Fleet Networking, compliance tools, capacity planners, and operators. It is explicitly **not** an operational system: no agents, no provisioning, no configuration delivery, no provider integrations.

## Usage

Every inventory kind is **cluster-scoped** and served by the
`inventory.miloapis.com` API group on the organization control plane. There is
no per-project copy, so no project needs to be selected. Manage them with
[`datumctl`](https://github.com/datum-cloud/datumctl):

```bash
cat <<'EOF' | datumctl apply -f -
apiVersion: inventory.miloapis.com/v1alpha1
kind: Region
metadata:
  name: us-east
spec:
  displayName: US East
EOF

datumctl get regions
```

See [docs/usage.md](docs/usage.md) for an end-to-end walkthrough, and
[docs/api/inventory.md](docs/api/inventory.md) for the generated API reference.

### The `datumctl inventory` plugin

This repository also ships a [`datumctl`](https://github.com/datum-cloud/datumctl)
plugin that gives the inventory a purpose-built read view -- filters, a
region/site/node tree, and fleet-wide counts -- rather than raw `get` output:

```bash
datumctl plugin install milo-os/inventory

datumctl inventory get sites              # or: datumctl inventory sites
datumctl inventory sites --provider netactuate
datumctl inventory nodes --cluster us-central-1-lab
datumctl inventory tree
datumctl inventory summary
```

`get` accepts either spelling of a kind, singular or plural and
case-insensitively, so `get sites`, `get site` and `get Site` are equivalent.
Each kind's own filter flags work after it.

The plugin source lives in [`cmd/datumctl-inventory`](cmd/datumctl-inventory)
and is released from this repository's tags by
[`.goreleaser-plugin.yaml`](.goreleaser-plugin.yaml), separately from the
operator's container image and kustomize bundle.

## Development

```bash
task build       # Build the binary
task test        # Run tests
task lint        # Run linter
task generate    # Run code generation
task manifests   # Generate CRD, RBAC, and webhook manifests
task api-docs    # Generate the API reference (docs/api) from the CRDs
task dev:setup   # Bring up a kind cluster, build, load, and deploy
task e2e         # Run chainsaw e2e tests against the dev cluster
```
