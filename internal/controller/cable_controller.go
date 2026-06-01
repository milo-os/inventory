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

// CableReconciler reconciles Cable objects. It resolves both endpoint Ports
// and marks the Cable Ready/EndpointsResolved.
type CableReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=cables,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=cables/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=ports,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *CableReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	cable := &inventoryv1alpha1.Cable{}
	if err := r.Get(ctx, req.NamespacedName, cable); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := cable.DeepCopy()

	for _, ep := range cable.Spec.Endpoints {
		port := &inventoryv1alpha1.Port{}
		err := r.Get(ctx, client.ObjectKey{Name: ep.Name}, port)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				cable.GetGeneration(),
				&cable.Status.Conditions,
				inventoryv1alpha1.CableEndpointNotFoundReason,
				fmt.Sprintf("Port %q not found", ep.Name),
			)
			meta.SetStatusCondition(&cable.Status.Conditions, metav1.Condition{
				Type:               inventoryv1alpha1.CableEndpointsResolved,
				Status:             metav1.ConditionFalse,
				Reason:             inventoryv1alpha1.CableEndpointNotFoundReason,
				Message:            fmt.Sprintf("Port %q not found", ep.Name),
				ObservedGeneration: cable.GetGeneration(),
			})
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, cable); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to get endpoint Port", "port", ep.Name)
			return ctrl.Result{}, err
		}
	}

	SetReady(cable.GetGeneration(), &cable.Status.Conditions, inventoryv1alpha1.CableReadyReason, "Cable accepted")
	meta.SetStatusCondition(&cable.Status.Conditions, metav1.Condition{
		Type:               inventoryv1alpha1.CableEndpointsResolved,
		Status:             metav1.ConditionTrue,
		Reason:             inventoryv1alpha1.CableReadyReason,
		Message:            "All endpoints resolved",
		ObservedGeneration: cable.GetGeneration(),
	})

	if err := r.patchStatusIfChanged(ctx, originalStatus, cable); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *CableReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Cable) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	return r.Status().Patch(ctx, current, client.MergeFrom(original))
}

func (r *CableReconciler) enqueueForPort(ctx context.Context, obj client.Object) []reconcile.Request {
	port, ok := obj.(*inventoryv1alpha1.Port)
	if !ok {
		return nil
	}
	cables := &inventoryv1alpha1.CableList{}
	if err := r.List(ctx, cables, client.MatchingFields{IndexCableEndpointName: port.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing Cables for Port", "port", port.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(cables.Items))
	for _, c := range cables.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: c.Name}})
	}
	return reqs
}

// SetupWithManager registers the Cable controller with the manager.
func (r *CableReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Cable{}).
		Watches(&inventoryv1alpha1.Port{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForPort)).
		Named("cable").
		Complete(r)
}
