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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-networkdevice,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=networkdevices,verbs=create;update,versions=v1alpha1,name=vnetworkdevice.inventory.miloapis.com,admissionReviewVersions=v1

var networkDeviceLog = logf.Log.WithName("networkdevice-webhook")

// NetworkDeviceValidator validates NetworkDevice CREATE/UPDATE operations.
// It verifies the referenced Cluster exists. The device's SiteRef is
// independent of the Cluster's control-plane Site: a cluster can span
// multiple sites, so cross-site attachment is a valid topology.
type NetworkDeviceValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.NetworkDevice] = &NetworkDeviceValidator{}

func (v *NetworkDeviceValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.NetworkDevice) (admission.Warnings, error) {
	return v.validate(ctx, obj)
}

func (v *NetworkDeviceValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.NetworkDevice) (admission.Warnings, error) {
	return v.validate(ctx, newObj)
}

func (v *NetworkDeviceValidator) ValidateDelete(ctx context.Context, obj *inventoryv1alpha1.NetworkDevice) (admission.Warnings, error) {
	return nil, nil
}

// validate checks that the referenced Cluster exists. It does NOT enforce
// that the device and cluster share a site — clusters can span sites.
func (v *NetworkDeviceValidator) validate(ctx context.Context, device *inventoryv1alpha1.NetworkDevice) (admission.Warnings, error) {
	networkDeviceLog.Info("validating", "name", device.Name)

	var cluster inventoryv1alpha1.Cluster
	if err := v.Client.Get(ctx, types.NamespacedName{Name: device.Spec.ClusterRef.Name}, &cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, apierrors.NewBadRequest(fmt.Sprintf("referenced Cluster %s not found", device.Spec.ClusterRef.Name))
		}
		return nil, err
	}

	return nil, nil
}

// SetupNetworkDeviceWebhookWithManager registers the NetworkDeviceValidator
// with the manager.
func SetupNetworkDeviceWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.NetworkDevice{}).
		WithValidator(&NetworkDeviceValidator{Client: mgr.GetClient()}).
		Complete()
}
