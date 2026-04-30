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

// +kubebuilder:webhook:path=/validate-inventory-miloapis-com-v1alpha1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=inventory.miloapis.com,resources=clusters,verbs=delete,versions=v1alpha1,name=vcluster.inventory.miloapis.com,admissionReviewVersions=v1

var clusterLog = logf.Log.WithName("cluster-webhook")

// ClusterValidator validates Cluster DELETE operations. It rejects deletion
// when any Node (via `.spec.assignment.clusterRef`) or NetworkDevice
// references the Cluster.
type ClusterValidator struct {
	Client client.Client
}

var _ admission.Validator[*inventoryv1alpha1.Cluster] = &ClusterValidator{}

func (v *ClusterValidator) ValidateCreate(ctx context.Context, obj *inventoryv1alpha1.Cluster) (admission.Warnings, error) {
	return nil, nil
}

func (v *ClusterValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *inventoryv1alpha1.Cluster) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete rejects the deletion when any Node (via its assignment) or
// NetworkDevice references the Cluster.
func (v *ClusterValidator) ValidateDelete(ctx context.Context, cluster *inventoryv1alpha1.Cluster) (admission.Warnings, error) {
	clusterLog.Info("validating delete", "name", cluster.Name)

	var nodes inventoryv1alpha1.NodeList
	if err := v.Client.List(ctx, &nodes, client.MatchingFields{controller.IndexNodeAssignmentClusterRef: cluster.Name}); err != nil {
		return nil, err
	}
	var devices inventoryv1alpha1.NetworkDeviceList
	if err := v.Client.List(ctx, &devices, client.MatchingFields{controller.IndexNetworkDeviceClusterRef: cluster.Name}); err != nil {
		return nil, err
	}

	total := len(nodes.Items) + len(devices.Items)
	if total == 0 {
		return nil, nil
	}

	var parts []string
	if n := len(nodes.Items); n > 0 {
		names := childNames(nodes.Items, func(x inventoryv1alpha1.Node) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d Node(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}
	if n := len(devices.Items); n > 0 {
		names := childNames(devices.Items, func(x inventoryv1alpha1.NetworkDevice) string { return x.Name })
		parts = append(parts, fmt.Sprintf("%d NetworkDevice(s) still reference it: %v%s", n, names, truncationSuffix(n)))
	}

	return nil, apierrors.NewBadRequest(fmt.Sprintf(
		"cannot delete Cluster %s: %s",
		cluster.Name, strings.Join(parts, "; "),
	))
}

// SetupClusterWebhookWithManager registers the ClusterValidator with the
// manager.
func SetupClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &inventoryv1alpha1.Cluster{}).
		WithValidator(&ClusterValidator{Client: mgr.GetClient()}).
		Complete()
}
