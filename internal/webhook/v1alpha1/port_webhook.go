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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-port,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=ports,verbs=delete,versions=v1alpha1,name=vport.inventory.miloapis.com,admissionReviewVersions=v1

var portLog = logf.Log.WithName("port-webhook")

// PortValidator rejects deletion of a Port while any Cable references it as
// an endpoint. CREATE/UPDATE are no-ops — CEL on the type handles deviceRef
// immutability and enums.
type PortValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Port] = &PortValidator{}

func (v *PortValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Port) (admission.Warnings, error) {
	return nil, nil
}

func (v *PortValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Port) (admission.Warnings, error) {
	return nil, nil
}

func (v *PortValidator) ValidateDelete(ctx context.Context, port *inventoryv1alpha1.Port) (admission.Warnings, error) {
	portLog.Info("validating delete", "name", port.Name)

	var cables inventoryv1alpha1.CableList
	if err := v.Client.List(ctx, &cables, client.MatchingFields{controller.IndexCableEndpointName: port.Name}); err != nil {
		return nil, err
	}
	if len(cables.Items) == 0 {
		return nil, nil
	}

	names := childNames(cables.Items, func(c inventoryv1alpha1.Cable) string { return c.Name })
	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Port %s: %d Cable(s) still reference it: %v%s",
		port.Name, len(cables.Items), names, truncationSuffix(len(cables.Items)),
	))
}

// SetupPortWebhookWithManager registers the PortValidator with the manager.
func SetupPortWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Port{}).
		WithValidator(&PortValidator{Client: mgr.GetClient()}).
		Complete()
}
