// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCanonicalKind(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"sites", "sites"},
		{"site", "sites"},
		{"Site", "sites"},
		{"SITE", "sites"},
		{"  site  ", "sites"},
		{"nodes", "nodes"},
		{"node", "nodes"},
		{"providers", "providers"},
		{"regions", "regions"},
		{"clusters", "clusters"},
		{"rack", "racks"},
		{"racks", "racks"},
		{"networkdevice", "networkdevices"},
		{"NetworkDevice", "networkdevices"},
		{"nd", "networkdevices"},
		{"port", "ports"},
		{"cable", "cables"},
		{"circuit", "circuits"},
		{"link", "links"},
		{"virtualmachine", "virtualmachines"},
		{"vm", "virtualmachines"},
		{"vms", "virtualmachines"},
	} {
		got, ok := canonicalKind(tc.in)
		if !ok {
			t.Errorf("canonicalKind(%q) not resolved", tc.in)
			continue
		}
		if got != tc.want {
			t.Errorf("canonicalKind(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	for _, in := range []string{"bogus", "", "sitez", "edge"} {
		if _, ok := canonicalKind(in); ok {
			t.Errorf("canonicalKind(%q) resolved, want rejection", in)
		}
	}
}

func TestKnownKinds(t *testing.T) {
	kinds := knownKinds()
	want := []string{
		"cables", "circuits", "clusters", "links", "networkdevices", "nodes",
		"ports", "providers", "racks", "regions", "sites", "virtualmachines",
	}
	if len(kinds) != len(want) {
		t.Fatalf("knownKinds() = %v, want %v", kinds, want)
	}
	for i, k := range want {
		if kinds[i] != k {
			t.Errorf("knownKinds()[%d] = %q, want %q (sorted, deduped)", i, kinds[i], k)
		}
	}
}

// getDispatchesToKind is the behaviour the `get` shim exists for: the kind
// argument has to reach the matching list subcommand rather than recursing
// back into the root command.
func TestGetDispatchesToKind(t *testing.T) {
	var dispatched string
	root := &cobra.Command{Use: "inventory"}
	root.PersistentFlags().StringP("output", "o", "table", "")
	for _, kind := range knownKinds() {
		root.AddCommand(&cobra.Command{
			Use:  kind,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				dispatched = cmd.Use
				return nil
			},
		})
	}
	get := newGetCmd(root)
	root.AddCommand(get)

	for _, tc := range []struct{ arg, want string }{
		{"sites", "sites"},
		{"Site", "sites"},
		{"node", "nodes"},
	} {
		dispatched = ""
		get.SetArgs([]string{tc.arg})
		if err := get.RunE(get, []string{tc.arg}); err != nil {
			t.Fatalf("get %s: %v", tc.arg, err)
		}
		if dispatched != tc.want {
			t.Errorf("get %s dispatched to %q, want %q", tc.arg, dispatched, tc.want)
		}
	}
}

func TestGetRejectsUnknownKind(t *testing.T) {
	root := &cobra.Command{Use: "inventory"}
	get := newGetCmd(root)
	err := get.RunE(get, []string{"bogus"})
	if err == nil {
		t.Fatal("get bogus should error")
	}
	if !strings.Contains(err.Error(), "unknown inventory kind") {
		t.Errorf("error = %q, want it to name the problem", err)
	}
	for _, kind := range knownKinds() {
		if !strings.Contains(err.Error(), kind) {
			t.Errorf("error %q should list valid kind %q", err, kind)
		}
	}
}
