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

	IndexSiteProviderRef = "spec.providerRef.name"

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

	// IndexRackSiteRef indexes Racks by .spec.siteRef.name.
	IndexRackSiteRef = "spec.siteRef.name"

	// IndexNodePlacementRackRef indexes Nodes by
	// .spec.placement.rackRef.name. Nodes without a placement are not indexed.
	IndexNodePlacementRackRef = "spec.placement.rackRef.name"

	// IndexNetworkDevicePlacementRackRef indexes NetworkDevices by
	// .spec.placement.rackRef.name. Devices without a placement are not
	// indexed.
	IndexNetworkDevicePlacementRackRef = "spec.placement.rackRef.name"

	// IndexLinkEndpointName indexes Links by every endpoint name in
	// .spec.endpoints[*].name. A single Link appears under up to two values
	// (one per endpoint). Webhooks and controllers must additionally check
	// the endpoint Kind before treating a match as relevant.
	IndexLinkEndpointName = "spec.endpoints.name"

	// IndexPortDeviceRef indexes Ports by .spec.deviceRef.name. The index is
	// kind-agnostic; consumers must additionally check deviceRef.kind.
	IndexPortDeviceRef = "spec.deviceRef.name"

	// IndexCableEndpointName indexes Cables by every endpoint Port name in
	// .spec.endpoints[*].name. A single Cable appears under up to two values.
	IndexCableEndpointName = "spec.cable.endpoints.name"

	// IndexLinkCableRef indexes Links by every cable name in
	// .spec.cableRefs[*].name.
	IndexLinkCableRef = "spec.cableRefs.name"

	// IndexCircuitProviderRef indexes Circuits by .spec.providerRef.name.
	IndexCircuitProviderRef = "spec.providerRef.name"

	// IndexCircuitSiteEndpoint indexes Circuits by the names of their Site-kind
	// endpoints (aEnd/zEnd where kind == Site). Port-kind endpoints are not
	// indexed here.
	IndexCircuitSiteEndpoint = "spec.circuit.siteEndpoints.name"

	// IndexVirtualMachineHostRef indexes VirtualMachines by .spec.hostRef.name.
	IndexVirtualMachineHostRef = "spec.hostRef.name"

	// IndexVirtualMachineProviderRef indexes VirtualMachines by
	// .spec.providerRef.name. VMs without a provider are not indexed.
	IndexVirtualMachineProviderRef = "spec.vm.providerRef.name"
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

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Site{}, IndexSiteProviderRef, func(obj client.Object) []string {
		site, ok := obj.(*inventoryv1alpha1.Site)
		if !ok || site.Spec.ProviderRef == nil || site.Spec.ProviderRef.Name == "" {
			return nil
		}
		return []string{site.Spec.ProviderRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Site.%s: %w", IndexSiteProviderRef, err)
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

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Rack{}, IndexRackSiteRef, func(obj client.Object) []string {
		rack, ok := obj.(*inventoryv1alpha1.Rack)
		if !ok || rack.Spec.SiteRef.Name == "" {
			return nil
		}
		return []string{rack.Spec.SiteRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Rack.%s: %w", IndexRackSiteRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Node{}, IndexNodePlacementRackRef, func(obj client.Object) []string {
		node, ok := obj.(*inventoryv1alpha1.Node)
		if !ok || node.Spec.Placement == nil || node.Spec.Placement.RackRef.Name == "" {
			return nil
		}
		return []string{node.Spec.Placement.RackRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Node.%s: %w", IndexNodePlacementRackRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.NetworkDevice{}, IndexNetworkDevicePlacementRackRef, func(obj client.Object) []string {
		dev, ok := obj.(*inventoryv1alpha1.NetworkDevice)
		if !ok || dev.Spec.Placement == nil || dev.Spec.Placement.RackRef.Name == "" {
			return nil
		}
		return []string{dev.Spec.Placement.RackRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing NetworkDevice.%s: %w", IndexNetworkDevicePlacementRackRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Port{}, IndexPortDeviceRef, func(obj client.Object) []string {
		port, ok := obj.(*inventoryv1alpha1.Port)
		if !ok || port.Spec.DeviceRef.Name == "" {
			return nil
		}
		return []string{port.Spec.DeviceRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Port.%s: %w", IndexPortDeviceRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Cable{}, IndexCableEndpointName, func(obj client.Object) []string {
		cable, ok := obj.(*inventoryv1alpha1.Cable)
		if !ok || len(cable.Spec.Endpoints) == 0 {
			return nil
		}
		names := make([]string, 0, len(cable.Spec.Endpoints))
		for _, ep := range cable.Spec.Endpoints {
			if ep.Name != "" {
				names = append(names, ep.Name)
			}
		}
		return names
	}); err != nil {
		return fmt.Errorf("indexing Cable.%s: %w", IndexCableEndpointName, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Link{}, IndexLinkCableRef, func(obj client.Object) []string {
		link, ok := obj.(*inventoryv1alpha1.Link)
		if !ok || len(link.Spec.CableRefs) == 0 {
			return nil
		}
		names := make([]string, 0, len(link.Spec.CableRefs))
		for _, ref := range link.Spec.CableRefs {
			if ref.Name != "" {
				names = append(names, ref.Name)
			}
		}
		return names
	}); err != nil {
		return fmt.Errorf("indexing Link.%s: %w", IndexLinkCableRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Circuit{}, IndexCircuitProviderRef, func(obj client.Object) []string {
		circuit, ok := obj.(*inventoryv1alpha1.Circuit)
		if !ok || circuit.Spec.ProviderRef.Name == "" {
			return nil
		}
		return []string{circuit.Spec.ProviderRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing Circuit.%s: %w", IndexCircuitProviderRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.Circuit{}, IndexCircuitSiteEndpoint, func(obj client.Object) []string {
		circuit, ok := obj.(*inventoryv1alpha1.Circuit)
		if !ok {
			return nil
		}
		var names []string
		for _, ep := range []inventoryv1alpha1.CircuitEndpoint{circuit.Spec.AEnd, circuit.Spec.ZEnd} {
			if ep.Kind == inventoryv1alpha1.CircuitEndpointKindSite && ep.Name != "" {
				names = append(names, ep.Name)
			}
		}
		return names
	}); err != nil {
		return fmt.Errorf("indexing Circuit.%s: %w", IndexCircuitSiteEndpoint, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.VirtualMachine{}, IndexVirtualMachineHostRef, func(obj client.Object) []string {
		vm, ok := obj.(*inventoryv1alpha1.VirtualMachine)
		if !ok || vm.Spec.HostRef.Name == "" {
			return nil
		}
		return []string{vm.Spec.HostRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing VirtualMachine.%s: %w", IndexVirtualMachineHostRef, err)
	}

	if err := idx.IndexField(ctx, &inventoryv1alpha1.VirtualMachine{}, IndexVirtualMachineProviderRef, func(obj client.Object) []string {
		vm, ok := obj.(*inventoryv1alpha1.VirtualMachine)
		if !ok || vm.Spec.ProviderRef == nil || vm.Spec.ProviderRef.Name == "" {
			return nil
		}
		return []string{vm.Spec.ProviderRef.Name}
	}); err != nil {
		return fmt.Errorf("indexing VirtualMachine.%s: %w", IndexVirtualMachineProviderRef, err)
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
