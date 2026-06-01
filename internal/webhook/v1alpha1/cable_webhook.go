// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"go.miloapis.com/inventory/internal/controller"
)

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-cable,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=cables,verbs=create;update;delete,versions=v1alpha1,name=vcable.inventory.miloapis.com,admissionReviewVersions=v1

var cableLog = logf.Log.WithName("cable-webhook")

// CableValidator validates Cable operations. On CREATE/UPDATE it verifies
// that both endpoint Ports exist (CEL already enforces exactly-two and
// distinctness). On DELETE it rejects while a Link references the Cable.
type CableValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Cable] = &CableValidator{}

func (v *CableValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Cable) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

func (v *CableValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Cable) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

func (v *CableValidator) validate(ctx context.Context, cable *inventoryv1alpha1.Cable) (admission.Warnings, error) {
	cableLog.Info("validating", "name", cable.Name)

	for _, ep := range cable.Spec.Endpoints {
		var port inventoryv1alpha1.Port
		if err := v.Client.Get(ctx, types.NamespacedName{Name: ep.Name}, &port); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, apierrors.NewBadRequest(fmt.Sprintf("endpoint Port %s not found", ep.Name))
			}
			return nil, err
		}
	}
	return nil, nil
}

// ValidateDelete rejects deletion while a Link references this Cable.
func (v *CableValidator) ValidateDelete(ctx context.Context, cable *inventoryv1alpha1.Cable) (admission.Warnings, error) {
	cableLog.Info("validating delete", "name", cable.Name)

	var links inventoryv1alpha1.LinkList
	if err := v.Client.List(ctx, &links, client.MatchingFields{controller.IndexLinkCableRef: cable.Name}); err != nil {
		return nil, err
	}
	if len(links.Items) == 0 {
		return nil, nil
	}

	names := childNames(links.Items, func(l inventoryv1alpha1.Link) string { return l.Name })
	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Cable %s: %d Link(s) still reference it: %v%s",
		cable.Name, len(links.Items), names, truncationSuffix(len(links.Items)),
	))
}

// SetupCableWebhookWithManager registers the CableValidator with the manager.
func SetupCableWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Cable{}).
		WithValidator(&CableValidator{Client: mgr.GetClient()}).
		Complete()
}
