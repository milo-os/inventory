# Inventory IAM

Authz on the Milo core control plane is **Milo IAM** (`ProtectedResource`-driven),
not Kubernetes RBAC. A CRD being `Established` and served grants **nobody**
access — every `inventory.miloapis.com` kind also needs a `ProtectedResource`,
a `Role` over the group, and a `PolicyBinding` for the principal that writes it.
This package ships the first two; the binding is environment-specific (see
below).

## What's here

- `protected-resources/` — one `ProtectedResource` per inventory kind
  (`inventory.miloapis.com-<singular>`). Cluster-scoped, root resources (no
  `parentResources`). Registering them is what makes the permission strings
  (`inventory.miloapis.com/<plural>.<verb>`) addressable by IAM.
- `roles/` — group-wide `Role`s, namespaced to `milo-system`:
  - `inventory.miloapis.com-viewer` — `get`/`list`/`watch` on all kinds.
  - `inventory.miloapis.com-editor` — viewer + `create`/`update`/`patch`/`delete`.
  - `inventory.miloapis.com-admin` — inherits editor (full access).
  - `inventory.miloapis.com-operator` — read + `update`/`patch` for the
    controller to reconcile, set conditions, and propagate topology labels.
- `policybindings.example.yaml` — a template binding a principal across the
  whole group. **Not** in `kustomization.yaml` — subjects are per-environment.

## Deployment

Mirrors `config/base/crd`: this targets **Milo**, not the cluster the manager
pod runs in (the `iam.miloapis.com` CRDs only exist on Milo). It is applied by
the same Flux Kustomization that applies `base/crd` + the webhook component —
add `config/base/iam` to that Kustomization's paths. It is intentionally **not**
wired into `config/base`, which forces the `inventory-system` namespace and is
also consumed by the dev overlay against a kind cluster with no IAM CRDs.

## Who can write inventory

`Role`s grant nothing on their own — they must be bound to a subject via a
`PolicyBinding`. Inventory kinds are root, cluster-scoped resources with no
parent, so a binding scopes through `resourceSelector.resourceKind` (which
applies the role to **all** resources of one kind — platform-wide). Because
`resourceKind` names a single kind, writing the full group means **one binding
per kind**.

To let a loader principal populate inventory, copy `policybindings.example.yaml`,
set the subject (a `User`, or a loader `ServiceAccount` with its `namespace`),
pick `-editor` or `-admin`, and apply it on Milo in `milo-system`.

Self-serve caveat: a caller holding only `policybindings.create` cannot bind
itself unless these `Role`s already exist on the cluster — without
`roles.{get,create}` it can neither find nor author one. Shipping the `Role`s
here closes that gap.

## Verification

With a binding for the loader principal:

```
bin/load-inventory --dry-run all   # expect 0 Forbidden
bin/load-inventory all             # Provider -> Region -> Site -> ...
```
