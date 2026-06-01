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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-circuit,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=circuits,verbs=create;update,versions=v1alpha1,name=vcircuit.inventory.miloapis.com,admissionReviewVersions=v1

var circuitLog = logf.Log.WithName("circuit-webhook")

// CircuitValidator validates Circuit CREATE/UPDATE operations. It verifies the
// referenced Provider exists and that every Site-kind endpoint exists. CEL on
// the type handles enums and providerRef immutability. Port-kind endpoints are
// not existence-checked yet (deferred until the Port kind ships).
type CircuitValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Circuit] = &CircuitValidator{}

func (v *CircuitValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Circuit) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

func (v *CircuitValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Circuit) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

func (v *CircuitValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha1.Circuit) (admission.Warnings, error) {
	return nil, nil
}

func (v *CircuitValidator) validate(ctx context.Context, circuit *inventoryv1alpha1.Circuit) (admission.Warnings, error) {
	circuitLog.Info("validating", "name", circuit.Name)

	var provider inventoryv1alpha1.Provider
	if err := v.Client.Get(ctx, types.NamespacedName{Name: circuit.Spec.ProviderRef.Name}, &provider); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, apierrors.NewBadRequest(fmt.Sprintf("referenced Provider %s not found", circuit.Spec.ProviderRef.Name))
		}
		return nil, err
	}

	for _, ep := range []inventoryv1alpha1.CircuitEndpoint{circuit.Spec.AEnd, circuit.Spec.ZEnd} {
		if ep.Kind != inventoryv1alpha1.CircuitEndpointKindSite {
			continue
		}
		var site inventoryv1alpha1.Site
		if err := v.Client.Get(ctx, types.NamespacedName{Name: ep.Name}, &site); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, apierrors.NewBadRequest(fmt.Sprintf("endpoint Site %s not found", ep.Name))
			}
			return nil, err
		}
	}
	return nil, nil
}

// SetupCircuitWebhookWithManager registers the CircuitValidator with the
// manager.
func SetupCircuitWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Circuit{}).
		WithValidator(&CircuitValidator{Client: mgr.GetClient()}).
		Complete()
}
