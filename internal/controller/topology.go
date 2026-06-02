// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// topologyLabelKeys enumerates the well-known topology keys that the
// controllers in this package manage. ensureLabels will only touch keys in
// this set — other labels on the object (user-managed, system, other
// controllers) are left untouched.
var topologyLabelKeys = []string{
	inventoryv1alpha1.TopologyRegionLabel,
	inventoryv1alpha1.TopologySiteLabel,
	inventoryv1alpha1.TopologySiteTypeLabel,
	inventoryv1alpha1.TopologyClusterLabel,
	inventoryv1alpha1.TopologyRackLabel,
}

// rackLabel returns the desired value for TopologyRackLabel given an optional
// Placement. An empty string (no placement) tells ensureLabels to clear the
// key.
func rackLabel(placement *inventoryv1alpha1.Placement) string {
	if placement == nil {
		return ""
	}
	return placement.RackRef.Name
}

// ensureLabels reconciles the well-known topology labels on obj toward the
// desired state described by want. Keys present in want with non-empty
// values are set (overwriting any prior value); keys present in want with
// empty-string values are removed. Keys absent from want are left alone
// — callers must pass the complete desired set for every key they care
// about. Non-topology labels are never modified.
//
// The returned bool is true if ensureLabels modified the label map on obj,
// which is the signal the caller uses to decide whether to Patch.
func ensureLabels(obj client.Object, want map[string]string) bool {
	existing := obj.GetLabels()
	changed := false

	for _, k := range topologyLabelKeys {
		desired, managed := want[k]
		if !managed {
			continue
		}

		current, have := existing[k]

		if desired == "" {
			// Removal requested.
			if have {
				delete(existing, k)
				changed = true
			}
			continue
		}

		if !have || current != desired {
			if existing == nil {
				existing = map[string]string{}
			}
			existing[k] = desired
			changed = true
		}
	}

	if changed {
		obj.SetLabels(existing)
	}
	return changed
}

// copyTopologyLabels returns the well-known topology labels carried by the
// given object. Keys the object lacks map to the empty string so that
// ensureLabels clears any stale value on the inheriting object.
func copyTopologyLabels(from client.Object) map[string]string {
	src := from.GetLabels()
	want := make(map[string]string, len(topologyLabelKeys))
	for _, k := range topologyLabelKeys {
		want[k] = src[k]
	}
	return want
}

// siteLabels returns the topology labels a child of the given Site should
// carry. It composes from labels already on the Site (specifically the
// region label, which the Site controller is responsible for propagating),
// the Site's own name, and its declared type.
//
// If the Site has not yet been labeled with a region (e.g. its own
// reconciliation has not run yet), the returned map uses an empty string
// for TopologyRegionLabel, which causes ensureLabels to clear that key on
// the child. A subsequent requeue from the Site watch re-propagates the
// correct value once the Site is labeled.
func siteLabels(site *inventoryv1alpha1.Site) map[string]string {
	var region string
	if site.Labels != nil {
		region = site.Labels[inventoryv1alpha1.TopologyRegionLabel]
	}
	return map[string]string{
		inventoryv1alpha1.TopologyRegionLabel:   region,
		inventoryv1alpha1.TopologySiteLabel:     site.Name,
		inventoryv1alpha1.TopologySiteTypeLabel: string(site.Spec.Type),
	}
}

// clusterLabels returns the topology labels a child of the given Cluster
// should carry. It composes from labels already on the Cluster (region,
// site, site-type — which the Cluster controller itself propagates from
// the parent Site) plus the Cluster's own name.
func clusterLabels(cluster *inventoryv1alpha1.Cluster) map[string]string {
	var region, site, siteType string
	if cluster.Labels != nil {
		region = cluster.Labels[inventoryv1alpha1.TopologyRegionLabel]
		site = cluster.Labels[inventoryv1alpha1.TopologySiteLabel]
		siteType = cluster.Labels[inventoryv1alpha1.TopologySiteTypeLabel]
	}
	return map[string]string{
		inventoryv1alpha1.TopologyRegionLabel:   region,
		inventoryv1alpha1.TopologySiteLabel:     site,
		inventoryv1alpha1.TopologySiteTypeLabel: siteType,
		inventoryv1alpha1.TopologyClusterLabel:  cluster.Name,
	}
}
