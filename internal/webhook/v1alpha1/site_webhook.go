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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-site,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=sites,verbs=create;update;delete,versions=v1alpha1,name=vsite.inventory.miloapis.com,admissionReviewVersions=v1

var siteLog = logf.Log.WithName("site-webhook")

// SiteValidator validates Site operations. On CREATE/UPDATE it is a no-op
// (CEL on the type handles immutability and enums). On DELETE it rejects
// deletion when any Node, Cluster, or NetworkDevice references the Site.
type SiteValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Site] = &SiteValidator{}

func (v *SiteValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Site) (admission.Warnings, error) {
	return nil, nil
}

func (v *SiteValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Site) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete rejects deletion when any Node, Cluster, or NetworkDevice
// references this Site.
func (v *SiteValidator) ValidateDelete(ctx context.Context, site *inventoryv1alpha1.Site) (admission.Warnings, error) {
	siteLog.Info("validating delete", "name", site.Name)

	var nodes inventoryv1alpha1.NodeList
	if err := v.Client.List(ctx, &nodes, client.MatchingFields{controller.IndexNodeSiteRef: site.Name}); err != nil {
		return nil, err
	}
	var clusters inventoryv1alpha1.ClusterList
	if err := v.Client.List(ctx, &clusters, client.MatchingFields{controller.IndexClusterControlPlaneSiteRef: site.Name}); err != nil {
		return nil, err
	}
	var devices inventoryv1alpha1.NetworkDeviceList
	if err := v.Client.List(ctx, &devices, client.MatchingFields{controller.IndexNetworkDeviceSiteRef: site.Name}); err != nil {
		return nil, err
	}

	total := len(nodes.Items) + len(clusters.Items) + len(devices.Items)
	if total == 0 {
		return nil, nil
	}

	var parts []string
	if n := len(nodes.Items); n > 0 {
		names := childNames(nodes.Items, func(x inventoryv1alpha1.Node) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d Node(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(clusters.Items); n > 0 {
		names := childNames(clusters.Items, func(x inventoryv1alpha1.Cluster) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d Cluster(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(devices.Items); n > 0 {
		names := childNames(devices.Items, func(x inventoryv1alpha1.NetworkDevice) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d NetworkDevice(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}

	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Site %s: %s",
		site.Name, strings.Join(parts, "; "),
	))
}

// SetupSiteWebhookWithManager registers the SiteValidator with the manager.
func SetupSiteWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Site{}).
		WithValidator(&SiteValidator{Client: mgr.GetClient()}).
		Complete()
}
