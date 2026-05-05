// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
)

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-node,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=nodes,verbs=create;update,versions=v1alpha1,name=vnode.inventory.miloapis.com,admissionReviewVersions=v1

var nodeLog = logf.Log.WithName("node-webhook")

// NodeValidator validates Node CREATE/UPDATE operations. v0.1 immutability
// (siteRef) and enum checks are handled by CEL on the type, so this validator
// is currently a pass-through. It is kept so future cross-resource checks
// (for example, validating that `spec.assignment.clusterRef` is consistent
// with `spec.assignment.role`) have a home.
type NodeValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Node] = &NodeValidator{}

func (v *NodeValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Node) (admission.Warnings, error) {
	_ = nodeLog
	return nil, nil
}

func (v *NodeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Node) (admission.Warnings, error) {
	return nil, nil
}

func (v *NodeValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha1.Node) (admission.Warnings, error) {
	return nil, nil
}

// SetupNodeWebhookWithManager registers the NodeValidator with the manager.
func SetupNodeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Node{}).
		WithValidator(&NodeValidator{Client: mgr.GetClient()}).
		Complete()
}
