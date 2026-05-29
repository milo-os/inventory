# Provider

An infrastructure provider relationship — hosting, colocation, transit,
internet exchange, dark fiber, or cloud — with contract metadata and the
service identifiers the provider assigns. A Provider has no parent references.

- **Group/Version**: `inventory.miloapis.com/v1alpha1`
- **Kind**: `Provider`
- **Scope**: Cluster

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `displayName` | string | yes | Human-readable name (min length 1). |
| `type` | enum | yes | One of `Hosting`, `Colocation`, `Transit`, `InternetExchange`, `DarkFiber`, `Cloud`. |
| `contract` | object | no | Contract metadata — see below. |
| `serviceIdentifiers` | []object | no | Named identifiers the provider assigns — see below. |

### `contract`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `contractID` | string | no | Provider's contract or agreement identifier. |
| `accountID` | string | no | Operator's account identifier with the provider. |
| `portalURL` | string | no | Link to the provider's management portal; must match `^https?://.+$`. |
| `notes` | string | no | Free-form contract context. |

### `serviceIdentifiers[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Identifier label, e.g. `ASN`, `LOA-CFA`, `IBX` (min length 1). |
| `identifier` | string | yes | The value the provider assigned (min length 1). |

## Status

- Condition `Ready`: `Accepted` once reconciled (`Pending` before).

## Validation & deletion

- Deletion is **rejected** while any [Site](site.md) references the Provider via
  `spec.providerRef` (`vprovider` webhook).

## Example

```yaml
apiVersion: inventory.miloapis.com/v1alpha1
kind: Provider
metadata:
  name: equinix
spec:
  displayName: Equinix
  type: Colocation
  contract:
    contractID: EQ-12345
    accountID: "987654"
    portalURL: https://portal.equinix.com
    notes: Primary colo provider for US East
  serviceIdentifiers:
    - name: IBX
      identifier: DC11
```
