// SPDX-License-Identifier: AGPL-3.0-only

package graph

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
	"go.miloapis.com/inventory/internal/controller"
)

// EdgeReconciler reconciles graph Edge objects. It re-validates the Edge
// (EdgeType, endpoint existence, endpoint-type constraints, attributes) and
// reflects the result in the Ready and EndpointsResolved conditions.
type EdgeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=edges,verbs=get;list;watch
// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=edges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=edgetypes,verbs=get;list;watch

func (r *EdgeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	edge := &inventoryv1alpha2.Edge{}
	if err := r.Get(ctx, req.NamespacedName, edge); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	original := edge.DeepCopy()

	if err := ValidateEdge(ctx, r.Client, edge); err != nil {
		controller.SetNotReady(edge.GetGeneration(), &edge.Status.Conditions,
			inventoryv1alpha2.EdgeEndpointNotFoundReason, err.Error())
		r.setEndpointsResolved(edge, metav1.ConditionFalse, err.Error())
	} else {
		controller.SetReady(edge.GetGeneration(), &edge.Status.Conditions,
			inventoryv1alpha2.EdgeReadyReason, "Edge accepted")
		r.setEndpointsResolved(edge, metav1.ConditionTrue, "Both endpoints resolve")
	}

	if reflect.DeepEqual(original.Status, edge.Status) {
		return ctrl.Result{}, nil
	}
	if err := r.Status().Patch(ctx, edge, client.MergeFrom(original)); err != nil {
		log.Error(err, "failed to patch Edge status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *EdgeReconciler) setEndpointsResolved(edge *inventoryv1alpha2.Edge, status metav1.ConditionStatus, msg string) {
	reason := inventoryv1alpha2.EdgeReadyReason
	if status == metav1.ConditionFalse {
		reason = inventoryv1alpha2.EdgeEndpointNotFoundReason
	}
	meta.SetStatusCondition(&edge.Status.Conditions, metav1.Condition{
		Type:               inventoryv1alpha2.EdgeEndpointsResolved,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: edge.GetGeneration(),
	})
}

// SetupWithManager registers the Edge controller. SetupIndexers must have
// already been called.
func (r *EdgeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha2.Edge{}).
		Watches(&inventoryv1alpha2.Node{}, handler.EnqueueRequestsFromMapFunc(r.edgesForNode)).
		Named("graph-edge").
		Complete(r)
}

// edgesForNode enqueues every Edge that has the changed Node as an endpoint
// so endpoint-existence status stays current.
func (r *EdgeReconciler) edgesForNode(ctx context.Context, obj client.Object) []reconcile.Request {
	edges, err := EdgesReferencing(ctx, r.Client, obj.GetName())
	if err != nil {
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(edges))
	for i := range edges {
		reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&edges[i])})
	}
	return reqs
}
