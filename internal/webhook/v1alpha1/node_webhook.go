// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"go.miloapis.com/inventory/internal/controller"
)

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-node,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=nodes,verbs=create;update;delete,versions=v1alpha1,name=vnode.inventory.miloapis.com,admissionReviewVersions=v1

var nodeLog = logf.Log.WithName("node-webhook")

// NodeValidator validates Node operations. Immutability (siteRef) and enum
// checks are handled by CEL on the type. On CREATE/UPDATE this validator
// enforces the cross-resource placement rules (rack existence, fit, and
// non-overlap) that CEL cannot express. On DELETE it rejects removal while a
// VirtualMachine is hosted on the Node.
type NodeValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Node] = &NodeValidator{}

func (v *NodeValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Node) (admission.Warnings, error) {
	nodeLog.Info("validating create", "name", obj.Name)
	return nil, validatePlacement(ctx, v.Client, obj.Spec.Placement, "Node", obj.Name)
}

func (v *NodeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Node) (admission.Warnings, error) {
	return nil, validatePlacement(ctx, v.Client, newObj.Spec.Placement, "Node", newObj.Name)
}

// ValidateDelete rejects deletion while any VirtualMachine is hosted on the
// Node.
func (v *NodeValidator) ValidateDelete(ctx context.Context, node *inventoryv1alpha1.Node) (admission.Warnings, error) {
	nodeLog.Info("validating delete", "name", node.Name)

	var vms inventoryv1alpha1.VirtualMachineList
	if err := v.Client.List(ctx, &vms, client.MatchingFields{controller.IndexVirtualMachineHostRef: node.Name}); err != nil {
		return nil, err
	}
	if len(vms.Items) == 0 {
		return nil, nil
	}

	names := childNames(vms.Items, func(x inventoryv1alpha1.VirtualMachine) string { return x.Name })
	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Node %s: %d VirtualMachine(s) still hosted on it: %v%s",
		node.Name, len(vms.Items), names, truncationSuffix(len(vms.Items)),
	))
}

// SetupNodeWebhookWithManager registers the NodeValidator with the manager.
func SetupNodeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Node{}).
		WithValidator(&NodeValidator{Client: mgr.GetClient()}).
		Complete()
}
