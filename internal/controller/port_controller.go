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

// Device Kind strings for Port.deviceRef. Compared against the CEL enum on
// the type.
const (
	deviceKindNode          = "Node"
	deviceKindNetworkDevice = "NetworkDevice"
	deviceKindRack          = "Rack"
)

// PortReconciler reconciles Port objects. It validates the Port's deviceRef
// and propagates the device's topology labels onto the Port.
type PortReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventRecorder emits best-effort activity events on Ready transitions.
	// May be nil (e.g. in unit tests), in which case emission is a no-op.
	EventRecorder *EventRecorder
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=ports,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=ports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=networkdevices,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=racks,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *PortReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	port := &inventoryv1alpha1.Port{}
	if err := r.Get(ctx, req.NamespacedName, port); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := port.DeepCopy()

	device, err := r.getDevice(ctx, port.Spec.DeviceRef)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			port.GetGeneration(),
			&port.Status.Conditions,
			inventoryv1alpha1.PortDeviceNotFoundReason,
			fmt.Sprintf("%s %q not found", port.Spec.DeviceRef.Kind, port.Spec.DeviceRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, port); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced device", "kind", port.Spec.DeviceRef.Kind, "name", port.Spec.DeviceRef.Name)
		return ctrl.Result{}, err
	}

	want := copyTopologyLabels(device)

	originalSpec := port.DeepCopy()
	if ensureLabels(port, want) {
		if err := r.Patch(ctx, port, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Port labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(port.GetGeneration(), &port.Status.Conditions, inventoryv1alpha1.PortReadyReason, "Port accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, port); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// getDevice resolves the device a Port references by Kind and Name. An
// unrecognized Kind is treated as a bad request (CEL normally prevents it).
func (r *PortReconciler) getDevice(ctx context.Context, ref inventoryv1alpha1.PortDeviceReference) (client.Object, error) {
	key := client.ObjectKey{Name: ref.Name}
	var obj client.Object
	switch ref.Kind {
	case deviceKindNode:
		obj = &inventoryv1alpha1.Node{}
	case deviceKindNetworkDevice:
		obj = &inventoryv1alpha1.NetworkDevice{}
	case deviceKindRack:
		obj = &inventoryv1alpha1.Rack{}
	default:
		return nil, apierrors.NewBadRequest(fmt.Sprintf("unknown device kind %q", ref.Kind))
	}
	if err := r.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *PortReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Port) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	if err := r.Status().Patch(ctx, current, client.MergeFrom(original)); err != nil {
		return err
	}
	r.EventRecorder.EmitReadyTransition(ctx, current,
		inventoryv1alpha1.GroupVersion.WithKind("Port"),
		current.Name,
		original.Status.Conditions, current.Status.Conditions)
	return nil
}

// enqueueForDeviceKind returns a MapFunc that enqueues Ports whose deviceRef
// references obj.Name with a matching Kind. Kind is checked in the map
// function because the deviceRef-name indexer is kind-agnostic.
func (r *PortReconciler) enqueueForDeviceKind(kind string) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		name := obj.GetName()
		if name == "" {
			return nil
		}
		ports := &inventoryv1alpha1.PortList{}
		if err := r.List(ctx, ports, client.MatchingFields{IndexPortDeviceRef: name}); err != nil {
			logf.FromContext(ctx).Error(err, "listing Ports for device", "kind", kind, "name", name)
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(ports.Items))
		for _, p := range ports.Items {
			if p.Spec.DeviceRef.Kind == kind && p.Spec.DeviceRef.Name == name {
				reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: p.Name}})
			}
		}
		return reqs
	}
}

// SetupWithManager registers the Port controller with the manager.
func (r *PortReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Port{}).
		Watches(&inventoryv1alpha1.Node{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForDeviceKind(deviceKindNode))).
		Watches(&inventoryv1alpha1.NetworkDevice{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForDeviceKind(deviceKindNetworkDevice))).
		Watches(&inventoryv1alpha1.Rack{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForDeviceKind(deviceKindRack))).
		Named("port").
		Complete(r)
}
