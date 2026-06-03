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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=providers,verbs=delete,versions=v1alpha1,name=vprovider.inventory.miloapis.com,admissionReviewVersions=v1

var providerLog = logf.Log.WithName("provider-webhook")

type ProviderValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Provider] = &ProviderValidator{}

func (v *ProviderValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	return nil, nil
}

func (v *ProviderValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	return nil, nil
}

func (v *ProviderValidator) ValidateDelete(ctx context.Context, provider *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	providerLog.Info("validating delete", "name", provider.Name)

	var sites inventoryv1alpha1.SiteList
	if err := v.Client.List(ctx, &sites, client.MatchingFields{controller.IndexSiteProviderRef: provider.Name}); err != nil {
		return nil, err
	}
	var circuits inventoryv1alpha1.CircuitList
	if err := v.Client.List(ctx, &circuits, client.MatchingFields{controller.IndexCircuitProviderRef: provider.Name}); err != nil {
		return nil, err
	}

	var vms inventoryv1alpha1.VirtualMachineList
	if err := v.Client.List(ctx, &vms, client.MatchingFields{controller.IndexVirtualMachineProviderRef: provider.Name}); err != nil {
		return nil, err
	}

	if len(sites.Items)+len(circuits.Items)+len(vms.Items) == 0 {
		return nil, nil
	}

	var parts []string
	if n := len(sites.Items); n > 0 {
		names := childNames(sites.Items, func(s inventoryv1alpha1.Site) string { return s.Name })
		parts = append(parts, fmt.Sprintf("%d Site(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(circuits.Items); n > 0 {
		names := childNames(circuits.Items, func(c inventoryv1alpha1.Circuit) string { return c.Name })
		parts = append(parts, fmt.Sprintf("%d Circuit(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(vms.Items); n > 0 {
		names := childNames(vms.Items, func(x inventoryv1alpha1.VirtualMachine) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d VirtualMachine(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}

	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Provider %s: %s",
		provider.Name, strings.Join(parts, "; "),
	))
}

func SetupProviderWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Provider{}).
		WithValidator(&ProviderValidator{Client: mgr.GetClient()}).
		Complete()
}
