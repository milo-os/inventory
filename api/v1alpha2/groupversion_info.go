// SPDX-License-Identifier: AGPL-3.0-only

// Package v1alpha2 contains API Schema definitions for the property-graph
// model of the inventory.miloapis.com API group. It replaces the per-kind
// v1alpha1 hierarchy with two generic kinds -- Node (graph vertex) and Edge
// (graph relationship) -- whose shape is described by the NodeType and
// EdgeType schema-registry kinds. See RFC milo-os/inventory#43.
//
// The prototype uses the group graph.inventory.miloapis.com so it installs
// side-by-side with the v1alpha1 hierarchy (whose `Node` kind would otherwise
// collide at the CRD level -- the very collision RFC #43 resolves). The
// production migration reclaims the inventory.miloapis.com group once
// v1alpha1 is removed.
//
// +kubebuilder:object:generate=true
// +groupName=graph.inventory.miloapis.com
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "graph.inventory.miloapis.com", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
