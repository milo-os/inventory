// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"reflect"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// RegionReconciler reconciles Region objects.
//
// Regions are the top of the inventory hierarchy and have no cross-resource
// references, so the reconciler does nothing beyond marking them Ready so
// that controllers watching Region status (and kubectl Ready columns)
// reflect acceptance.
type RegionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=regions,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=regions/status,verbs=get;update;patch

// Reconcile implements reconcile.Reconciler.
func (r *RegionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	region := &inventoryv1alpha1.Region{}
	if err := r.Get(ctx, req.NamespacedName, region); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	original := region.DeepCopy()

	SetReady(region.GetGeneration(), &region.Status.Conditions, inventoryv1alpha1.RegionReadyReason, "Region accepted")

	if reflect.DeepEqual(original.Status, region.Status) {
		return ctrl.Result{}, nil
	}

	if err := r.Status().Patch(ctx, region, client.MergeFrom(original)); err != nil {
		log.Error(err, "failed to patch Region status")
		return ctrl.Result{}, err
	}

	r.EventRecorder.EmitReadyTransition(ctx, region,
		inventoryv1alpha1.GroupVersion.WithKind("Region"),
		displayNameOrName(region.Spec.DisplayName, region.Name),
		original.Status.Conditions, region.Status.Conditions)

	return ctrl.Result{}, nil
}

// SetupWithManager registers the Region controller with the supplied
// manager. SetupIndexers must have already been called.
func (r *RegionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Region{}).
		Named("region").
		Complete(r)
}
