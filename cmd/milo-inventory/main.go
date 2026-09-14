// SPDX-License-Identifier: AGPL-3.0-only

// Command milo-inventory is a datumctl plugin that provides a read view over
// the Datum Cloud physical inventory (providers, regions, sites, clusters,
// nodes). Invoked as `datumctl inventory ...` once installed.
package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.datum.net/datumctl/plugin"
)

func main() {
	// Serve --plugin-manifest via the datumctl SDK before cobra runs, so the
	// manifest is emitted even if flag parsing would otherwise fail.
	// ServeManifest prints the JSON and exits 0 when the flag is present;
	// otherwise it returns.
	plugin.ServeManifest(pluginManifest())

	root := &cobra.Command{
		Use:   "inventory",
		Short: "Browse the Datum Cloud physical inventory",
		Long: `Browse the Datum Cloud physical inventory: providers, regions, sites,
clusters, and nodes.

These records describe the real infrastructure Datum Cloud runs on — which
provider owns a site, which region a site sits in, and which nodes are assigned
to which cluster. Use the list subcommands to query one kind at a time,
'inventory tree' to see the region/site/node hierarchy, and 'inventory summary'
for fleet-wide counts.

Inventory lives on the Datum Cloud platform root, so these commands talk to the
platform API directly; they do not take an organization or project scope.`,
		Example: `  # List every region
  datumctl inventory regions

  # Sites in one region, or by provider
  datumctl inventory sites --region us-central-2
  datumctl inventory sites --provider netactuate

  # Nodes at a site or in a cluster
  datumctl inventory nodes --site us-central-2a
  datumctl inventory nodes --cluster my-edge-cluster

  # The same listings via 'get', singular or plural
  datumctl inventory get sites
  datumctl inventory get Site

  # Region -> site -> node hierarchy, and fleet-wide counts
  datumctl inventory tree
  datumctl inventory summary`,
		SilenceUsage: true,
	}
	root.PersistentFlags().StringP("output", "o", "table", "Output format. One of: table, json, yaml.")

	root.AddCommand(
		newListCmd(providersView),
		newListCmd(regionsView),
		newListCmd(sitesView),
		newListCmd(clustersView),
		newListCmd(nodesView),
		newListCmd(racksView),
		newListCmd(networkDevicesView),
		newListCmd(portsView),
		newListCmd(cablesView),
		newListCmd(circuitsView),
		newListCmd(linksView),
		newListCmd(virtualMachinesView),
		newTreeCmd(),
		newSummaryCmd(),
		newApplyCmd(),
	)
	root.AddCommand(newGetCmd(root))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
