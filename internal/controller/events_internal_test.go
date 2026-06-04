// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"context"
	"errors"
	"testing"

	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// fakeEventClient records the Events passed to Create so tests can assert on
// emission without a live API server.
type fakeEventClient struct {
	created []*eventsv1.Event
	err     error
}

func (f *fakeEventClient) Create(_ context.Context, e *eventsv1.Event, _ metav1.CreateOptions) (*eventsv1.Event, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.created = append(f.created, e)
	return e, nil
}

func ready(reason string) *metav1.Condition {
	return &metav1.Condition{Type: readyConditionType, Status: metav1.ConditionTrue, Reason: reason}
}

func notReady(reason string) *metav1.Condition {
	return &metav1.Condition{Type: readyConditionType, Status: metav1.ConditionFalse, Reason: reason}
}

func TestReadyTransitioned(t *testing.T) {
	tests := []struct {
		name       string
		prev, curr *metav1.Condition
		wantEmit   bool
		wantReady  bool
	}{
		{"nil->true emits ready", nil, ready("Accepted"), true, true},
		{"false->true emits ready", notReady("ProviderNotFound"), ready("Accepted"), true, true},
		{"true->true no emit", ready("Accepted"), ready("Accepted"), false, true},
		{"nil->false emits not-ready (initial)", nil, notReady("ProviderNotFound"), true, false},
		{"true->false emits not-ready", ready("Accepted"), notReady("ProviderNotFound"), true, false},
		{"false->false same reason no emit", notReady("ProviderNotFound"), notReady("ProviderNotFound"), false, false},
		{"false->false reason change emits", notReady("ProviderNotFound"), notReady("EndpointNotFound"), true, false},
		{"nil curr no emit", ready("Accepted"), nil, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			emit, isReady := readyTransitioned(tc.prev, tc.curr)
			if emit != tc.wantEmit || isReady != tc.wantReady {
				t.Fatalf("readyTransitioned() = (%v,%v), want (%v,%v)", emit, isReady, tc.wantEmit, tc.wantReady)
			}
		})
	}
}

// subject builds a minimal cluster-scoped client.Object for emission tests.
func subject(name, uid string) *metav1.PartialObjectMetadata {
	return &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(uid), Generation: 3},
	}
}

func conds(c *metav1.Condition) []metav1.Condition {
	if c == nil {
		return nil
	}
	return []metav1.Condition{*c}
}

func TestEmitReadyTransition_ReadyEvent(t *testing.T) {
	fc := &fakeEventClient{}
	r := &EventRecorder{client: fc, namespace: "default", instance: "pod-1"}
	gvk := schema.GroupVersionKind{Group: "inventory.miloapis.com", Version: "v1alpha1", Kind: "Region"}

	r.EmitReadyTransition(context.Background(), subject("us-east", "uid-1"), gvk, "US East",
		conds(nil), conds(ready("Accepted")))

	if len(fc.created) != 1 {
		t.Fatalf("expected 1 event, got %d", len(fc.created))
	}
	e := fc.created[0]
	if e.Reason != EventReasonReady {
		t.Errorf("reason = %q, want %q", e.Reason, EventReasonReady)
	}
	if e.Type != "Normal" {
		t.Errorf("type = %q, want Normal", e.Type)
	}
	if got := e.Annotations[AnnotationEventType]; got != "inventory.region.ready" {
		t.Errorf("event-type annotation = %q, want inventory.region.ready", got)
	}
	if got := e.Annotations[AnnotationDisplayName]; got != "US East" {
		t.Errorf("display-name annotation = %q, want US East", got)
	}
	if got := e.Annotations[AnnotationConditionReason]; got != "Accepted" {
		t.Errorf("condition-reason annotation = %q, want Accepted", got)
	}
	if e.Regarding.Kind != "Region" || e.Regarding.Name != "us-east" || string(e.Regarding.UID) != "uid-1" {
		t.Errorf("regarding = %+v, want Region/us-east/uid-1", e.Regarding)
	}
	if e.Namespace != "default" {
		t.Errorf("event namespace = %q, want default", e.Namespace)
	}
}

func TestEmitReadyTransition_NotReadyEvent(t *testing.T) {
	fc := &fakeEventClient{}
	r := &EventRecorder{client: fc, namespace: "default"}
	gvk := schema.GroupVersionKind{Group: "inventory.miloapis.com", Version: "v1alpha1", Kind: "NetworkDevice"}

	curr := notReady("ClusterNotFound")
	curr.Message = `Cluster "c1" not found`
	r.EmitReadyTransition(context.Background(), subject("nd-1", "uid-2"), gvk, "nd-1",
		conds(ready("Accepted")), conds(curr))

	if len(fc.created) != 1 {
		t.Fatalf("expected 1 event, got %d", len(fc.created))
	}
	e := fc.created[0]
	if e.Reason != EventReasonNotReady || e.Type != "Warning" {
		t.Errorf("got reason=%q type=%q, want NotReady/Warning", e.Reason, e.Type)
	}
	if got := e.Annotations[AnnotationEventType]; got != "inventory.networkdevice.not_ready" {
		t.Errorf("event-type annotation = %q, want inventory.networkdevice.not_ready", got)
	}
	if e.Note != `Cluster "c1" not found` {
		t.Errorf("note = %q, want the condition message", e.Note)
	}
}

func TestEmitReadyTransition_NoEmitOnNoTransition(t *testing.T) {
	fc := &fakeEventClient{}
	r := &EventRecorder{client: fc, namespace: "default"}
	gvk := schema.GroupVersionKind{Group: "inventory.miloapis.com", Version: "v1alpha1", Kind: "Site"}

	r.EmitReadyTransition(context.Background(), subject("s1", "uid"), gvk, "s1",
		conds(ready("Accepted")), conds(ready("Accepted")))

	if len(fc.created) != 0 {
		t.Fatalf("expected no event on stable Ready, got %d", len(fc.created))
	}
}

func TestEmitReadyTransition_NilSafe(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "inventory.miloapis.com", Version: "v1alpha1", Kind: "Site"}

	// nil recorder must not panic.
	var nilRec *EventRecorder
	nilRec.EmitReadyTransition(context.Background(), subject("s1", "uid"), gvk, "s1", nil, conds(ready("Accepted")))

	// recorder with nil client must not panic.
	(&EventRecorder{}).EmitReadyTransition(context.Background(), subject("s1", "uid"), gvk, "s1", nil, conds(ready("Accepted")))
}

func TestEmitReadyTransition_CreateErrorSwallowed(t *testing.T) {
	fc := &fakeEventClient{err: errors.New("boom")}
	r := &EventRecorder{client: fc, namespace: "default"}
	gvk := schema.GroupVersionKind{Group: "inventory.miloapis.com", Version: "v1alpha1", Kind: "Region"}

	// Must not panic or propagate the error.
	r.EmitReadyTransition(context.Background(), subject("r1", "uid"), gvk, "r1", nil, conds(ready("Accepted")))
}
