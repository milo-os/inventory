// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Annotation keys carried on emitted Events. The Activity system's
// ActivityPolicy event rules read these to render human-readable timelines, so
// the keys here form a contract with config/milo/activity/policies/*. All keys
// are prefixed with the inventory API group per the platform convention.
const (
	// AnnotationEventType carries the dot-notation activity type, e.g.
	// "inventory.site.ready" or "inventory.circuit.not_ready". Policy event
	// rules match on this for a deterministic, kind-specific selector.
	AnnotationEventType = "inventory.miloapis.com/event-type"
	// AnnotationObservedGeneration carries the object generation the event
	// reflects.
	AnnotationObservedGeneration = "inventory.miloapis.com/observed-generation"
	// AnnotationResourceKind carries the inventory Kind (e.g. "Site").
	AnnotationResourceKind = "inventory.miloapis.com/resource-kind"
	// AnnotationResourceName carries the object name.
	AnnotationResourceName = "inventory.miloapis.com/resource-name"
	// AnnotationDisplayName carries a pre-computed human-friendly name so that
	// CEL summary templates stay simple (display name falls back to the object
	// name when a kind has no spec.displayName).
	AnnotationDisplayName = "inventory.miloapis.com/display-name"
	// AnnotationConditionReason carries the Ready condition reason backing the
	// event (e.g. "Accepted", "ProviderNotFound").
	AnnotationConditionReason = "inventory.miloapis.com/condition-reason"
)

// Event reason codes (PascalCase per the Kubernetes convention). CRUD
// lifecycle events (created/updated/deleted) are intentionally absent — those
// are captured by audit logs. Only async controller outcomes that audit logs
// cannot observe are emitted here.
const (
	// EventReasonReady marks a Ready=True transition.
	EventReasonReady = "Ready"
	// EventReasonNotReady marks a Ready=False transition. The specific
	// condition reason (e.g. "ProviderNotFound") is carried in the
	// AnnotationConditionReason annotation and in the event note.
	EventReasonNotReady = "NotReady"
)

const (
	// eventReportingController identifies the inventory controller manager as
	// the source of emitted events.
	eventReportingController = "inventory.miloapis.com/inventory-controller"
	// defaultEventNamespace is where Events for cluster-scoped inventory
	// objects are created. events.k8s.io/v1 Events are namespaced, but
	// inventory kinds are cluster-scoped (regarding.namespace is empty), so
	// the Event object itself needs a home namespace. Kubernetes places
	// events for cluster-scoped objects in "default"; the Activity system
	// reads events across all namespaces. Overridable via
	// INVENTORY_ACTIVITY_NAMESPACE.
	defaultEventNamespace = "default"
)

// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

// EventClient is the minimal interface for creating events.k8s.io/v1 Events.
// Obtained via kubernetes.Clientset.EventsV1().Events(namespace). Defined as an
// interface so tests can substitute a fake without a live API server.
type EventClient interface {
	Create(ctx context.Context, event *eventsv1.Event, opts metav1.CreateOptions) (*eventsv1.Event, error)
}

// EventRecorder emits best-effort events.k8s.io/v1 Events on Ready condition
// transitions. Events are emitted in addition to status conditions, not
// instead of them: conditions remain the machine-readable status while events
// feed the Activity system's human-readable timelines. All emission is
// best-effort — failures are logged and swallowed so they never block
// reconciliation. A nil *EventRecorder is a no-op, so controllers can run
// without one (e.g. in unit tests).
type EventRecorder struct {
	client    EventClient
	namespace string
	instance  string
}

// NewEventRecorder builds an EventRecorder from the manager's rest.Config. The
// recorder writes Event objects into the namespace named by
// INVENTORY_ACTIVITY_NAMESPACE (default "default") and stamps ReportingInstance
// with POD_NAME.
func NewEventRecorder(cfg *rest.Config) (*EventRecorder, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building events clientset: %w", err)
	}
	ns := os.Getenv("INVENTORY_ACTIVITY_NAMESPACE")
	if ns == "" {
		ns = defaultEventNamespace
	}
	return &EventRecorder{
		client:    cs.EventsV1().Events(ns),
		namespace: ns,
		instance:  os.Getenv("POD_NAME"),
	}, nil
}

// EmitReadyTransition emits an Event when the object's Ready condition
// transitions in a way worth surfacing on a timeline:
//
//   - to True (from absent or False): a Normal "Ready" event;
//   - to False (from absent, True, or a different reason): a Warning
//     "NotReady" event carrying the specific condition reason.
//
// Emitting on the first-time False (absent -> False) is deliberate: it lets a
// "created X referencing a missing Y" entry appear before the later "X is
// ready" entry once Y exists. The recorder must be called after the status
// patch succeeds so that curr reflects the persisted conditions.
func (r *EventRecorder) EmitReadyTransition(
	ctx context.Context,
	obj client.Object,
	gvk schema.GroupVersionKind,
	displayName string,
	prev, curr []metav1.Condition,
) {
	if r == nil || r.client == nil {
		return
	}

	currReady := meta.FindStatusCondition(curr, readyConditionType)
	if currReady == nil {
		return
	}
	prevReady := meta.FindStatusCondition(prev, readyConditionType)

	emit, ready := readyTransitioned(prevReady, currReady)
	if !emit {
		return
	}

	kindLower := strings.ToLower(gvk.Kind)
	annotations := map[string]string{
		AnnotationObservedGeneration: strconv.FormatInt(obj.GetGeneration(), 10),
		AnnotationResourceKind:       gvk.Kind,
		AnnotationResourceName:       obj.GetName(),
		AnnotationDisplayName:        displayName,
		AnnotationConditionReason:    currReady.Reason,
	}

	var reason, eventType, note string
	if ready {
		reason = EventReasonReady
		eventType = "Normal"
		annotations[AnnotationEventType] = "inventory." + kindLower + ".ready"
		note = currReady.Message
	} else {
		reason = EventReasonNotReady
		eventType = "Warning"
		annotations[AnnotationEventType] = "inventory." + kindLower + ".not_ready"
		note = currReady.Message
	}

	evt := r.buildEvent(obj.GetName(), annotations, eventType, reason, note)
	evt.Regarding = corev1.ObjectReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       obj.GetName(),
		UID:        obj.GetUID(),
	}

	if _, err := r.client.Create(ctx, evt, metav1.CreateOptions{}); err != nil {
		logf.FromContext(ctx).Error(err, "failed to emit inventory activity event",
			"reason", reason, "kind", gvk.Kind, "name", obj.GetName())
	}
}

// buildEvent constructs an eventsv1.Event with a unique name and the common
// fields shared by every inventory event. The caller sets Regarding (and
// optionally Related).
func (r *EventRecorder) buildEvent(
	subjectName string,
	annotations map[string]string,
	eventType, reason, note string,
) *eventsv1.Event {
	// Unique name <subject>.<nanosecond-hex> avoids conflicts on rapid
	// re-emission. Trim to the 253-char name limit if needed.
	name := fmt.Sprintf("%s.%x", subjectName, time.Now().UnixNano())
	if len(name) > 253 {
		name = name[len(name)-253:]
	}
	return &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   r.namespace,
			Annotations: annotations,
		},
		EventTime:           metav1.NewMicroTime(time.Now()),
		Action:              reason,
		Reason:              reason,
		Note:                note,
		Type:                eventType,
		ReportingController: eventReportingController,
		ReportingInstance:   r.instance,
	}
}

// readyTransitioned reports whether a Ready condition change warrants an event,
// and whether the new state is Ready=True.
//
//   - to True from non-True (or absent): (true, true)
//   - to False from True, absent, or a different reason: (true, false)
//   - no meaningful change: (false, _)
func readyTransitioned(prev, curr *metav1.Condition) (emit, ready bool) {
	if curr == nil {
		return false, false
	}
	if curr.Status == metav1.ConditionTrue {
		if prev == nil || prev.Status != metav1.ConditionTrue {
			return true, true
		}
		return false, true
	}
	// curr is Ready=False (or Unknown): emit on first observation, on a flip
	// from True, or when the reason changed (e.g. ProviderNotFound ->
	// EndpointNotFound).
	if prev == nil || prev.Status == metav1.ConditionTrue || prev.Reason != curr.Reason {
		return true, false
	}
	return false, false
}

// displayNameOrName returns display when non-empty, otherwise name. Kinds with
// a spec.displayName pass it through; kinds without one fall back to the
// object name.
func displayNameOrName(display, name string) string {
	if display != "" {
		return display
	}
	return name
}
