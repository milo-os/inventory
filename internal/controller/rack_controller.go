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

// RackReconciler reconciles Rack objects. It validates the Rack's SiteRef and
// propagates topology labels (region, site, site-type) from the parent Site.
type RackReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=racks,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=racks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *RackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	rack := &inventoryv1alpha1.Rack{}
	if err := r.Get(ctx, req.NamespacedName, rack); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := rack.DeepCopy()

	site := &inventoryv1alpha1.Site{}
	err := r.Get(ctx, client.ObjectKey{Name: rack.Spec.SiteRef.Name}, site)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			rack.GetGeneration(),
			&rack.Status.Conditions,
			inventoryv1alpha1.RackSiteNotFoundReason,
			fmt.Sprintf("Site %q not found", rack.Spec.SiteRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, rack); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Site", "site", rack.Spec.SiteRef.Name)
		return ctrl.Result{}, err
	}

	want := siteLabels(site)

	originalSpec := rack.DeepCopy()
	if ensureLabels(rack, want) {
		if err := r.Patch(ctx, rack, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Rack labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(rack.GetGeneration(), &rack.Status.Conditions, inventoryv1alpha1.RackReadyReason, "Rack accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, rack); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RackReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Rack) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	if err := r.Status().Patch(ctx, current, client.MergeFrom(original)); err != nil {
		return err
	}
	r.EventRecorder.EmitReadyTransition(ctx, current,
		inventoryv1alpha1.GroupVersion.WithKind("Rack"),
		current.Name,
		original.Status.Conditions, current.Status.Conditions)
	return nil
}

func (r *RackReconciler) enqueueForSite(ctx context.Context, obj client.Object) []reconcile.Request {
	site, ok := obj.(*inventoryv1alpha1.Site)
	if !ok {
		return nil
	}
	racks := &inventoryv1alpha1.RackList{}
	if err := r.List(ctx, racks, client.MatchingFields{IndexRackSiteRef: site.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Racks for Site", "site", site.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(racks.Items))
	for _, rk := range racks.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: rk.Name}})
	}
	return reqs
}

// SetupWithManager registers the Rack controller with the manager.
func (r *RackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Rack{}).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForSite)).
		Named("rack").
		Complete(r)
}
