// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

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

// requeueAfterMissingRef is the gap between reconciliations when a
// reference is dangling. The parent-object watch takes care of the
// happy-path wake-up as soon as the reference appears; the periodic
// requeue is a backstop in case an event is missed.
const requeueAfterMissingRef = 30 * time.Second

// SiteReconciler reconciles Site objects. It validates the Site's
// RegionRef and propagates topology labels (region, site, site-type) onto
// itself so that child kinds (Cluster, Node, NetworkDevice) can copy them
// in turn.
type SiteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=regions,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=providers,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *SiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	site := &inventoryv1alpha1.Site{}
	if err := r.Get(ctx, req.NamespacedName, site); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := site.DeepCopy()

	region := &inventoryv1alpha1.Region{}
	err := r.Get(ctx, client.ObjectKey{Name: site.Spec.RegionRef.Name}, region)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			site.GetGeneration(),
			&site.Status.Conditions,
			inventoryv1alpha1.SiteRegionNotFoundReason,
			fmt.Sprintf("Region %q not found", site.Spec.RegionRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, site); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Region", "region", site.Spec.RegionRef.Name)
		return ctrl.Result{}, err
	}

	if site.Spec.ProviderRef != nil {
		provider := &inventoryv1alpha1.Provider{}
		err := r.Get(ctx, client.ObjectKey{Name: site.Spec.ProviderRef.Name}, provider)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				site.GetGeneration(),
				&site.Status.Conditions,
				inventoryv1alpha1.SiteProviderNotFoundReason,
				fmt.Sprintf("Provider %q not found", site.Spec.ProviderRef.Name),
			)
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, site); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to get referenced Provider", "provider", site.Spec.ProviderRef.Name)
			return ctrl.Result{}, err
		}
	}

	// Region exists — propagate labels onto the Site itself.
	want := map[string]string{
		inventoryv1alpha1.TopologyRegionLabel:   site.Spec.RegionRef.Name,
		inventoryv1alpha1.TopologySiteLabel:     site.Name,
		inventoryv1alpha1.TopologySiteTypeLabel: string(site.Spec.Type),
	}

	originalSpec := site.DeepCopy()
	if ensureLabels(site, want) {
		if err := r.Patch(ctx, site, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Site labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(site.GetGeneration(), &site.Status.Conditions, inventoryv1alpha1.SiteReadyReason, "Site accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, site); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SiteReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Site) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	if err := r.Status().Patch(ctx, current, client.MergeFrom(original)); err != nil {
		return err
	}
	r.EventRecorder.EmitReadyTransition(ctx, current,
		inventoryv1alpha1.GroupVersion.WithKind("Site"),
		displayNameOrName(current.Spec.DisplayName, current.Name),
		original.Status.Conditions, current.Status.Conditions)
	return nil
}

// enqueueForRegion returns the Sites whose spec.regionRef.name matches the
// given Region's name. Called on Region Create/Update/Delete events.
func (r *SiteReconciler) enqueueForRegion(ctx context.Context, obj client.Object) []reconcile.Request {
	region, ok := obj.(*inventoryv1alpha1.Region)
	if !ok {
		return nil
	}
	sites := &inventoryv1alpha1.SiteList{}
	if err := r.List(ctx, sites, client.MatchingFields{IndexSiteRegionRef: region.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Sites for Region", "region", region.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(sites.Items))
	for _, s := range sites.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: s.Name}})
	}
	return reqs
}

func (r *SiteReconciler) enqueueForProvider(ctx context.Context, obj client.Object) []reconcile.Request {
	provider, ok := obj.(*inventoryv1alpha1.Provider)
	if !ok {
		return nil
	}
	sites := &inventoryv1alpha1.SiteList{}
	if err := r.List(ctx, sites, client.MatchingFields{IndexSiteProviderRef: provider.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Sites for Provider", "provider", provider.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(sites.Items))
	for _, s := range sites.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: s.Name}})
	}
	return reqs
}

// SetupWithManager registers the Site controller with the supplied manager.
func (r *SiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Site{}).
		Watches(&inventoryv1alpha1.Region{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForRegion)).
		Watches(&inventoryv1alpha1.Provider{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForProvider)).
		Named("site").
		Complete(r)
}
