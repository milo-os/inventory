# Inventory Operator -- Claude Code guide

This repo is a Kubernetes operator implementing the **Asset Inventory** API
under `inventory.miloapis.com/v1alpha1`. It is a pure-data-model registry --
no agents, no provisioning, no provider integrations. Controllers only
validate cross-resource references and propagate topology labels.

## Stack

- **Kubebuilder v4** scaffolding (see `PROJECT`)
- **Go 1.24.x**
- **controller-runtime 0.22.x**
- **Build tool**: [go-task](https://taskfile.dev) via `Taskfile.yaml` -- **not
  make**. There is no Makefile.
- **E2E**: [chainsaw](https://github.com/kyverno/chainsaw) tests under
  `test/e2e/`, run against a kind cluster

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

## Adding a new kind

The kinds are near-identical in shape, so copy an existing one (`Region` is the
simplest; `Site` shows an optional/validated ref). A complete kind touches:

1. `api/v1alpha1/<kind>_types.go` — Spec/Status, `Ready` constants, print
   columns, `SchemeBuilder.Register`.
2. Code generation (see below) — deepcopy + CRD/RBAC/webhook manifests.
3. `internal/controller/<kind>_controller.go` — reconciler (mark `Ready`;
   validate refs by setting a `…NotFound` condition and requeuing; `Watches`
   the referent to wake on change).
4. `internal/webhook/v1alpha1/<kind>_webhook.go` — validator + `Setup…WebhookWithManager`.
5. `internal/controller/indexers.go` — a field index for any new ref
   (`Index…`). Webhooks and the enqueue map-funcs reuse these.
6. Wire it in **both** `cmd/inventory/main.go` **and**
   `internal/controller/suite_test.go` (reconciler + webhook). Forgetting the
   suite means envtest doesn't exercise it.
7. `config/base/crd/kustomization.yaml` — add the generated CRD base.
8. `config/base/webhook/kustomization.yaml` — the JSON6902 patch rewrites
   `clientConfig.service.{name,namespace}` per webhook **by index**.
   controller-gen emits webhooks sorted alphabetically by `v<kind>` name, so a
   new kind shifts indices — but every webhook gets the same `inventory-webhook`
   / `inventory-system` values, so just append one more index pair for the new
   webhook count.
9. `config/samples/` — add a sample + list it in the samples kustomization.

Reference integrity split: parent kinds enforce **delete** guards (Region,
Site, Cluster — reject DELETE while referenced, via the field indexers).
NetworkDevice and Link validate their refs **on create/update** (reject if the
referent is absent). Nodes are leaf assets (no guard).

## Code generation

`controller-gen` (v0.16.4) drives both deepcopy and manifests. The wrapper
tasks are `task generate` (deepcopy) and `task manifests` (CRD/RBAC/webhook).
If the task runner can't init in your environment, invoke the tool directly:

```bash
GOBIN=$PWD/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.4
./bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
./bin/controller-gen rbac:roleName=inventory \
  crd:generateEmbeddedObjectMeta=true,allowDangerousTypes=true webhook paths="./..." \
  output:crd:artifacts:config=config/base/crd/bases \
  output:rbac:artifacts:config=config/components/controller_rbac \
  output:webhook:artifacts:config=config/base/webhook
```

envtest binaries: `setup-envtest use 1.34.1` (see `ENVTEST_K8S_VERSION`), then
`KUBEBUILDER_ASSETS=... go test ./internal/controller/...`. `go build
./cmd/inventory` drops an `/inventory` binary at the repo root — it is
gitignored; do not commit it.

## Deployment & Milo integration

How this operator is deployed is non-obvious; mis-modeling it causes crash
loops (see issues #9, #13).

- The operator is deployed as **two OCI artifacts** built by CI from this repo:
  the `inventory` **container image** and the `inventory-kustomize` **config
  bundle** (`config/`, image tag baked in). A change under `config/` ships in
  the **bundle**, not the image.
- The manager pod and the **Milo control plane** it targets live on
  **different clusters**. The pod's API client is pointed at Milo via
  `KUBECONFIG=.../milo-kubeconfig.yaml`, set by the `milo-integration`
  component. So:
  - The CRDs and `ValidatingWebhookConfiguration` live **on Milo**, applied by a
    separate Flux Kustomization (`base/crd` + `webhook` component) — not on the
    cluster the pod runs in.
  - **Leader election** writes its `Lease` on Milo. `inventory-system` exists
    on the pod's local cluster, not Milo, so the base default would fail with
    `namespaces "inventory-system" not found`. The base sets
    `--leader-elect-namespace=$(LEADER_ELECT_NAMESPACE)` with a default env of
    `inventory-system`; `milo-integration` overrides the env to `milo-system`
    (which exists on Milo). Override the **env var via strategic merge**, never
    the arg by positional index.
  - The webhook serving cert (cert-manager CSI) must mount at the
    controller-runtime default `/tmp/k8s-webhook-server/serving-certs` unless
    you also set the server `CertDir`.
- `milo-integration` is a **kustomize Component** patching the base Deployment.
  Prefer strategic-merge env/volume patches (matched by name) over positional
  JSON6902 patches, which break when base ordering changes.

## Code Comments

**Default to zero comments.** Well-named identifiers and the surrounding code
should convey what the code does. Do not narrate changes, reference tasks/PRs,
or annotate "added for X" / "removed Y".

The only exception: if a passage is genuinely difficult for a human to
comprehend, leave a single short comment that reads exactly `here be dragons`.

## License header

Every Go file (including generated files -- see `hack/boilerplate.go.txt`)
starts with:

```go
// SPDX-License-Identifier: AGPL-3.0-only
```

## References

- Design plan (rationale, CRD shapes, controller/webhook contracts, e2e
  suites): `/Users/scotwells/.claude/plans/happy-hugging-sky.md`
