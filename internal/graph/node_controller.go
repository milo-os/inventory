// SPDX-License-Identifier: AGPL-3.0-only

package graph

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"
	"go.miloapis.com/inventory/internal/controller"
)

// NodeReconciler reconciles graph Node objects. It re-validates the Node
// against its NodeType and reflects the result in the Ready condition. The
// admission webhook is the primary gate; this controller keeps status honest
// when a NodeType changes after the Node was admitted.
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=graph.inventory.miloapis.com,resources=nodetypes,verbs=get;list;watch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	node := &inventoryv1alpha2.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	original := node.DeepCopy()

	if err := ValidateNode(ctx, r.Client, node); err != nil {
		controller.SetNotReady(node.GetGeneration(), &node.Status.Conditions,
			inventoryv1alpha2.NodeTypeNotFoundReason, err.Error())
	} else {
		controller.SetReady(node.GetGeneration(), &node.Status.Conditions,
			inventoryv1alpha2.NodeReadyReason, "Node accepted")
	}

	if reflect.DeepEqual(original.Status, node.Status) {
		return ctrl.Result{}, nil
	}
	if err := r.Status().Patch(ctx, node, client.MergeFrom(original)); err != nil {
		log.Error(err, "failed to patch Node status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the Node controller. SetupIndexers must have
// already been called.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha2.Node{}).
		Watches(&inventoryv1alpha2.NodeType{}, handler.EnqueueRequestsFromMapFunc(r.nodesForType)).
		Named("graph-node").
		Complete(r)
}

// nodesForType enqueues every Node whose spec.type matches the changed
// NodeType so their Ready condition tracks schema edits.
func (r *NodeReconciler) nodesForType(ctx context.Context, obj client.Object) []reconcile.Request {
	var nodes inventoryv1alpha2.NodeList
	if err := r.List(ctx, &nodes); err != nil {
		return nil
	}
	var reqs []reconcile.Request
	for i := range nodes.Items {
		if nodes.Items[i].Spec.Type == obj.GetName() {
			reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&nodes.Items[i])})
		}
	}
	return reqs
}
