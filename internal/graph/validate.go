// SPDX-License-Identifier: AGPL-3.0-only

// Package graph holds the reconcilers, field indexers, and shared validation
// logic for the v1alpha2 property-graph model (Node + Edge, described by
// NodeType + EdgeType). The admission webhooks in internal/webhook/v1alpha2
// reuse the validation helpers here so the webhook and controller halves stay
// in sync. See RFC milo-os/inventory#43.
package graph

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
)

// ValidateAttributes checks an attribute bag against a closed schema. It
// rejects unknown keys, missing required keys, values that do not parse as
// their declared type, and String values outside a declared enum. The schema
// owner string (e.g. `NodeType "Site"`) is used only in error messages.
func ValidateAttributes(owner string, schema []inventoryv1alpha2.AttributeSchema, attrs map[string]string) error {
	byKey := make(map[string]inventoryv1alpha2.AttributeSchema, len(schema))
	for _, s := range schema {
		byKey[s.Key] = s
	}

	for key, val := range attrs {
		s, ok := byKey[key]
		if !ok {
			return apierrors.NewBadRequest(fmt.Sprintf("%s does not allow attribute %q", owner, key))
		}
		if err := validateValue(owner, s, val); err != nil {
			return err
		}
	}

	for _, missing := range missingRequired(byKey, attrs) {
		return apierrors.NewBadRequest(fmt.Sprintf("%s requires attribute %q", owner, missing))
	}
	return nil
}

func missingRequired(byKey map[string]inventoryv1alpha2.AttributeSchema, attrs map[string]string) []string {
	var missing []string
	for key, s := range byKey {
		if !s.Required {
			continue
		}
		if _, ok := attrs[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func validateValue(owner string, s inventoryv1alpha2.AttributeSchema, val string) error {
	switch s.Type {
	case inventoryv1alpha2.AttributeInteger:
		if _, err := strconv.ParseInt(val, 10, 64); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("%s attribute %q must be an integer, got %q", owner, s.Key, val))
		}
	case inventoryv1alpha2.AttributeFloat:
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("%s attribute %q must be a float, got %q", owner, s.Key, val))
		}
	case inventoryv1alpha2.AttributeBoolean:
		if val != "true" && val != "false" {
			return apierrors.NewBadRequest(fmt.Sprintf("%s attribute %q must be true or false, got %q", owner, s.Key, val))
		}
	case inventoryv1alpha2.AttributeString:
		if len(s.Enum) > 0 && !contains(s.Enum, val) {
			return apierrors.NewBadRequest(fmt.Sprintf("%s attribute %q must be one of %v, got %q", owner, s.Key, s.Enum, val))
		}
	}
	return nil
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// ValidateNode resolves the Node's NodeType and validates its attributes
// against that type's schema. A missing NodeType is rejected.
func ValidateNode(ctx context.Context, c client.Client, node *inventoryv1alpha2.Node) error {
	var nt inventoryv1alpha2.NodeType
	if err := c.Get(ctx, types.NamespacedName{Name: node.Spec.Type}, &nt); err != nil {
		if apierrors.IsNotFound(err) {
			return apierrors.NewBadRequest(fmt.Sprintf("NodeType %q not found", node.Spec.Type))
		}
		return err
	}
	return ValidateAttributes(fmt.Sprintf("NodeType %q", nt.Name), nt.Spec.Attributes, node.Spec.Attributes)
}

// ValidateEdge resolves the Edge's EdgeType, verifies both endpoint Nodes
// exist and satisfy the type's endpoint constraints, and validates the edge's
// attributes against the type's schema.
func ValidateEdge(ctx context.Context, c client.Client, edge *inventoryv1alpha2.Edge) error {
	var et inventoryv1alpha2.EdgeType
	if err := c.Get(ctx, types.NamespacedName{Name: edge.Spec.Type}, &et); err != nil {
		if apierrors.IsNotFound(err) {
			return apierrors.NewBadRequest(fmt.Sprintf("EdgeType %q not found", edge.Spec.Type))
		}
		return err
	}

	from, err := getNode(ctx, c, edge.Spec.From.Name, "from")
	if err != nil {
		return err
	}
	to, err := getNode(ctx, c, edge.Spec.To.Name, "to")
	if err != nil {
		return err
	}

	if cs := et.Spec.Endpoints.FromTypes; len(cs) > 0 && !contains(cs, from.Spec.Type) {
		return apierrors.NewBadRequest(fmt.Sprintf(
			"EdgeType %q does not allow a %q node as `from` (allowed: %v)", et.Name, from.Spec.Type, cs))
	}
	if cs := et.Spec.Endpoints.ToTypes; len(cs) > 0 && !contains(cs, to.Spec.Type) {
		return apierrors.NewBadRequest(fmt.Sprintf(
			"EdgeType %q does not allow a %q node as `to` (allowed: %v)", et.Name, to.Spec.Type, cs))
	}

	return ValidateAttributes(fmt.Sprintf("EdgeType %q", et.Name), et.Spec.Attributes, edge.Spec.Attributes)
}

func getNode(ctx context.Context, c client.Client, name, side string) (*inventoryv1alpha2.Node, error) {
	var n inventoryv1alpha2.Node
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &n); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, apierrors.NewBadRequest(fmt.Sprintf("%s endpoint Node %q not found", side, name))
		}
		return nil, err
	}
	return &n, nil
}
