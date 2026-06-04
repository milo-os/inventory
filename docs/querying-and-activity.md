# Querying and Activity

This guide is for staff and operators who run inventory on the Milo control
plane. It covers two everyday tasks:

1. **Querying inventory** — list and slice kinds with `kubectl`/`datumctl`,
   printer columns, and topology label selectors.
2. **Reading inventory activity** — turn controller events and audit logs into
   human-readable timelines with `datumctl activity query`.

Inventory kinds (`Region`, `Provider`, `Site`, `Rack`, `Cluster`, `Node`,
`NetworkDevice`, `Link`, `Port`, `Cable`, `Circuit`, `VirtualMachine`) are
**cluster-scoped** infrastructure objects. See the [API
reference](api/inventory.md) for the full field reference.

## Querying inventory

### List a kind

`kubectl get <kind>` and `datumctl get <kind>` print the columns the operator
exposes for each kind, so you rarely need `-o yaml` to get oriented:

```bash
task kubectl -- get sites
# or, against a control plane you've targeted with datumctl:
datumctl get sites
```

```text
NAME         REGION    PROVIDER   TYPE         AGE   READY   REASON
us-east-1a   us-east   equinix    Datacenter   5m    True    Accepted
```

Other kinds carry columns tuned to what you usually want at a glance. For
example:

```bash
task kubectl -- get circuits
```

```text
NAME        PROVIDER   TYPE       CIRCUITID    MBPS    READY   REASON
iad-tx-01   lumen      Transit    LOA-99213    10000   True    Accepted
```

```bash
task kubectl -- get nodes
```

```text
NAME       SITE         CLUSTER      ROLE     ARCH    CPU   PHASE     READY   REASON
node-001   us-east-1a   prod-east    worker   arm64   64    Running   True    Accepted
```

The `READY` and `REASON` columns reflect the `Ready` condition. `REASON` shows
why a kind is not ready (for example, `ProviderNotFound`).

### Slice by topology with label selectors

The operator stamps five well-known topology labels onto placed kinds (Nodes,
NetworkDevices, VMs, and other site-bound objects). Each label answers one
question:

| Label | Answers |
|-------|---------|
| `topology.inventory.miloapis.com/region` | Which region is this in? |
| `topology.inventory.miloapis.com/site` | Which site is this in? |
| `topology.inventory.miloapis.com/site-type` | What kind of site (for example, `Datacenter`)? |
| `topology.inventory.miloapis.com/cluster` | Which cluster does this belong to? |
| `topology.inventory.miloapis.com/rack` | Which rack does this sit in? |

Selectors are the ergonomic way to ask "everything in X":

```bash
# Everything in one site.
task kubectl -- get nodes --selector topology.inventory.miloapis.com/site=us-east-1a

# Everything in a region.
task kubectl -- get nodes,networkdevices --selector topology.inventory.miloapis.com/region=us-east

# Everything in one cluster.
task kubectl -- get nodes --selector topology.inventory.miloapis.com/cluster=prod-east

# Everything in one rack.
task kubectl -- get nodes --selector topology.inventory.miloapis.com/rack=r14

# Combine labels to narrow further — datacenter nodes in one region.
task kubectl -- get nodes \
  --selector topology.inventory.miloapis.com/region=us-east,topology.inventory.miloapis.com/site-type=Datacenter
```

### Find objects by reference

The operator maintains field indexers on reference fields (for example,
`spec.providerRef.name` on Circuits, Sites, and VMs, and the various
`siteRef`/`clusterRef`/`regionRef` fields). The controllers and validating
webhooks use these for fast parent-to-child lookups.

As a client, you do not query indexers directly. Use the equivalent field
selector, or the topology labels above when the relationship is positional:

```bash
# All Circuits delivered by a given Provider.
task kubectl -- get circuits --field-selector spec.providerRef.name=lumen
```

## Reading inventory activity

### How it works

The Activity capability (`activity.miloapis.com`) renders human-readable
timelines from two sources:

- **Audit logs** supply CRUD actions — who created, updated, or deleted an
  object, and when.
- **Controller events** supply async outcomes that audit logs cannot see —
  whether an object became ready, or why it could not (for example, a
  referenced Provider does not yet exist).

The inventory operator ships a declarative `ActivityPolicy` per kind (in
`config/milo/activity/policies/`) and emits `events.k8s.io/v1` Events on every
`Ready` / `NotReady` transition. The Activity system matches audit logs and
these events against the policies to produce `Activity` records. Note that the
operator emits events only for async outcomes; CRUD activity comes entirely
from audit logs, not from duplicate "created/updated/deleted" events.

### Query timelines with datumctl

Query activity records with a CEL filter:

```bash
datumctl activity query --filter "<CEL expression>"
```

The examples below are **illustrative**. The activity filter language operates
over activity records, and exact field names can vary by platform version —
confirm the available fields with:

```bash
datumctl activity query --help
```

Common filters:

```bash
# All inventory activity.
datumctl activity query --filter "resource.apiGroup == 'inventory.miloapis.com'"

# Activity for one kind.
datumctl activity query --filter "resource.kind == 'Circuit'"

# Activity for one object.
datumctl activity query \
  --filter "resource.kind == 'Circuit' && resource.name == 'iad-tx-01'"

# Activity by a specific actor.
datumctl activity query --filter "actor.email == 'operator@datum.net'"

# Activity in a time window (combine with the command's time-range flags).
datumctl activity query \
  --filter "resource.apiGroup == 'inventory.miloapis.com'" \
  --since 24h
```

You can also fetch the recent history for a single object directly:

```bash
datumctl activity history circuits iad-tx-01 --diff
```

### Worked example: a Circuit waiting on its Provider

This example mirrors the milestone exit criteria. Create a `Circuit` that
references a `Provider` that does not exist yet, then create the `Provider` and
watch the timeline complete.

Create the Circuit first:

```bash
cat <<'EOF' | task kubectl -- apply -f -
apiVersion: inventory.miloapis.com/v1alpha1
kind: Circuit
metadata:
  name: iad-tx-01
spec:
  providerRef:
    name: lumen
  type: Transit
  circuitID: LOA-99213
  bandwidthMbps: 10000
  endpoint:
    kind: Site
    name: us-east-1a
EOF
```

The Circuit is not ready because its Provider is missing:

```bash
task kubectl -- get circuit iad-tx-01
```

```text
NAME        PROVIDER   TYPE      CIRCUITID   MBPS    READY   REASON
iad-tx-01   lumen      Transit   LOA-99213   10000   False   ProviderNotFound
```

The timeline shows the create and the not-ready outcome:

```bash
datumctl activity query \
  --filter "resource.kind == 'Circuit' && resource.name == 'iad-tx-01'"
```

```text
10:00:00  operator@datum.net created Circuit iad-tx-01
10:00:01  Circuit iad-tx-01 is not ready: Provider lumen not found
```

Now create the Provider:

```bash
cat <<'EOF' | task kubectl -- apply -f -
apiVersion: inventory.miloapis.com/v1alpha1
kind: Provider
metadata:
  name: lumen
spec:
  displayName: Lumen
EOF
```

The operator re-reconciles the Circuit, the reference resolves, and the
timeline records the Circuit becoming ready:

```text
10:00:00  operator@datum.net created Circuit iad-tx-01
10:00:01  Circuit iad-tx-01 is not ready: Provider lumen not found
10:02:30  operator@datum.net created Provider lumen
10:02:31  Circuit iad-tx-01 is ready
```

## Scope and surfaces

Inventory kinds are cluster-scoped and infra-facing, so their activity is
**platform/operator-scoped, not per-project**. These timelines describe shared
infrastructure (regions, sites, racks, circuits, and the hardware in them), not
a single tenant's resources.

Inventory activity surfaces to staff through the staff portal and `datumctl`.
It does not appear on consumer or project-facing surfaces.
