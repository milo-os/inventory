// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
	"go.miloapis.com/inventory/internal/graph"
)

// +kubebuilder:webhook:path=/validate-graph-inventory-miloapis-com-v1alpha2-node,mutating=false,failurePolicy=fail,sideEffects=None,groups=graph.inventory.miloapis.com,resources=nodes,verbs=create;update;delete,versions=v1alpha2,name=vnode.v1alpha2.graph.inventory.miloapis.com,admissionReviewVersions=v1

var nodeLog = logf.Log.WithName("graph-node-webhook")

// NodeValidator validates graph Node admission. CREATE/UPDATE validate the
// attribute bag against the Node's NodeType; DELETE rejects removal while any
// Edge still references the Node (the generic delete-guard that replaces the
// per-kind v1alpha1 guards).
type NodeValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha2.Node] = &NodeValidator{}

func (v *NodeValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha2.Node) (admission.Warnings, error) {
	nodeLog.Info("validating create", "name", obj.Name, "type", obj.Spec.Type)
	return nil, graph.ValidateNode(ctx, v.Client, obj)
}

func (v *NodeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha2.Node) (admission.Warnings, error) {
	return nil, graph.ValidateNode(ctx, v.Client, newObj)
}

func (v *NodeValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha2.Node) (admission.Warnings, error) {
	nodeLog.Info("validating delete", "name", obj.Name)

	edges, err := graph.EdgesReferencing(ctx, v.Client, obj.Name)
	if err != nil {
		return nil, err
	}
	if len(edges) > 0 {
		return nil, apierrors.NewBadRequest(fmt.Sprintf(
			"cannot delete Node %s: %d Edge(s) still reference it: %v%s",
			obj.Name, len(edges), edgeNames(edges), truncationSuffix(len(edges)),
		))
	}
	return nil, nil
}

// SetupNodeWebhookWithManager registers the NodeValidator with the manager.
func SetupNodeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha2.Node{}).
		WithValidator(&NodeValidator{Client: mgr.GetClient()}).
		Complete()
}
