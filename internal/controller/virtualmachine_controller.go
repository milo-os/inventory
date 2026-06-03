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

// VirtualMachineReconciler reconciles VirtualMachine objects. It validates the
// hostRef (Node) and optional providerRef, and propagates the host's topology
// labels onto the VM. The cross-group projectRef is recorded but not resolved.
type VirtualMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=virtualmachines,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=virtualmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=providers,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *VirtualMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	vm := &inventoryv1alpha1.VirtualMachine{}
	if err := r.Get(ctx, req.NamespacedName, vm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := vm.DeepCopy()

	host := &inventoryv1alpha1.Node{}
	err := r.Get(ctx, client.ObjectKey{Name: vm.Spec.HostRef.Name}, host)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			vm.GetGeneration(),
			&vm.Status.Conditions,
			inventoryv1alpha1.VirtualMachineHostNotFoundReason,
			fmt.Sprintf("Node %q not found", vm.Spec.HostRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, vm); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get host Node", "node", vm.Spec.HostRef.Name)
		return ctrl.Result{}, err
	}

	if vm.Spec.ProviderRef != nil {
		provider := &inventoryv1alpha1.Provider{}
		err := r.Get(ctx, client.ObjectKey{Name: vm.Spec.ProviderRef.Name}, provider)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				vm.GetGeneration(),
				&vm.Status.Conditions,
				inventoryv1alpha1.VirtualMachineProviderNotFoundReason,
				fmt.Sprintf("Provider %q not found", vm.Spec.ProviderRef.Name),
			)
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, vm); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to get referenced Provider", "provider", vm.Spec.ProviderRef.Name)
			return ctrl.Result{}, err
		}
	}

	// Inherit the host's topology labels.
	want := make(map[string]string, len(topologyLabelKeys))
	for _, k := range topologyLabelKeys {
		want[k] = host.Labels[k]
	}

	originalSpec := vm.DeepCopy()
	if ensureLabels(vm, want) {
		if err := r.Patch(ctx, vm, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch VirtualMachine labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(vm.GetGeneration(), &vm.Status.Conditions, inventoryv1alpha1.VirtualMachineReadyReason, "VirtualMachine accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, vm); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *VirtualMachineReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.VirtualMachine) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	if err := r.Status().Patch(ctx, current, client.MergeFrom(original)); err != nil {
		return err
	}
	r.EventRecorder.EmitReadyTransition(ctx, current,
		inventoryv1alpha1.GroupVersion.WithKind("VirtualMachine"),
		current.Name,
		original.Status.Conditions, current.Status.Conditions)
	return nil
}

func (r *VirtualMachineReconciler) enqueueForNode(ctx context.Context, obj client.Object) []reconcile.Request {
	node, ok := obj.(*inventoryv1alpha1.Node)
	if !ok {
		return nil
	}
	vms := &inventoryv1alpha1.VirtualMachineList{}
	if err := r.List(ctx, vms, client.MatchingFields{IndexVirtualMachineHostRef: node.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing VirtualMachines for Node", "node", node.Name)
		return nil
	}
	return vmRequests(vms)
}

func (r *VirtualMachineReconciler) enqueueForProvider(ctx context.Context, obj client.Object) []reconcile.Request {
	provider, ok := obj.(*inventoryv1alpha1.Provider)
	if !ok {
		return nil
	}
	vms := &inventoryv1alpha1.VirtualMachineList{}
	if err := r.List(ctx, vms, client.MatchingFields{IndexVirtualMachineProviderRef: provider.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing VirtualMachines for Provider", "provider", provider.Name)
		return nil
	}
	return vmRequests(vms)
}

func vmRequests(vms *inventoryv1alpha1.VirtualMachineList) []reconcile.Request {
	reqs := make([]reconcile.Request, 0, len(vms.Items))
	for _, vm := range vms.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: vm.Name}})
	}
	return reqs
}

// SetupWithManager registers the VirtualMachine controller with the manager.
func (r *VirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.VirtualMachine{}).
		Watches(&inventoryv1alpha1.Node{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForNode)).
		Watches(&inventoryv1alpha1.Provider{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForProvider)).
		Named("virtualmachine").
		Complete(r)
}
