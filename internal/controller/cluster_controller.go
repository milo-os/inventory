// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"fmt"
	"reflect"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ClusterReconciler reconciles Cluster objects. It validates the
// ControlPlaneSiteRef and propagates topology labels (region, site,
// site-type, cluster) onto the Cluster itself. The propagated labels
// describe the control plane's location; worker Nodes and NetworkDevices
// inherit their own placement from their respective SiteRefs.
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=clusters,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	cluster := &inventoryv1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := cluster.DeepCopy()

	site := &inventoryv1alpha1.Site{}
	err := r.Get(ctx, client.ObjectKey{Name: cluster.Spec.ControlPlaneSiteRef.Name}, site)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			cluster.GetGeneration(),
			&cluster.Status.Conditions,
			inventoryv1alpha1.ClusterSiteNotFoundReason,
			fmt.Sprintf("Site %q not found", cluster.Spec.ControlPlaneSiteRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, cluster); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Site", "site", cluster.Spec.ControlPlaneSiteRef.Name)
		return ctrl.Result{}, err
	}

	// Compose labels: everything the Site provides plus this cluster's own name.
	want := siteLabels(site)
	want[inventoryv1alpha1.TopologyClusterLabel] = cluster.Name

	originalSpec := cluster.DeepCopy()
	if ensureLabels(cluster, want) {
		if err := r.Patch(ctx, cluster, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Cluster labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(cluster.GetGeneration(), &cluster.Status.Conditions, inventoryv1alpha1.ClusterReadyReason, "Cluster accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, cluster); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Cluster) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	return r.Status().Patch(ctx, current, client.MergeFrom(original))
}

// enqueueForSite returns the Clusters whose spec.controlPlaneSiteRef.name
// matches the given Site's name.
func (r *ClusterReconciler) enqueueForSite(ctx context.Context, obj client.Object) []reconcile.Request {
	site, ok := obj.(*inventoryv1alpha1.Site)
	if !ok {
		return nil
	}
	clusters := &inventoryv1alpha1.ClusterList{}
	if err := r.List(ctx, clusters, client.MatchingFields{IndexClusterControlPlaneSiteRef: site.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Clusters for Site", "site", site.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(clusters.Items))
	for _, c := range clusters.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: c.Name}})
	}
	return reqs
}

// SetupWithManager registers the Cluster controller with the manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Cluster{}).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForSite)).
		Named("cluster").
		Complete(r)
}
