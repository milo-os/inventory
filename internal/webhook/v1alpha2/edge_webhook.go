// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
	"go.miloapis.com/inventory/internal/graph"
)

// +kubebuilder:webhook:path=/validate-graph-inventory-miloapis-com-v1alpha2-edge,mutating=false,failurePolicy=fail,sideEffects=None,groups=graph.inventory.miloapis.com,resources=edges,verbs=create;update,versions=v1alpha2,name=vedge.v1alpha2.graph.inventory.miloapis.com,admissionReviewVersions=v1

var edgeLog = logf.Log.WithName("graph-edge-webhook")

// EdgeValidator validates graph Edge CREATE/UPDATE: the EdgeType exists, both
// endpoint Nodes exist and satisfy the type's endpoint constraints, and the
// attribute bag matches the type schema.
type EdgeValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha2.Edge] = &EdgeValidator{}

func (v *EdgeValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha2.Edge) (admission.Warnings, error) {
	edgeLog.Info("validating create", "name", obj.Name, "type", obj.Spec.Type)
	return nil, graph.ValidateEdge(ctx, v.Client, obj)
}

func (v *EdgeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha2.Edge) (admission.Warnings, error) {
	return nil, graph.ValidateEdge(ctx, v.Client, newObj)
}

func (v *EdgeValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha2.Edge) (admission.Warnings, error) {
	return nil, nil
}

// SetupEdgeWebhookWithManager registers the EdgeValidator with the manager.
func SetupEdgeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha2.Edge{}).
		WithValidator(&EdgeValidator{Client: mgr.GetClient()}).
		Complete()
}

// edgeNames returns up to five edge names for inclusion in a rejection
// message; truncationSuffix summarizes any remainder.
func edgeNames(edges []inventoryv1alpha2.Edge) []string {
	limit := len(edges)
	if limit > maxEdgeNames {
		limit = maxEdgeNames
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, edges[i].Name)
	}
	return out
}

const maxEdgeNames = 5

func truncationSuffix(total int) string {
	if total <= maxEdgeNames {
		return ""
	}
	return fmt.Sprintf(" (and %d more)", total-maxEdgeNames)
}
