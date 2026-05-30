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

type ProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=providers,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=providers/status,verbs=get;update;patch

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	provider := &inventoryv1alpha1.Provider{}
	if err := r.Get(ctx, req.NamespacedName, provider); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	original := provider.DeepCopy()

	SetReady(provider.GetGeneration(), &provider.Status.Conditions, inventoryv1alpha1.ProviderReadyReason, "Provider accepted")

	if reflect.DeepEqual(original.Status, provider.Status) {
		return ctrl.Result{}, nil
	}

	if err := r.Status().Patch(ctx, provider, client.MergeFrom(original)); err != nil {
		log.Error(err, "failed to patch Provider status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Provider{}).
		Named("provider").
		Complete(r)
}
