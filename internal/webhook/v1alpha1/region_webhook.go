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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-region,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=regions,verbs=delete,versions=v1alpha1,name=vregion.inventory.miloapis.com,admissionReviewVersions=v1

var regionLog = logf.Log.WithName("region-webhook")

// RegionValidator validates Region DELETE operations. It rejects deletion
// whenever any Site references the Region via `.spec.regionRef.name`.
type RegionValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Region] = &RegionValidator{}

// ValidateCreate is a no-op. CEL on the type handles structural validation.
func (v *RegionValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Region) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate is a no-op. CEL on the type handles structural/immutability
// validation.
func (v *RegionValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Region) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete rejects the deletion when any Site still references this
// Region.
func (v *RegionValidator) ValidateDelete(ctx context.Context, region *inventoryv1alpha1.Region) (admission.Warnings, error) {
	regionLog.Info("validating delete", "name", region.Name)

	var sites inventoryv1alpha1.SiteList
	if err := v.Client.List(ctx, &sites, client.MatchingFields{controller.IndexSiteRegionRef: region.Name}); err != nil {
		return nil, err
	}
	if len(sites.Items) > 0 {
		names := childNames(sites.Items, func(s inventoryv1alpha1.Site) string { return s.Name })
		return nil, apierrors.NewBadRequest(fmt.Sprintf(
			"cannot delete Region %s: %d Site(s) still reference it: %v%s",
			region.Name, len(sites.Items), names, truncationSuffix(len(sites.Items)),
		))
	}
	return nil, nil
}

// SetupRegionWebhookWithManager registers the RegionValidator with the
// manager.
func SetupRegionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Region{}).
		WithValidator(&RegionValidator{Client: mgr.GetClient()}).
		Complete()
}
