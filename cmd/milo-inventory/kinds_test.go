// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The point of sourcing `kinds` from the scheme is that a kind the API serves
// but the plugin does not wire up is reported, not silently dropped.
func TestCollectKindsCoversEveryServedKind(t *testing.T) {
	served := servedKinds()
	if len(served) == 0 {
		t.Fatal("servedKinds() returned nothing; scheme registration changed?")
	}
	list := collectKinds()
	if len(list.Kinds) != len(served) {
		t.Errorf("collectKinds() reported %d kinds, scheme serves %d", len(list.Kinds), len(served))
	}
	for _, want := range served {
		var found bool
		for _, got := range list.Kinds {
			if got.Kind == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("served kind %q missing from kinds output", want)
		}
	}
}

// servedKinds filters the scheme's List and options types; a leak would show
// up here as a bogus row.
func TestServedKindsExcludesNonResourceTypes(t *testing.T) {
	for _, k := range servedKinds() {
		if strings.HasSuffix(k, "List") || strings.HasSuffix(k, "Options") {
			t.Errorf("servedKinds() leaked non-resource type %q", k)
		}
	}
}

func TestCollectKindsReportsApplyAndAliases(t *testing.T) {
	byKind := map[string]kindInfo{}
	for _, k := range collectKinds().Kinds {
		byKind[k.Kind] = k
	}

	site, ok := byKind["Site"]
	if !ok {
		t.Fatal("Site missing")
	}
	if !site.Supported || !site.Apply {
		t.Errorf("Site should be supported and applyable, got %+v", site)
	}
	if site.Plural != "sites" {
		t.Errorf("Site plural = %q, want sites", site.Plural)
	}
	if !contains(site.Aliases, "site") {
		t.Errorf("Site aliases %v should include the singular", site.Aliases)
	}
	if contains(site.Aliases, "sites") {
		t.Errorf("Site aliases %v should not repeat the plural", site.Aliases)
	}

	vm, ok := byKind["VirtualMachine"]
	if !ok {
		t.Fatal("VirtualMachine missing")
	}
	for _, want := range []string{"vm", "vms", "virtualmachine"} {
		if !contains(vm.Aliases, want) {
			t.Errorf("VirtualMachine aliases %v should include %q", vm.Aliases, want)
		}
	}
}

// Every kind the plugin claims to support must resolve through `get`, or the
// two maps have drifted apart.
func TestKindsAgreeWithGet(t *testing.T) {
	for _, k := range collectKinds().Kinds {
		if !k.Supported {
			continue
		}
		if _, ok := canonicalKind(k.Plural); !ok {
			t.Errorf("kinds reports %q supported but get rejects %q", k.Kind, k.Plural)
		}
		for _, alias := range k.Aliases {
			if _, ok := canonicalKind(alias); !ok {
				t.Errorf("kinds advertises alias %q for %s but get rejects it", alias, k.Kind)
			}
		}
	}
}

func TestKindsListMarshalsAsJSON(t *testing.T) {
	b, err := json.Marshal(collectKinds())
	if err != nil {
		t.Fatal(err)
	}
	var round struct {
		Kinds []map[string]any `json:"kinds"`
	}
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if len(round.Kinds) == 0 {
		t.Fatal("kinds key missing or empty in JSON output")
	}
	if _, ok := round.Kinds[0]["kind"]; !ok {
		t.Errorf("entry missing 'kind' field: %v", round.Kinds[0])
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
