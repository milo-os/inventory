// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"go.miloapis.com/inventory/internal/controller"
)

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-rack,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=racks,verbs=delete,versions=v1alpha1,name=vrack.inventory.miloapis.com,admissionReviewVersions=v1

var rackLog = logf.Log.WithName("rack-webhook")

// RackValidator rejects deletion of a Rack while any Node or NetworkDevice is
// placed in it. CREATE/UPDATE are no-ops — CEL on the type handles heightU and
// siteRef immutability.
type RackValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Rack] = &RackValidator{}

func (v *RackValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Rack) (admission.Warnings, error) {
	return nil, nil
}

func (v *RackValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Rack) (admission.Warnings, error) {
	return nil, nil
}

func (v *RackValidator) ValidateDelete(ctx context.Context, rack *inventoryv1alpha1.Rack) (admission.Warnings, error) {
	rackLog.Info("validating delete", "name", rack.Name)

	var nodes inventoryv1alpha1.NodeList
	if err := v.Client.List(ctx, &nodes, client.MatchingFields{controller.IndexNodePlacementRackRef: rack.Name}); err != nil {
		return nil, err
	}
	var devices inventoryv1alpha1.NetworkDeviceList
	if err := v.Client.List(ctx, &devices, client.MatchingFields{controller.IndexNetworkDevicePlacementRackRef: rack.Name}); err != nil {
		return nil, err
	}

	total := len(nodes.Items) + len(devices.Items)
	if total == 0 {
		return nil, nil
	}

	var parts []string
	if n := len(nodes.Items); n > 0 {
		names := childNames(nodes.Items, func(x inventoryv1alpha1.Node) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d Node(s) still placed in it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(devices.Items); n > 0 {
		names := childNames(devices.Items, func(x inventoryv1alpha1.NetworkDevice) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d NetworkDevice(s) still placed in it: %v%s", n, names, truncationSuffix(n)))
	}

	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Rack %s: %s",
		rack.Name, strings.Join(parts, "; "),
	))
}

// SetupRackWebhookWithManager registers the RackValidator with the manager.
func SetupRackWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Rack{}).
		WithValidator(&RackValidator{Client: mgr.GetClient()}).
		Complete()
}
