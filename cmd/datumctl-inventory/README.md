# datumctl-inventory

A [datumctl](https://github.com/datum-cloud/datumctl) plugin that provides a
read view over the Datum Cloud physical inventory served by this operator
(`inventory.miloapis.com/v1alpha1`). Once installed it is invoked as
`datumctl inventory ...`.

## Commands

| Command | Description |
|---|---|
| `datumctl inventory providers` | List providers |
| `datumctl inventory regions` | List regions |
| `datumctl inventory sites [--region R] [--provider P]` | List sites |
| `datumctl inventory clusters [--region R] [--site S]` | List clusters |
| `datumctl inventory nodes [--region R] [--site S] [--cluster C]` | List nodes |
| `datumctl inventory tree [--region R]` | region → site → node hierarchy |
| `datumctl inventory summary` | Fleet-wide counts |

All subcommands accept `-o table|json|yaml` (default `table`).

`--region`, `--site`, and `--cluster` filter server-side using the
`topology.inventory.miloapis.com/*` labels the operator propagates onto
inventory objects. `--provider` filters on the site's `providerRef`.

Inventory objects are cluster-scoped on the Datum Cloud platform root, so the
plugin talks to the platform API directly and takes no organization or project
scope.

## How it works

datumctl injects context via environment variables and execs the plugin. The
plugin reads `DATUM_API_HOST`, fetches a short-lived token through the
credentials helper (`plugin.Token()`), and builds a controller-runtime client
against the platform root using this repo's own typed API
(`go.miloapis.com/inventory/api/v1alpha1`). See the
[datumctl plugin docs](https://github.com/datum-cloud/datumctl/blob/main/docs/developer/plugins.md).

## Build

```sh
go build -o datumctl-inventory ./cmd/datumctl-inventory
```

The version reported in `--plugin-manifest` is set via
`-ldflags "-X main.version=<version>"` at release time.

## Releases

The plugin **shares the operator's version and tag**. When an `inventory`
release is published, `.github/workflows/release-datumctl-inventory.yaml` runs
goreleaser and *appends* `datumctl-inventory_{OS}_{Arch}` archives plus
`checksums.txt` to that same GitHub release (alongside the operator image
published by `publish.yaml`). So plugin `vX.Y.Z` == operator `vX.Y.Z`.

## Local use

Build the binary onto your `PATH` named `datumctl-inventory`, then:

```sh
datumctl plugin trust inventory   # unmanaged plugins must be trusted once
datumctl inventory summary
```
