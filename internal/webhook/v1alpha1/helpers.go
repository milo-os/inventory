// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"go.miloapis.com/inventory/internal/controller"
)

// maxChildNames is the maximum number of child names included inline in a
// deletion-rejection message. Any additional children are summarized by
// truncationSuffix.
const maxChildNames = 5

// childNames returns up to maxChildNames names extracted from items via the
// given name accessor. The generic helper keeps the webhook files free of
// per-kind boilerplate when formatting rejection messages.
func childNames[T any](items []T, name func(T) string) []string {
	limit := len(items)
	if limit > maxChildNames {
		limit = maxChildNames
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, name(items[i]))
	}
	return out
}

// truncationSuffix returns a " (and N more)" fragment when total exceeds
// maxChildNames, or the empty string otherwise. The returned fragment is
// designed to be appended directly after the child-name list.
func truncationSuffix(total int) string {
	if total <= maxChildNames {
		return ""
	}
	return fmt.Sprintf(" (and %d more)", total-maxChildNames)
}

// occupant is one device's footprint within a rack face, used by the overlap
// check. kind/name identify the device so the validator can skip the object
// being admitted.
type occupant struct {
	kind  string
	name  string
	face  inventoryv1alpha1.RackFace
	start int32
	end   int32
}

// validatePlacement enforces that placement (if set) refers to an existing
// Rack, fits within the Rack's HeightU, and does not overlap any other device
// mounted on the same face. selfKind/selfName identify the object being
// admitted so it is not compared against itself on UPDATE.
func validatePlacement(ctx context.Context, c client.Client, placement *inventoryv1alpha1.Placement, selfKind, selfName string) error {
	if placement == nil {
		return nil
	}

	var rack inventoryv1alpha1.Rack
	if err := c.Get(ctx, types.NamespacedName{Name: placement.RackRef.Name}, &rack); err != nil {
		if apierrors.IsNotFound(err) {
			return apierrors.NewBadRequest(fmt.Sprintf("referenced Rack %s not found", placement.RackRef.Name))
		}
		return err
	}

	// here be dragons
	face := placement.Face
	if face == "" {
		face = inventoryv1alpha1.RackFaceFront
	}
	start := placement.StartUnit
	end := placement.StartUnit + placement.UnitHeight - 1

	if end > rack.Spec.HeightU {
		return apierrors.NewBadRequest(fmt.Sprintf(
			"placement U%d-U%d does not fit within Rack %s (height %dU)",
			start, end, rack.Name, rack.Spec.HeightU,
		))
	}

	occupants, err := rackOccupants(ctx, c, rack.Name, selfKind, selfName)
	if err != nil {
		return err
	}
	for _, o := range occupants {
		if o.face != face {
			continue
		}
		if start <= o.end && o.start <= end {
			return apierrors.NewBadRequest(fmt.Sprintf(
				"placement U%d-U%d (%s) overlaps %s %s at U%d-U%d in Rack %s",
				start, end, face, o.kind, o.name, o.start, o.end, rack.Name,
			))
		}
	}
	return nil
}

// rackOccupants returns the footprints of every Node and NetworkDevice placed
// in the named rack, excluding the device identified by selfKind/selfName.
func rackOccupants(ctx context.Context, c client.Client, rackName, selfKind, selfName string) ([]occupant, error) {
	var out []occupant

	var nodes inventoryv1alpha1.NodeList
	if err := c.List(ctx, &nodes, client.MatchingFields{controller.IndexNodePlacementRackRef: rackName}); err != nil {
		return nil, err
	}
	for _, n := range nodes.Items {
		if selfKind == "Node" && n.Name == selfName {
			continue
		}
		p := n.Spec.Placement
		out = append(out, occupant{
			kind:  "Node",
			name:  n.Name,
			face:  placementFace(p),
			start: p.StartUnit,
			end:   p.StartUnit + p.UnitHeight - 1,
		})
	}

	var devices inventoryv1alpha1.NetworkDeviceList
	if err := c.List(ctx, &devices, client.MatchingFields{controller.IndexNetworkDevicePlacementRackRef: rackName}); err != nil {
		return nil, err
	}
	for _, d := range devices.Items {
		if selfKind == "NetworkDevice" && d.Name == selfName {
			continue
		}
		p := d.Spec.Placement
		out = append(out, occupant{
			kind:  "NetworkDevice",
			name:  d.Name,
			face:  placementFace(p),
			start: p.StartUnit,
			end:   p.StartUnit + p.UnitHeight - 1,
		})
	}

	return out, nil
}

func placementFace(p *inventoryv1alpha1.Placement) inventoryv1alpha1.RackFace {
	if p.Face == "" {
		return inventoryv1alpha1.RackFaceFront
	}
	return p.Face
}
