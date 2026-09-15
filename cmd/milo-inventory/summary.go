// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
)

func newSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show fleet-wide inventory counts",
		Long: `Print fleet-wide counts: totals per kind, sites and nodes per region, and
sites per provider.`,
		Example:      "  datumctl inventory summary",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			var counts []kindCount

			var providers inventoryv1alpha1.ProviderList
			var regions inventoryv1alpha1.RegionList
			var sites inventoryv1alpha1.SiteList
			var clusters inventoryv1alpha1.ClusterList
			var nodes inventoryv1alpha1.NodeList
			var racks inventoryv1alpha1.RackList
			var devices inventoryv1alpha1.NetworkDeviceList
			var ports inventoryv1alpha1.PortList
			var cables inventoryv1alpha1.CableList
			var links inventoryv1alpha1.LinkList
			var circuits inventoryv1alpha1.CircuitList
			var vms inventoryv1alpha1.VirtualMachineList

			// Ordered so the totals table reads as the containment hierarchy
			// does: geography, then physical plant, then logical overlay.
			for _, l := range []struct {
				name string
				list client.ObjectList
			}{
				{"providers", &providers},
				{"regions", &regions},
				{"sites", &sites},
				{"racks", &racks},
				{"clusters", &clusters},
				{"nodes", &nodes},
				{"networkdevices", &devices},
				{"virtualmachines", &vms},
				{"ports", &ports},
				{"cables", &cables},
				{"links", &links},
				{"circuits", &circuits},
			} {
				if err := c.List(ctx, l.list); err != nil {
					return listErr(l.name, err)
				}
				counts = append(counts, kindCount{l.name, meta.LenList(l.list)})
			}

			printSummary(cmd.OutOrStdout(), counts, sites, nodes)
			return nil
		},
	}
}

// kindCount is one row of the totals table.
type kindCount struct {
	kind  string
	count int
}

func printSummary(out io.Writer, counts []kindCount, sites inventoryv1alpha1.SiteList, nodes inventoryv1alpha1.NodeList) {
	fmt.Fprintln(out, "Totals")
	totalRows := make([][]string, 0, len(counts))
	for _, kc := range counts {
		totalRows = append(totalRows, []string{kc.kind, strconv.Itoa(kc.count)})
	}
	_ = printTable(out, []string{"KIND", "COUNT"}, totalRows)

	sitesPerRegion := map[string]int{}
	for _, s := range sites.Items {
		sitesPerRegion[s.Spec.RegionRef.Name]++
	}
	nodesPerRegion := map[string]int{}
	for _, n := range nodes.Items {
		r := n.Labels[inventoryv1alpha1.TopologyRegionLabel]
		if r == "" {
			r = none
		}
		nodesPerRegion[r]++
	}
	fmt.Fprintln(out, "\nPer region")
	regionRows := make([][]string, 0)
	for _, r := range sortedUnion(sitesPerRegion, nodesPerRegion) {
		regionRows = append(regionRows, []string{r, strconv.Itoa(sitesPerRegion[r]), strconv.Itoa(nodesPerRegion[r])})
	}
	_ = printTable(out, []string{"REGION", "SITES", "NODES"}, regionRows)

	sitesPerProvider := map[string]int{}
	for _, s := range sites.Items {
		p := none
		if s.Spec.ProviderRef != nil {
			p = s.Spec.ProviderRef.Name
		}
		sitesPerProvider[p]++
	}
	fmt.Fprintln(out, "\nSites per provider")
	providerRows := make([][]string, 0)
	for _, p := range sortedKeys(sitesPerProvider) {
		providerRows = append(providerRows, []string{p, strconv.Itoa(sitesPerProvider[p])})
	}
	_ = printTable(out, []string{"PROVIDER", "SITES"}, providerRows)
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedUnion(a, b map[string]int) []string {
	seen := map[string]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
