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

// Endpoint Kind strings. Duplicated as literals here because the AssetRef
// field is a free-form string gated by a CEL enum on the type; we compare
// against the same enum values.
const (
	endpointKindSite          = "Site"
	endpointKindCluster       = "Cluster"
	endpointKindNetworkDevice = "NetworkDevice"
)

// LinkReconciler reconciles Link objects. It resolves each endpoint
// reference, marks the Link Ready/EndpointsResolved, and propagates
// topology labels when both endpoints share a region or site.
type LinkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=links,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=links/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=sites,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=inventory.miloapis.com,resources=networkdevices,verbs=get;list;watch

// Reconcile implements reconcile.Reconciler.
func (r *LinkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	link := &inventoryv1alpha1.Link{}
	if err := r.Get(ctx, req.NamespacedName, link); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := link.DeepCopy()

	// Resolve each endpoint. We collect each endpoint's region and site
	// (when known) so that we can compare them after both resolve.
	resolved := make([]endpointInfo, 0, len(link.Spec.Endpoints))
	for i, ep := range link.Spec.Endpoints {
		info, err := r.resolveEndpoint(ctx, ep)
		switch {
		case apierrors.IsNotFound(err):
			SetNotReady(
				link.GetGeneration(),
				&link.Status.Conditions,
				inventoryv1alpha1.LinkEndpointNotFoundReason,
				fmt.Sprintf("endpoint %s/%s not found", ep.Kind, ep.Name),
			)
			// Also reflect the negative status in EndpointsResolved.
			meta.SetStatusCondition(&link.Status.Conditions, metav1.Condition{
				Type:               inventoryv1alpha1.LinkEndpointsResolved,
				Status:             metav1.ConditionFalse,
				Reason:             inventoryv1alpha1.LinkEndpointNotFoundReason,
				Message:            fmt.Sprintf("endpoint %s/%s not found", ep.Kind, ep.Name),
				ObservedGeneration: link.GetGeneration(),
			})
			if statusErr := r.patchStatusIfChanged(ctx, originalStatus, link); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: requeueAfterMissingRef}, nil
		case err != nil:
			log.Error(err, "failed to resolve Link endpoint", "index", i, "kind", ep.Kind, "name", ep.Name)
			return ctrl.Result{}, err
		}
		resolved = append(resolved, info)
	}

	// Compute the label set for this Link — a key is propagated only when
	// both endpoints agree on it. Using empty strings for disagreements
	// causes ensureLabels to remove any stale value.
	want := map[string]string{
		inventoryv1alpha1.TopologyRegionLabel: commonLabel(resolved, inventoryv1alpha1.TopologyRegionLabel),
		inventoryv1alpha1.TopologySiteLabel:   commonLabel(resolved, inventoryv1alpha1.TopologySiteLabel),
	}

	originalSpec := link.DeepCopy()
	if ensureLabels(link, want) {
		if err := r.Patch(ctx, link, client.MergeFrom(originalSpec)); err != nil {
			log.Error(err, "failed to patch Link labels")
			return ctrl.Result{}, err
		}
	}

	SetReady(link.GetGeneration(), &link.Status.Conditions, inventoryv1alpha1.LinkReadyReason, "Link accepted")
	meta.SetStatusCondition(&link.Status.Conditions, metav1.Condition{
		Type:               inventoryv1alpha1.LinkEndpointsResolved,
		Status:             metav1.ConditionTrue,
		Reason:             inventoryv1alpha1.LinkReadyReason,
		Message:            "All endpoints resolved",
		ObservedGeneration: link.GetGeneration(),
	})

	if err := r.patchStatusIfChanged(ctx, originalStatus, link); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// endpointInfo captures the subset of a resolved endpoint that the Link
// controller cares about — specifically, the topology labels it inherits
// from the endpoint object.
type endpointInfo struct {
	labels map[string]string
}

// resolveEndpoint looks up the endpoint object by Kind and Name and
// returns its topology labels. If the Kind is unrecognized it is
// treated as "not found" via apierrors.NewNotFound so the caller can
// surface it with the usual reason. In practice CEL on the type prevents
// this, but we guard anyway.
func (r *LinkReconciler) resolveEndpoint(ctx context.Context, ep inventoryv1alpha1.AssetReference) (endpointInfo, error) {
	key := client.ObjectKey{Name: ep.Name}
	switch ep.Kind {
	case endpointKindSite:
		obj := &inventoryv1alpha1.Site{}
		if err := r.Get(ctx, key, obj); err != nil {
			return endpointInfo{}, err
		}
		return endpointInfo{labels: obj.GetLabels()}, nil
	case endpointKindCluster:
		obj := &inventoryv1alpha1.Cluster{}
		if err := r.Get(ctx, key, obj); err != nil {
			return endpointInfo{}, err
		}
		return endpointInfo{labels: obj.GetLabels()}, nil
	case endpointKindNetworkDevice:
		obj := &inventoryv1alpha1.NetworkDevice{}
		if err := r.Get(ctx, key, obj); err != nil {
			return endpointInfo{}, err
		}
		return endpointInfo{labels: obj.GetLabels()}, nil
	default:
		// An unknown Kind is equivalent to a missing object for our
		// purposes — the controller can report an EndpointNotFound reason
		// and the webhook will usually already have rejected the write.
		return endpointInfo{}, apierrors.NewBadRequest(fmt.Sprintf("unknown endpoint kind %q", ep.Kind))
	}
}

// commonLabel returns the label value shared by every endpoint under the
// given key, or "" if they disagree or any endpoint lacks it.
func commonLabel(eps []endpointInfo, key string) string {
	if len(eps) == 0 {
		return ""
	}
	first, ok := eps[0].labels[key]
	if !ok || first == "" {
		return ""
	}
	for _, ep := range eps[1:] {
		if v, ok := ep.labels[key]; !ok || v != first {
			return ""
		}
	}
	return first
}

func (r *LinkReconciler) patchStatusIfChanged(ctx context.Context, original, current *inventoryv1alpha1.Link) error {
	if reflect.DeepEqual(original.Status, current.Status) {
		return nil
	}
	return r.Status().Patch(ctx, current, client.MergeFrom(original))
}

// enqueueForEndpointKind returns a MapFunc that enqueues Links whose
// endpoints reference obj.Name with a matching Kind. We verify Kind in the
// map function because the endpoint-name indexer is kind-agnostic — it
// emits every endpoint name regardless of kind, so two different kinds
// with the same name would otherwise produce false positives.
func (r *LinkReconciler) enqueueForEndpointKind(kind string) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		name := obj.GetName()
		if name == "" {
			return nil
		}
		links := &inventoryv1alpha1.LinkList{}
		if err := r.List(ctx, links, client.MatchingFields{IndexLinkEndpointName: name}); err != nil {
			logf.FromContext(ctx).Error(err, "listing Links for endpoint", "kind", kind, "name", name)
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(links.Items))
		for _, l := range links.Items {
			for _, ep := range l.Spec.Endpoints {
				if ep.Kind == kind && ep.Name == name {
					reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Name: l.Name}})
					break
				}
			}
		}
		return reqs
	}
}

// SetupWithManager registers the Link controller with the manager.
func (r *LinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Link{}).
		Watches(&inventoryv1alpha1.Site{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForEndpointKind(endpointKindSite))).
		Watches(&inventoryv1alpha1.Cluster{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForEndpointKind(endpointKindCluster))).
		Watches(&inventoryv1alpha1.NetworkDevice{}, handler.EnqueueRequestsFromMapFunc(r.enqueueForEndpointKind(endpointKindNetworkDevice))).
		Named("link").
		Complete(r)
}
