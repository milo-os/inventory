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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=providers,verbs=delete,versions=v1alpha1,name=vprovider.inventory.miloapis.com,admissionReviewVersions=v1

var providerLog = logf.Log.WithName("provider-webhook")

// ProviderValidator validates Provider DELETE operations. It rejects
// deletion whenever any Site references the Provider via
// `.spec.providerRef.name`.
type ProviderValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Provider] = &ProviderValidator{}

// ValidateCreate is a no-op. CEL on the type handles structural validation.
func (v *ProviderValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate is a no-op. CEL on the type handles structural validation.
func (v *ProviderValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete rejects the deletion when any Site still references this
// Provider.
func (v *ProviderValidator) ValidateDelete(ctx context.Context, provider *inventoryv1alpha1.Provider) (admission.Warnings, error) {
	providerLog.Info("validating delete", "name", provider.Name)

	var sites inventoryv1alpha1.SiteList
	if err := v.Client.List(ctx, &sites, client.MatchingFields{controller.IndexSiteProviderRef: provider.Name}); err != nil {
		return nil, err
	}
	if len(sites.Items) > 0 {
		names := childNames(sites.Items, func(s inventoryv1alpha1.Site) string { return s.Name })
		return nil, apierrors.NewBadRequest(fmt.Sprintf(
			"cannot delete Provider %s: %d Site(s) still reference it: %v%s",
			provider.Name, len(sites.Items), names, truncationSuffix(len(sites.Items)),
		))
	}
	return nil, nil
}

// SetupProviderWebhookWithManager registers the ProviderValidator with the
// manager.
func SetupProviderWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Provider{}).
		WithValidator(&ProviderValidator{Client: mgr.GetClient()}).
		Complete()
}
