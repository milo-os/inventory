// SPDX-License-Identifier: AGPL-3.0-only

// SetupIndexers must be called once from main before SetupWithManager for
// each controller. The controllers and the validating webhooks both rely on
// these field indexers to perform O(1) parent -> child lookups.

package controller

import (
	"context"
	"fmt"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Exported index field names. These are the string keys passed to
// client.MatchingFields / IndexField. They are exported because the
// validating webhooks reuse them to count dependents; keeping them stable
// keeps the webhook and controller halves of the operator in sync.
const (
	// IndexSiteRegionRef indexes Sites by .spec.regionRef.name.
	IndexSiteRegionRef = "spec.regionRef.name"

	// IndexClusterControlPlaneSiteRef indexes Clusters by
	// .spec.controlPlaneSiteRef.name (the Site hosting the API server).
	IndexClusterControlPlaneSiteRef = "spec.controlPlaneSiteRef.name"

	// IndexNetworkDeviceSiteRef indexes NetworkDevices by .spec.siteRef.name.
	IndexNetworkDeviceSiteRef = "spec.siteRef.name"

	// IndexNetworkDeviceClusterRef indexes NetworkDevices by
	// .spec.clusterRef.name.
	IndexNetworkDeviceClusterRef = "spec.clusterRef.name"

	// IndexNodeSiteRef indexes Nodes by .spec.siteRef.name.
	IndexNodeSiteRef = "spec.siteRef.name"

	// IndexNodeAssignmentClusterRef indexes Nodes by
	// .spec.assignment.clusterRef.name. Nodes without an assignment are not
	// indexed under this key.
	IndexNodeAssignmentClusterRef = "spec.assignment.clusterRef.name"

	// IndexLinkEndpointName indexes Links by every endpoint name in
	// .spec.endpoints[*].name. A single Link appears under up to two values
	// (one per endpoint). Webhooks and controllers must additionally check
	// the endpoint Kind before treating a match as relevant.
	IndexLinkEndpointName = "spec.endpoints.name"
)

// SetupIndexers registers every field indexer this package relies on
// against the supplied manager's cache. It must be called once, before any
// controller is started, and before the manager begins serving.
func SetupIndexers(ctx context.Context, mgr ctrl.Manager) error {
	idx := mgr.GetFieldIndexer()

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Site{}, IndexSiteRegionRef, func(obj client.Object) []string {
		site, ok := obj.(*inventoryv1alpha1.Site)
		if !ok || site.Spec.RegionRef.Name == "" {
			return nil
		}
		return []string{site.Spec.RegionRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Site.%s: %w", IndexSiteRegionRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Cluster{}, IndexClusterControlPlaneSiteRef, func(obj client.Object) []string {
		cluster, ok := obj.(*inventoryv1alpha1.Cluster)
		if !ok || cluster.Spec.ControlPlaneSiteRef.Name == "" {
			return nil
		}
		return []string{cluster.Spec.ControlPlaneSiteRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Cluster.%s: %w", IndexClusterControlPlaneSiteRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.NetworkDevice{}, IndexNetworkDeviceSiteRef, func(obj client.Object) []string {
		dev, ok := obj.(*inventoryv1alpha1.NetworkDevice)
		if !ok || dev.Spec.SiteRef.Name == "" {
			return nil
		}
		return []string{dev.Spec.SiteRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing NetworkDevice.%s: %w", IndexNetworkDeviceSiteRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.NetworkDevice{}, IndexNetworkDeviceClusterRef, func(obj client.Object) []string {
		dev, ok := obj.(*inventoryv1alpha1.NetworkDevice)
		if !ok || dev.Spec.ClusterRef.Name == "" {
			return nil
		}
		return []string{dev.Spec.ClusterRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing NetworkDevice.%s: %w", IndexNetworkDeviceClusterRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Node{}, IndexNodeSiteRef, func(obj client.Object) []string {
		node, ok := obj.(*inventoryv1alpha1.Node)
		if !ok || node.Spec.SiteRef.Name == "" {
			return nil
		}
		return []string{node.Spec.SiteRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Node.%s: %w", IndexNodeSiteRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Node{}, IndexNodeAssignmentClusterRef, func(obj client.Object) []string {
		node, ok := obj.(*inventoryv1alpha1.Node)
		if !ok || node.Spec.Assignment == nil || node.Spec.Assignment.ClusterRef.Name == "" {
			return nil
		}
		return []string{node.Spec.Assignment.ClusterRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Node.%s: %w", IndexNodeAssignmentClusterRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Link{}, IndexLinkEndpointName, func(obj client.Object) []string {
		link, ok := obj.(*inventoryv1alpha1.Link)
		if !ok {
			return nil
		}
		if len(link.Spec.Endpoints) == 0 {
			return nil
		}
		names := make([]string, 0, len(link.Spec.Endpoints))
		for _, ep := range link.Spec.Endpoints {
			if ep.Name != "" {
				names = append(names, ep.Name)
			}
		}
		return names
	}); err != nil {
		return fmt.Errorf("indexing Link.%s: %w", IndexLinkEndpointName, err)
	}

	return nil
}
