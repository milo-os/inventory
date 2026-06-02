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

// NodeReconciler reconciles Node objects. It validates the Node's SiteRef
// and (if present) its Assignment.ClusterRef, sets the Phase accordingly,
// and propagates topology labels.
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=nodes,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=clusters,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	node := &inventoryv1alpha1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := node.DeepCopy()

	// Validate SiteRef — Site is required.
	site := &inventoryv1alpha1.Site{}
	err := r.Get(ctx, client.ObjectKey{Name: node.Spec.SiteRef.Name}, site)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			node.GetGeneration(),
			&node.Status.Conditions,
			inventoryv1alpha1.NodeSiteNotFoundReason,
			fmt.Sprintf("Site %q not found", node.Spec.SiteRef.Name),
		)
		node.Status.Phase = inventoryv1alpha1.NodePhaseUnavailable
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, node); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Site", "site", node.Spec.SiteRef.Name)
		return ctrl.Result{}, err
	}

	// Decide label set + phase based on whether the Node is assigned.
	var want map[string]string
	var phase inventoryv1alpha1.NodePhase

	if node.Spec.Assignment != nil {
		cluster := &inventoryv1alpha1.Cluster{}
		err := r.Get(ctx, client.ObjectKey{Name: node.Spec.Assignment.ClusterRef.Name}, cluster)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				node.GetGeneration(),
				&node.Status.Conditions,
				inventoryv1alpha1.NodeClusterNotFoundReason,
				fmt.Sprintf("Cluster %q not found", node.Spec.Assignment.ClusterRef.Name),
			)
			node.Status.Phase = inventoryv1alpha1.NodePhaseUnavailable
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, node); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to get referenced Cluster", "cluster", node.Spec.Assignment.ClusterRef.Name)
			return ctrl.Result{}, err
		}

		want = clusterLabels(cluster)
		phase = inventoryv1alpha1.NodePhaseAssigned
	} else {
		want = siteLabels(site)
		// Explicitly clear the cluster label when unassigned — an
		// empty-string value tells ensureLabels to delete the key.
		want[inventoryv1alpha1.TopologyClusterLabel] = ""
		phase = inventoryv1alpha1.NodePhaseUnassigned
	}

	want[inventoryv1alpha1.TopologyRackLabel] = rackLabel(node.Spec.Placement)

	originalSpec := node.DeepCopy()
	if ensureLabels(node, want) {
		if err := r.Patch(ctx, node, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Node labels")
			return ctrl.Result{}, err
		}
	}

	node.Status.Phase = phase
	SetReady(node.GetGeneration(), &node.Status.Conditions, inventoryv1alpha1.NodeReadyReason, "Node accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, node); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *NodeReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Node) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	return r.Status().Patch(ctx, current, client.MergeFrom(original))
}

func (r *NodeReconciler) enqueueForSite(ctx context.Context, obj client.Object) []reconcile.Request {
	site, ok := obj.(*inventoryv1alpha1.Site)
	if !ok {
		return nil
	}
	nodes := &inventoryv1alpha1.NodeList{}
	if err := r.List(ctx, nodes, client.MatchingFields{IndexNodeSiteRef: site.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Nodes for Site", "site", site.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(nodes.Items))
	for _, n := range nodes.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: n.Name}})
	}
	return reqs
}

func (r *NodeReconciler) enqueueForCluster(ctx context.Context, obj client.Object) []reconcile.Request {
	cluster, ok := obj.(*inventoryv1alpha1.Cluster)
	if !ok {
		return nil
	}
	nodes := &inventoryv1alpha1.NodeList{}
	if err := r.List(ctx, nodes, client.MatchingFields{IndexNodeAssignmentClusterRef: cluster.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Nodes for Cluster", "cluster", cluster.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(nodes.Items))
	for _, n := range nodes.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: n.Name}})
	}
	return reqs
}

// SetupWithManager registers the Node controller with the manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Node{}).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForSite)).
		Watches(&inventoryv1alpha1.Cluster{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForCluster)).
		Named("node").
		Complete(r)
}
