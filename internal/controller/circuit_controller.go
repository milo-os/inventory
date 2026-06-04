// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"fmt"
	"reflect"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CircuitReconciler reconciles Circuit objects. It validates the providerRef
// and every Site-kind endpoint, and marks the Circuit Ready/EndpointsResolved.
// Port-kind endpoints are accepted but not existence-checked (deferred until
// the Port kind ships). The cross-group serviceRef is recorded but not
// resolved — its target may live on another cluster.
type CircuitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=circuits,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=circuits/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=providers,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *CircuitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	circuit := &inventoryv1alpha1.Circuit{}
	if err := r.Get(ctx, req.NamespacedName, circuit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := circuit.DeepCopy()

	provider := &inventoryv1alpha1.Provider{}
	err := r.Get(ctx, client.ObjectKey{Name: circuit.Spec.ProviderRef.Name}, provider)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			circuit.GetGeneration(),
			&circuit.Status.Conditions,
			inventoryv1alpha1.CircuitProviderNotFoundReason,
			fmt.Sprintf("Provider %q not found", circuit.Spec.ProviderRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, circuit); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Provider", "provider", circuit.Spec.ProviderRef.Name)
		return ctrl.Result{}, err
	}

	for _, ep := range []inventoryv1alpha1.CircuitEndpoint{circuit.Spec.AEnd, circuit.Spec.ZEnd} {
		if ep.Kind != inventoryv1alpha1.CircuitEndpointKindSite {
			continue
		}
		site := &inventoryv1alpha1.Site{}
		err := r.Get(ctx, client.ObjectKey{Name: ep.Name}, site)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				circuit.GetGeneration(),
				&circuit.Status.Conditions,
				inventoryv1alpha1.CircuitEndpointNotFoundReason,
				fmt.Sprintf("Site %q not found", ep.Name),
			)
			meta.SetStatusCondition(&circuit.Status.Conditions, metav1.Condition{
				Type:               inventoryv1alpha1.CircuitEndpointsResolved,
				Status:             metav1.ConditionFalse,
				Reason:             inventoryv1alpha1.CircuitEndpointNotFoundReason,
				Message:            fmt.Sprintf("Site %q not found", ep.Name),
				ObservedGeneration: circuit.GetGeneration(),
			})
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, circuit); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to get endpoint Site", "site", ep.Name)
			return ctrl.Result{}, err
		}
	}

	SetReady(circuit.GetGeneration(), &circuit.Status.Conditions, inventoryv1alpha1.CircuitReadyReason, "Circuit accepted")
	meta.SetStatusCondition(&circuit.Status.Conditions, metav1.Condition{
		Type:               inventoryv1alpha1.CircuitEndpointsResolved,
		Status:             metav1.ConditionTrue,
		Reason:             inventoryv1alpha1.CircuitReadyReason,
		Message:            "All Site endpoints resolved",
		ObservedGeneration: circuit.GetGeneration(),
	})

	if err := r.patchStatusIfChanged(ctx, originalStatus, circuit); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *CircuitReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Circuit) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	if err := r.Status().Patch(ctx, current, client.MergeFrom(original)); err != nil {
		return err
	}
	r.EventRecorder.EmitReadyTransition(ctx, current,
		inventoryv1alpha1.GroupVersion.WithKind("Circuit"),
		current.Name,
		original.Status.Conditions, current.Status.Conditions)
	return nil
}

func (r *CircuitReconciler) enqueueForProvider(ctx context.Context, obj client.Object) []reconcile.Request {
	provider, ok := obj.(*inventoryv1alpha1.Provider)
	if !ok {
		return nil
	}
	circuits := &inventoryv1alpha1.CircuitList{}
	if err := r.List(ctx, circuits, client.MatchingFields{IndexCircuitProviderRef: provider.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Circuits for Provider", "provider", provider.Name)
		return nil
	}
	return circuitRequests(circuits)
}

func (r *CircuitReconciler) enqueueForSite(ctx context.Context, obj client.Object) []reconcile.Request {
	site, ok := obj.(*inventoryv1alpha1.Site)
	if !ok {
		return nil
	}
	circuits := &inventoryv1alpha1.CircuitList{}
	if err := r.List(ctx, circuits, client.MatchingFields{IndexCircuitSiteEndpoint: site.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Circuits for Site", "site", site.Name)
		return nil
	}
	return circuitRequests(circuits)
}

func circuitRequests(circuits *inventoryv1alpha1.CircuitList) []reconcile.Request {
	reqs := make([]reconcile.Request, 0, len(circuits.Items))
	for _, c := range circuits.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: c.Name}})
	}
	return reqs
}

// SetupWithManager registers the Circuit controller with the manager.
func (r *CircuitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Circuit{}).
		Watches(&inventoryv1alpha1.Provider{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForProvider)).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForSite)).
		Named("circuit").
		Complete(r)
}
