// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
)

// kindInfo describes one inventory kind as the plugin sees it.
type kindInfo struct {
	Kind      string   `json:"kind"`
	Plural    string   `json:"plural,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	Supported bool     `json:"supported"`
	Apply     bool     `json:"apply"`
}

type kindsList struct {
	Kinds []kindInfo `json:"kinds"`
}

// pluralToKind maps a plural back to its Go kind name. `get` keys off the
// plural and `apply` keys off the kind; this is the one place the two meet.
var pluralToKind = map[string]string{
	"providers":       "Provider",
	"regions":         "Region",
	"sites":           "Site",
	"racks":           "Rack",
	"clusters":        "Cluster",
	"nodes":           "Node",
	"networkdevices":  "NetworkDevice",
	"virtualmachines": "VirtualMachine",
	"ports":           "Port",
	"cables":          "Cable",
	"links":           "Link",
	"circuits":        "Circuit",
}

// aliasesFor returns every spelling that resolves to a plural, excluding the
// plural itself, so `kinds` reports exactly what `get` accepts.
func aliasesFor(plural string) []string {
	var out []string
	for alias, canonical := range kindAliases {
		if canonical == plural && alias != plural {
			out = append(out, alias)
		}
	}
	sort.Strings(out)
	return out
}

// servedKinds returns every kind registered in the v1alpha1 scheme. Deriving
// this from the scheme rather than a literal list means a kind added to the API
// but not wired into the plugin is reported as unsupported instead of vanishing.
func servedKinds() []string {
	scheme := runtime.NewScheme()
	if err := inventoryv1alpha1.AddToScheme(scheme); err != nil {
		return nil
	}
	var out []string
	for kind := range scheme.KnownTypes(inventoryv1alpha1.GroupVersion) {
		// The scheme also carries List types and the metav1 options types
		// that AddToScheme registers; neither is an inventory kind.
		if strings.HasSuffix(kind, "List") || strings.HasSuffix(kind, "Options") ||
			kind == "WatchEvent" || strings.HasPrefix(kind, "APIGroup") ||
			strings.HasPrefix(kind, "APIResource") || strings.HasPrefix(kind, "APIVersions") ||
			kind == "Status" || kind == "CreateOptions" {
			continue
		}
		out = append(out, kind)
	}
	sort.Strings(out)
	return out
}

// pluralFor inverts pluralToKind.
func pluralFor(kind string) string {
	for p, k := range pluralToKind {
		if k == kind {
			return p
		}
	}
	return ""
}

func collectKinds() kindsList {
	out := make([]kindInfo, 0, len(pluralToKind))
	for _, kind := range servedKinds() {
		plural := pluralFor(kind)
		_, applyable := kindOrder(kind)
		info := kindInfo{
			Kind:      kind,
			Plural:    plural,
			Supported: plural != "",
			Apply:     applyable,
		}
		if plural != "" {
			info.Aliases = aliasesFor(plural)
		}
		out = append(out, info)
	}
	return kindsList{Kinds: out}
}

func newKindsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kinds",
		Short: "List the inventory kinds this plugin supports",
		Long: `List the inventory kinds this plugin supports, the spellings 'get' accepts
for each, and whether 'apply' can create or update them.

Kinds are read from the inventory API scheme, so a kind the API serves but the
plugin does not yet handle is listed with SUPPORTED=no rather than omitted.`,
		Example: `  datumctl inventory kinds
  datumctl inventory kinds -o json`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			list := collectKinds()
			rows := make([][]string, 0, len(list.Kinds))
			for _, k := range list.Kinds {
				rows = append(rows, []string{
					k.Kind,
					orNone(k.Plural),
					orNone(strings.Join(k.Aliases, ", ")),
					yesNo(k.Supported),
					yesNo(k.Apply),
				})
			}
			return emit(cmd, list, []string{"KIND", "PLURAL", "ALIASES", "SUPPORTED", "APPLY"}, rows)
		},
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
