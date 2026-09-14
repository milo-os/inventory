// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// kindAliases maps what a caller is likely to type for an inventory kind onto
// the canonical list subcommand name. Both the singular and plural spellings
// resolve, and matching is case-insensitive, so `get Site`, `get site` and
// `get sites` are all the same request.
var kindAliases = map[string]string{
	"provider": "providers",
	"region":   "regions",
	"site":     "sites",
	"cluster":  "clusters",
	"node":     "nodes",
}

func canonicalKind(arg string) (string, bool) {
	k := strings.ToLower(strings.TrimSpace(arg))
	if canonical, ok := kindAliases[k]; ok {
		return canonical, true
	}
	for _, canonical := range kindAliases {
		if k == canonical {
			return canonical, true
		}
	}
	return "", false
}

func knownKinds() []string {
	seen := make(map[string]struct{}, len(kindAliases))
	kinds := make([]string, 0, len(kindAliases))
	for _, canonical := range kindAliases {
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		kinds = append(kinds, canonical)
	}
	sort.Strings(kinds)
	return kinds
}

// newGetCmd adds `inventory get <KIND>` as a front door onto the per-kind list
// subcommands. The bare forms (`inventory sites`) keep working; this exists so
// that the kubectl-shaped spelling people reach for first also resolves.
func newGetCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "get KIND",
		Short: "List inventory objects of a kind",
		Long: `List inventory objects of a kind.

KIND accepts the singular or plural spelling, case-insensitively: 'get Site',
'get site' and 'get sites' are equivalent. Each kind's own filter flags are
accepted after the kind, so 'get sites --provider netactuate' behaves exactly
as 'sites --provider netactuate' does.`,
		Example: `  datumctl inventory get sites
  datumctl inventory get Site
  datumctl inventory get nodes --cluster us-central-1-lab`,
		Args:               cobra.MinimumNArgs(1),
		SilenceUsage:       true,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a kind is required; one of: %s", strings.Join(knownKinds(), ", "))
			}
			canonical, ok := canonicalKind(args[0])
			if !ok {
				return fmt.Errorf("unknown inventory kind %q; one of: %s", args[0], strings.Join(knownKinds(), ", "))
			}
			target, _, err := root.Find([]string{canonical})
			if err != nil {
				return err
			}
			// Invoke the target's RunE directly. Calling Execute on a
			// subcommand would re-enter the root command and recurse.
			target.SetContext(cmd.Context())
			if err := target.ParseFlags(args[1:]); err != nil {
				return err
			}
			return target.RunE(target, target.Flags().Args())
		},
	}
}
