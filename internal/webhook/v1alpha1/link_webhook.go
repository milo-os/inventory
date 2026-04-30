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
)

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-link,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=links,verbs=create;update,versions=v1alpha1,name=vlink.inventory.miloapis.com,admissionReviewVersions=v1

var linkLog = logf.Log.WithName("link-webhook")

// LinkValidator validates Link CREATE/UPDATE operations. It verifies that
// both endpoints reference an existing asset of the declared Kind. CEL on
// the type already enforces exactly-two, distinct, and allowed-kind checks.
type LinkValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Link] = &LinkValidator{}

func (v *LinkValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Link) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

func (v *LinkValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Link) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

func (v *LinkValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha1.Link) (admission.Warnings, error) {
	return nil, nil
}

// validate checks every endpoint for existence of the referenced asset.
func (v *LinkValidator) validate(ctx context.Context, link *inventoryv1alpha1.Link) (admission.Warnings, error) {
	linkLog.Info("validating", "name", link.Name)

	for _, ep := range link.Spec.Endpoints {
		if err := v.checkEndpoint(ctx, ep); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// checkEndpoint resolves a single endpoint and returns an apierror suitable
// for returning directly from the validator.
func (v *LinkValidator) checkEndpoint(ctx context.Context, ep inventoryv1alpha1.AssetReference) error {
	var target client.Object
	switch ep.Kind {
	case "Site":
		target = &inventoryv1alpha1.Site{}
	case "Cluster":
		target = &inventoryv1alpha1.Cluster{}
	case "NetworkDevice":
		target = &inventoryv1alpha1.NetworkDevice{}
	default:
		return apierrors.NewBadRequest(fmt.Sprintf("endpoint kind %q is not allowed", ep.Kind))
	}

	if err := v.Client.Get(ctx, types.NamespacedName{Name: ep.Name}, target); err != nil {
		if apierrors.IsNotFound(err) {
			return apierrors.NewBadRequest(fmt.Sprintf("endpoint %s/%s not found", ep.Kind, ep.Name))
		}
		return err
	}
	return nil
}

// SetupLinkWebhookWithManager registers the LinkValidator with the manager.
func SetupLinkWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Link{}).
		WithValidator(&LinkValidator{Client: mgr.GetClient()}).
		Complete()
}
