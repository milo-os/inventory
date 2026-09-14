// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"sort"
	"strings"
	"testing"
)

const sampleManifest = `
apiVersion: inventory.miloapis.com/v1alpha1
kind: Node
metadata:
  name: node-a
spec:
  siteRef:
    name: us-central-1a
  hardware:
    cpuCores: 8
    cpuArchitecture: amd64
    memoryBytes: 1073741824
---
apiVersion: inventory.miloapis.com/v1alpha1
kind: Provider
metadata:
  name: netactuate
spec:
  displayName: NetActuate
  type: Hosting
---
apiVersion: inventory.miloapis.com/v1alpha1
kind: Site
metadata:
  name: us-central-1a
spec:
  displayName: Dallas
  type: AvailabilityZone
  regionRef:
    name: us-central-1
`

func TestReadManifestsParsesAndOrders(t *testing.T) {
	objs, err := readManifests(strings.NewReader(sampleManifest), []string{"-"})
	if err != nil {
		t.Fatalf("readManifests: %v", err)
	}
	if len(objs) != 3 {
		t.Fatalf("got %d objects, want 3", len(objs))
	}
	sort.SliceStable(objs, func(i, j int) bool { return objs[i].order < objs[j].order })
	gotKinds := []string{objs[0].kind, objs[1].kind, objs[2].kind}
	want := []string{"Provider", "Site", "Node"}
	for i := range want {
		if gotKinds[i] != want[i] {
			t.Errorf("order[%d] = %s, want %s (full: %v)", i, gotKinds[i], want[i], gotKinds)
		}
	}
	// GVK must be set on each object so server-side apply has apiVersion/kind.
	for _, o := range objs {
		if o.obj.GetObjectKind().GroupVersionKind().Kind == "" {
			t.Errorf("%s/%s missing GVK", o.kind, o.obj.GetName())
		}
	}
}

func TestReadManifestsRejectsUnsupportedKind(t *testing.T) {
	// A kind outside the inventory group entirely: apply handles all twelve
	// inventory kinds, so the rejection path needs something it never will.
	const m = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: not-inventory
`
	_, err := readManifests(strings.NewReader(m), []string{"-"})
	if err == nil {
		t.Fatal("want an error for a non-inventory kind, got nil")
	}
}

func TestReadManifestsEmpty(t *testing.T) {
	objs, err := readManifests(strings.NewReader("\n---\n"), []string{"-"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(objs) != 0 {
		t.Fatalf("got %d objects, want 0", len(objs))
	}
}

func TestKindOrder(t *testing.T) {
	for _, k := range []string{
		"Provider", "Region", "Site", "Rack", "Cluster", "Node",
		"NetworkDevice", "VirtualMachine", "Port", "Cable", "Link", "Circuit",
	} {
		if _, ok := kindOrder(k); !ok {
			t.Errorf("%s should be applyable", k)
		}
	}
	if _, ok := kindOrder("ConfigMap"); ok {
		t.Error("ConfigMap should not be applyable")
	}
}

// Apply order has to satisfy the admission-time reference checks: a referent
// must already exist when its dependant is admitted.
func TestKindOrderSatisfiesDependencies(t *testing.T) {
	for _, dep := range []struct{ before, after string }{
		{"Provider", "Site"},
		{"Provider", "Circuit"},
		{"Provider", "VirtualMachine"},
		{"Region", "Site"},
		{"Site", "Rack"},
		{"Site", "Cluster"},
		{"Site", "Node"},
		{"Rack", "Node"},          // Node.placement.rackRef
		{"Rack", "NetworkDevice"}, // NetworkDevice.placement.rackRef
		{"Cluster", "Node"},
		{"Cluster", "NetworkDevice"},
		{"Node", "VirtualMachine"}, // VirtualMachine.hostRef
		{"Node", "Port"},           // Port.deviceRef
		{"NetworkDevice", "Port"},
		{"Port", "Cable"},   // Cable.endpoints
		{"Cable", "Link"},   // Link.cableRefs
		{"Port", "Circuit"}, // Circuit endpoints may terminate at a Port
	} {
		b, okB := kindOrder(dep.before)
		a, okA := kindOrder(dep.after)
		if !okB || !okA {
			t.Fatalf("both %s and %s must be applyable", dep.before, dep.after)
		}
		if b >= a {
			t.Errorf("%s (%d) must be applied before %s (%d)", dep.before, b, dep.after, a)
		}
	}
}
