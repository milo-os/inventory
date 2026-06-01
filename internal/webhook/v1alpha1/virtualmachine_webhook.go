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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-virtualmachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=virtualmachines,verbs=create;update,versions=v1alpha1,name=vvirtualmachine.inventory.miloapis.com,admissionReviewVersions=v1

var virtualMachineLog = logf.Log.WithName("virtualmachine-webhook")

// VirtualMachineValidator validates VirtualMachine CREATE/UPDATE operations.
// It verifies the host Node exists and, when set, the referenced Provider.
// CEL on the type handles hostRef immutability and allocation bounds.
type VirtualMachineValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.VirtualMachine] = &VirtualMachineValidator{}

func (v *VirtualMachineValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.VirtualMachine) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

func (v *VirtualMachineValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.VirtualMachine) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

func (v *VirtualMachineValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha1.VirtualMachine) (admission.Warnings, error) {
	return nil, nil
}

func (v *VirtualMachineValidator) validate(ctx context.Context, vm *inventoryv1alpha1.VirtualMachine) (admission.Warnings, error) {
	virtualMachineLog.Info("validating", "name", vm.Name)

	var host inventoryv1alpha1.Node
	if err := v.Client.Get(ctx, types.NamespacedName{Name: vm.Spec.HostRef.Name}, &host); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, apierrors.NewBadRequest(fmt.Sprintf("host Node %s not found", vm.Spec.HostRef.Name))
		}
		return nil, err
	}

	if vm.Spec.ProviderRef != nil {
		var provider inventoryv1alpha1.Provider
		if err := v.Client.Get(ctx, types.NamespacedName{Name: vm.Spec.ProviderRef.Name}, &provider); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, apierrors.NewBadRequest(fmt.Sprintf("referenced Provider %s not found", vm.Spec.ProviderRef.Name))
			}
			return nil, err
		}
	}
	return nil, nil
}

// SetupVirtualMachineWebhookWithManager registers the VirtualMachineValidator
// with the manager.
func SetupVirtualMachineWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.VirtualMachine{}).
		WithValidator(&VirtualMachineValidator{Client: mgr.GetClient()}).
		Complete()
}
