// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
)

func quantity(q *resource.Quantity) string {
	if q == nil {
		return none
	}
	return q.String()
}

func mbps(v *int64) string {
	if v == nil {
		return none
	}
	return strconv.FormatInt(*v, 10)
}

// placementCell renders a device's rack position as "<rack> U<start>-U<end>
// (<face>)", the form an operator reads off the rack itself.
func placementCell(p *inventoryv1alpha1.Placement) string {
	if p == nil {
		return none
	}
	face := p.Face
	if face == "" {
		face = inventoryv1alpha1.RackFaceFront
	}
	end := p.StartUnit + p.UnitHeight - 1
	return fmt.Sprintf("%s U%d-U%d (%s)", p.RackRef.Name, p.StartUnit, end, face)
}

var racksView = resourceView{
	use:   "racks",
	short: "List inventory racks",
	bind: func(cmd *cobra.Command) runFunc {
		site := cmd.Flags().String("site", "", "Filter by site name")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.RackList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("racks", err)
			}
			if *site != "" {
				kept := list.Items[:0]
				for _, r := range list.Items {
					if r.Spec.SiteRef.Name == *site {
						kept = append(kept, r)
					}
				}
				list.Items = kept
			}
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, r := range list.Items {
				feeds := make([]string, 0, len(r.Spec.PowerFeeds))
				for _, f := range r.Spec.PowerFeeds {
					feeds = append(feeds, f.Name)
				}
				rows = append(rows, []string{
					r.Name,
					orNone(r.Spec.SiteRef.Name),
					orNone(r.Spec.Cage),
					orNone(r.Spec.Row),
					strconv.Itoa(int(r.Spec.HeightU)) + "U",
					orNone(strings.Join(feeds, ",")),
					ready(r.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "SITE", "CAGE", "ROW", "HEIGHT", "FEEDS", "READY"}, rows, nil
		}
	},
}

var networkDevicesView = resourceView{
	use:   "networkdevices",
	short: "List inventory network devices",
	bind: func(cmd *cobra.Command) runFunc {
		site := cmd.Flags().String("site", "", "Filter by site name")
		cluster := cmd.Flags().String("cluster", "", "Filter by cluster name")
		rack := cmd.Flags().String("rack", "", "Filter by rack name")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.NetworkDeviceList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("networkdevices", err)
			}
			kept := list.Items[:0]
			for _, d := range list.Items {
				if *site != "" && d.Spec.SiteRef.Name != *site {
					continue
				}
				if *cluster != "" && d.Spec.ClusterRef.Name != *cluster {
					continue
				}
				if *rack != "" && (d.Spec.Placement == nil || d.Spec.Placement.RackRef.Name != *rack) {
					continue
				}
				kept = append(kept, d)
			}
			list.Items = kept
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, d := range list.Items {
				rows = append(rows, []string{
					d.Name,
					orNone(d.Spec.SiteRef.Name),
					orNone(d.Spec.ClusterRef.Name),
					string(d.Spec.Role),
					placementCell(d.Spec.Placement),
					orNone(d.Spec.ManagementAddress),
					ready(d.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "SITE", "CLUSTER", "ROLE", "PLACEMENT", "MGMT", "READY"}, rows, nil
		}
	},
}

var portsView = resourceView{
	use:   "ports",
	short: "List inventory ports",
	bind: func(cmd *cobra.Command) runFunc {
		device := cmd.Flags().String("device", "", "Filter by device name (Node, NetworkDevice or Rack)")
		portType := cmd.Flags().String("type", "", "Filter by port type")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.PortList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("ports", err)
			}
			kept := list.Items[:0]
			for _, p := range list.Items {
				if *device != "" && p.Spec.DeviceRef.Name != *device {
					continue
				}
				if *portType != "" && !strings.EqualFold(string(p.Spec.Type), *portType) {
					continue
				}
				kept = append(kept, p)
			}
			list.Items = kept
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, p := range list.Items {
				rows = append(rows, []string{
					p.Name,
					orNone(p.Spec.DeviceRef.Kind),
					orNone(p.Spec.DeviceRef.Name),
					orNone(p.Spec.Name),
					string(p.Spec.Type),
					quantity(p.Spec.Speed),
					ready(p.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "DEVICE-KIND", "DEVICE", "PORT", "TYPE", "SPEED", "READY"}, rows, nil
		}
	},
}

var cablesView = resourceView{
	use:   "cables",
	short: "List inventory cables",
	bind: func(cmd *cobra.Command) runFunc {
		port := cmd.Flags().String("port", "", "Filter by an endpoint Port name")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.CableList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("cables", err)
			}
			if *port != "" {
				kept := list.Items[:0]
				for _, cb := range list.Items {
					for _, e := range cb.Spec.Endpoints {
						if e.Name == *port {
							kept = append(kept, cb)
							break
						}
					}
				}
				list.Items = kept
			}
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, cb := range list.Items {
				a, z := none, none
				if len(cb.Spec.Endpoints) > 0 {
					a = orNone(cb.Spec.Endpoints[0].Name)
				}
				if len(cb.Spec.Endpoints) > 1 {
					z = orNone(cb.Spec.Endpoints[1].Name)
				}
				rows = append(rows, []string{
					cb.Name, a, z, string(cb.Spec.Media),
					quantity(cb.Spec.LengthM), orNone(cb.Spec.Label),
					ready(cb.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "A-PORT", "Z-PORT", "MEDIA", "LENGTH-M", "LABEL", "READY"}, rows, nil
		}
	},
}

var circuitsView = resourceView{
	use:   "circuits",
	short: "List inventory circuits",
	bind: func(cmd *cobra.Command) runFunc {
		provider := cmd.Flags().String("provider", "", "Filter by provider name")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.CircuitList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("circuits", err)
			}
			if *provider != "" {
				kept := list.Items[:0]
				for _, cc := range list.Items {
					if cc.Spec.ProviderRef.Name == *provider {
						kept = append(kept, cc)
					}
				}
				list.Items = kept
			}
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, cc := range list.Items {
				rows = append(rows, []string{
					cc.Name,
					orNone(cc.Spec.ProviderRef.Name),
					string(cc.Spec.Type),
					orNone(cc.Spec.CircuitID),
					fmt.Sprintf("%s/%s", cc.Spec.AEnd.Kind, cc.Spec.AEnd.Name),
					fmt.Sprintf("%s/%s", cc.Spec.ZEnd.Kind, cc.Spec.ZEnd.Name),
					mbps(cc.Spec.BandwidthMbps),
					ready(cc.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "PROVIDER", "TYPE", "CIRCUIT-ID", "A-END", "Z-END", "MBPS", "READY"}, rows, nil
		}
	},
}

var linksView = resourceView{
	use:   "links",
	short: "List inventory links",
	bind: func(cmd *cobra.Command) runFunc {
		linkType := cmd.Flags().String("type", "", "Filter by link type")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.LinkList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("links", err)
			}
			if *linkType != "" {
				kept := list.Items[:0]
				for _, l := range list.Items {
					if strings.EqualFold(string(l.Spec.Type), *linkType) {
						kept = append(kept, l)
					}
				}
				list.Items = kept
			}
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, l := range list.Items {
				ends := make([]string, 0, len(l.Spec.Endpoints))
				for _, e := range l.Spec.Endpoints {
					ends = append(ends, fmt.Sprintf("%s/%s", e.Kind, e.Name))
				}
				cables := make([]string, 0, len(l.Spec.CableRefs))
				for _, cr := range l.Spec.CableRefs {
					cables = append(cables, cr.Name)
				}
				rows = append(rows, []string{
					l.Name,
					string(l.Spec.Type),
					orNone(strings.Join(ends, " <-> ")),
					mbps(l.Spec.CapacityMbps),
					quantity(l.Spec.LatencyMs),
					orNone(strings.Join(cables, ",")),
					ready(l.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "TYPE", "ENDPOINTS", "MBPS", "LATENCY-MS", "CABLES", "READY"}, rows, nil
		}
	},
}

var virtualMachinesView = resourceView{
	use:   "virtualmachines",
	short: "List inventory virtual machines",
	bind: func(cmd *cobra.Command) runFunc {
		host := cmd.Flags().String("host", "", "Filter by host Node name")
		provider := cmd.Flags().String("provider", "", "Filter by provider name")
		return func(ctx context.Context, c client.Client) (runtime.Object, []string, [][]string, error) {
			var list inventoryv1alpha1.VirtualMachineList
			if err := c.List(ctx, &list); err != nil {
				return nil, nil, nil, listErr("virtualmachines", err)
			}
			kept := list.Items[:0]
			for _, vm := range list.Items {
				if *host != "" && vm.Spec.HostRef.Name != *host {
					continue
				}
				if *provider != "" && (vm.Spec.ProviderRef == nil || vm.Spec.ProviderRef.Name != *provider) {
					continue
				}
				kept = append(kept, vm)
			}
			list.Items = kept
			sort.Slice(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
			rows := make([][]string, 0, len(list.Items))
			for _, vm := range list.Items {
				prov := none
				if vm.Spec.ProviderRef != nil {
					prov = orNone(vm.Spec.ProviderRef.Name)
				}
				rows = append(rows, []string{
					vm.Name,
					orNone(vm.Spec.HostRef.Name),
					prov,
					strconv.Itoa(int(vm.Spec.Allocation.VCPUs)),
					strconv.FormatInt(vm.Spec.Allocation.MemoryBytes, 10),
					ready(vm.Status.Conditions),
				})
			}
			return &list, []string{"NAME", "HOST", "PROVIDER", "VCPUS", "MEMORY-BYTES", "READY"}, rows, nil
		}
	},
}
