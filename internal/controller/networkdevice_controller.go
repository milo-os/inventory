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

// NetworkDeviceReconciler reconciles NetworkDevice objects. It validates
// the ClusterRef and SiteRef and propagates topology labels from the
// parent Cluster. The validating webhook ensures the NetworkDevice's
// SiteRef matches its Cluster's SiteRef, so cluster-derived labels are
// authoritative for both region and site.
type NetworkDeviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=networkdevices,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=networkdevices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *NetworkDeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	device := &inventoryv1alpha1.NetworkDevice{}
	if err := r.Get(ctx, req.NamespacedName, device); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := device.DeepCopy()

	// Validate ClusterRef first — the cluster is what drives label
	// propagation, and the webhook guarantees siteRef is consistent with
	// cluster.siteRef, so checking cluster before site keeps the "most
	// specific missing ref" reporting clean.
	cluster := &inventoryv1alpha1.Cluster{}
	err := r.Get(ctx, client.ObjectKey{Name: device.Spec.ClusterRef.Name}, cluster)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			device.GetGeneration(),
			&device.Status.Conditions,
			inventoryv1alpha1.NetworkDeviceClusterNotFoundReason,
			fmt.Sprintf("Cluster %q not found", device.Spec.ClusterRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, device); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Cluster", "cluster", device.Spec.ClusterRef.Name)
		return ctrl.Result{}, err
	}

	// Validate SiteRef.
	site := &inventoryv1alpha1.Site{}
	err = r.Get(ctx, client.ObjectKey{Name: device.Spec.SiteRef.Name}, site)
	switch {
	case apierrors.IsNotFound(err):
		SetNotReady(
			device.GetGeneration(),
			&device.Status.Conditions,
			inventoryv1alpha1.NetworkDeviceSiteNotFoundReason,
			fmt.Sprintf("Site %q not found", device.Spec.SiteRef.Name),
		)
		if statusErr := r.patchStatusIfChanged(ctx, originalStatus, device); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
	case err != nil:
		log.Error(err, "failed to get referenced Site", "site", device.Spec.SiteRef.Name)
		return ctrl.Result{}, err
	}

	// Propagate cluster labels. The cluster already carries the region +
	// site + site-type it inherited from its Site.
	want := clusterLabels(cluster)
	want[inventoryv1alpha1.TopologyRackLabel] = rackLabel(device.Spec.Placement)

	originalSpec := device.DeepCopy()
	if ensureLabels(device, want) {
		if err := r.Patch(ctx, device, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch NetworkDevice labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(device.GetGeneration(), &device.Status.Conditions, inventoryv1alpha1.NetworkDeviceReadyReason, "NetworkDevice accepted")

	if err := r.patchStatusIfChanged(ctx, originalStatus, device); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *NetworkDeviceReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.NetworkDevice) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	return r.Status().Patch(ctx, current, client.MergeFrom(original))
}

func (r *NetworkDeviceReconciler) enqueueForCluster(ctx context.Context, obj client.Object) []reconcile.Request {
	cluster, ok := obj.(*inventoryv1alpha1.Cluster)
	if !ok {
		return nil
	}
	devices := &inventoryv1alpha1.NetworkDeviceList{}
	if err := r.List(ctx, devices, client.MatchingFields{IndexNetworkDeviceClusterRef: cluster.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing NetworkDevices for Cluster", "cluster", cluster.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(devices.Items))
	for _, d := range devices.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: d.Name}})
	}
	return reqs
}

func (r *NetworkDeviceReconciler) enqueueForSite(ctx context.Context, obj client.Object) []reconcile.Request {
	site, ok := obj.(*inventoryv1alpha1.Site)
	if !ok {
		return nil
	}
	devices := &inventoryv1alpha1.NetworkDeviceList{}
	if err := r.List(ctx, devices, client.MatchingFields{IndexNetworkDeviceSiteRef: site.Name}); err != nil {
		logf.FromContext(ctx).Error(err, "listing NetworkDevices for Site", "site", site.Name)
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(devices.Items))
	for _, d := range devices.Items {
		reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: d.Name}})
	}
	return reqs
}

// SetupWithManager registers the NetworkDevice controller with the manager.
func (r *NetworkDeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.NetworkDevice{}).
		Watches(&inventoryv1alpha1.Cluster{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForCluster)).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForSite)).
		Named("networkdevice").
		Complete(r)
}
