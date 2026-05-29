# Resource reference

All inventory kinds are **cluster-scoped** and served by the
`inventory.miloapis.com/v1alpha1` API group. Each carries a `Ready` status
condition the operator sets once the object is accepted and its references
resolve.

The hierarchy:

```
Region
  └── Site               (regionRef)
        ├── Cluster       (controlPlaneSiteRef)
        ├── Node          (siteRef; optional assignment → Cluster)
        └── NetworkDevice (siteRef + clusterRef)
Link                      (connects two of: Site, Cluster, NetworkDevice)
```

| Kind | Summary |
|------|---------|
| [Region](region.md) | Top-level geographic grouping. |
| [Site](site.md) | A facility / AZ / edge location within a Region. |
| [Cluster](cluster.md) | A Kubernetes cluster; its control plane lives at one Site. |
| [Node](node.md) | A physical/virtual machine in a Site, optionally assigned to a Cluster. |
| [NetworkDevice](networkdevice.md) | A switch/router/firewall in a Site, part of a Cluster. |
| [Link](link.md) | Connectivity between two assets. |

## Conventions

- **References** use a `LocalObjectReference` (`{ name: <other-object> }`) since
  all kinds are cluster-scoped. `Link` endpoints use a typed `AssetReference`
  (`{ kind, name }`).
- **Dangling references** are surfaced on `.status` (e.g. `RegionNotFound`)
  rather than rejected at admission, except where noted — the controller
  re-reconciles when the referent appears.
- **Deletion guards** are enforced by validating webhooks: a parent cannot be
  deleted while children reference it.
- **Topology labels** (`topology.inventory.miloapis.com/{region,site,site-type,cluster}`)
  are propagated onto objects by controllers so any client can select by
  topology.
