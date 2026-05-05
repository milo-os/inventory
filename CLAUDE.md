# Inventory Operator -- Claude Code guide

This repo is a Kubernetes operator implementing Datum Cloud's **Asset Inventory**
under the API group `inventory.miloapis.com/v1alpha1`. It is a pure-data-model
registry -- no agents, no provisioning, no provider integrations. Controllers
only validate cross-resource references and propagate topology labels.

## Stack

- **Kubebuilder v4** scaffolding (see `PROJECT`)
- **Go 1.24.x**
- **controller-runtime 0.22.x**
- **Build tool**: [go-task](https://taskfile.dev) via `Taskfile.yaml` -- **not
  make**. There is no Makefile.
- **E2E**: [chainsaw](https://github.com/kyverno/chainsaw) tests under
  `test/e2e/`, run against a [datum-cloud/test-infra](https://github.com/datum-cloud/test-infra)
  kind cluster

## API conventions

- **Group**: `inventory.miloapis.com` (flat; no sub-groups)
- **Version**: `v1alpha1`
- **Scope**: every kind is **Cluster-scoped** (`namespaced: false` in `PROJECT`)
- **Kinds** (v0.1): `Region`, `Site`, `Cluster`, `Node`, `NetworkDevice`, `Link`
- **Cross-resource refs**: inline struct; use `LocalObjectReference { Name string }`
  for single-kind refs and `AssetReference { Kind, Name string }` for
  polymorphic refs (CEL-restricts the `Kind`)

## Conditions pattern

- `Status.Conditions []metav1.Condition` on every kind
- `Ready` is the primary condition type; additional types (e.g.
  `EndpointsResolved` on `Link`) may exist alongside it
- Condition type/reason string **constants live next to the type** in
  `api/v1alpha1/<kind>_types.go`
- Default status is **seeded via `+kubebuilder:default=`** on the `Status`
  field so clients see a populated `Ready=Unknown` condition immediately on
  create
- Set conditions with `k8s.io/apimachinery/pkg/api/meta.SetStatusCondition`;
  small helpers in `internal/controller/conditions.go` keep reconcilers
  uniform

## Deletion protection

Parent kinds (Region, Site, Cluster) protect against deletion while dependents
exist **via validating webhooks that reject the DELETE operation**. We do
**not** use finalizers for this -- the enhancement frames the operator as "a
record of what is declared", so an inconsistent DELETE should be rejected up
front rather than leaving a zombie terminating object. See
`internal/webhook/` and the webhook table in the plan.

## License header

Every Go file (including generated files -- see `hack/boilerplate.go.txt`)
starts with:

```go
// SPDX-License-Identifier: AGPL-3.0-only
```

## References

- Design plan (rationale, CRD shapes, controller/webhook contracts, e2e
  suites): `/Users/scotwells/.claude/plans/happy-hugging-sky.md`
- Source enhancement: [github.com/datum-cloud/infra/docs/enhancements/infrastructure-platform/asset-inventory](https://github.com/datum-cloud/infra/tree/main/docs/enhancements/infrastructure-platform/asset-inventory)
- Layout/tooling reference: [datum-cloud/billing](https://github.com/datum-cloud/billing/tree/feat/03-implementation) (branch `feat/03-implementation`)
