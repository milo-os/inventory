// SPDX-License-Identifier: AGPL-3.0-only

package graph

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
)

const (
	// IndexEdgeEndpointName indexes Edges by both endpoint Node names
	// (spec.from.name and spec.to.name). A single Edge appears under up to two
	// values. This one index replaces the per-kind reference indexers of the
	// v1alpha1 model: the Node delete-guard counts Edges touching a Node, and
	// the Edge controller wakes on endpoint changes through it.
	IndexEdgeEndpointName = "spec.endpoints.name"
)

// SetupIndexers registers the graph field indexers against the manager's
// cache. It must be called once before the graph controllers start.
func SetupIndexers(ctx context.Context, mgr ctrl.Manager) error {
	idx := mgr.GetFieldIndexer()

	if err := idx.IndexField(ctx, &inventoryv1alpha2.Edge{}, IndexEdgeEndpointName, func(obj client.Object) []string {
		edge, ok := obj.(*inventoryv1alpha2.Edge)
		if !ok {
			return nil
		}
		var names []string
		if edge.Spec.From.Name != "" {
			names = append(names, edge.Spec.From.Name)
		}
		if edge.Spec.To.Name != "" {
			names = append(names, edge.Spec.To.Name)
		}
		return names
	}); err != nil {
		return fmt.Errorf("indexing Edge.%s: %w", IndexEdgeEndpointName, err)
	}

	return nil
}

// EdgesReferencing returns the Edges that have the named Node as either
// endpoint. The Node delete-guard webhook uses this.
func EdgesReferencing(ctx context.Context, c client.Client, nodeName string) ([]inventoryv1alpha2.Edge, error) {
	var edges inventoryv1alpha2.EdgeList
	if err := c.List(ctx, &edges, client.MatchingFields{IndexEdgeEndpointName: nodeName}); err != nil {
		return nil, err
	}
	return edges.Items, nil
}
