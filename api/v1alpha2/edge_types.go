// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EdgeSpec defines the desired state of a graph Edge.
//
// +kubebuilder:validation:XValidation:rule="self.type == oldSelf.type",message="type is immutable"
// +kubebuilder:validation:XValidation:rule="self.from.name != self.to.name",message="edge endpoints must be distinct"
type EdgeSpec struct {
	// Type names the relationship class. The value must match the name of an
	// existing EdgeType, which constrains the endpoint node types and the
	// attributes this edge may carry. Typical values: located-in, member-of,
	// mounted-in, connects, realized-by, provided-by. This field is immutable.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`

	// From is the source endpoint Node.
	//
	// +kubebuilder:validation:Required
	From NodeReference `json:"from"`

	// To is the target endpoint Node.
	//
	// +kubebuilder:validation:Required
	To NodeReference `json:"to"`

	// Attributes is the edge's key/value attribute bag. The admission webhook
	// validates these against the matching EdgeType's attribute schema.
	//
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
}

// EdgeStatus defines the observed state of a graph Edge.
type EdgeStatus struct {
	// Represents the observations of an edge's current state.
	// Known condition types are: "Ready", "EndpointsResolved".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// EdgeReady is the condition type indicating the Edge has been accepted
	// and both endpoints resolve.
	EdgeReady = "Ready"

	// EdgeEndpointsResolved is set once the controller has verified both
	// endpoint Nodes exist. Reported alongside Ready for observability.
	EdgeEndpointsResolved = "EndpointsResolved"
)

const (
	// EdgeReadyReason indicates the Edge is accepted and ready for use.
	EdgeReadyReason = "Accepted"

	// EdgePendingReason indicates the Edge has not yet been reconciled.
	EdgePendingReason = "Pending"

	// EdgeTypeNotFoundReason indicates the referenced EdgeType does not exist.
	EdgeTypeNotFoundReason = "EdgeTypeNotFound"

	// EdgeEndpointNotFoundReason indicates an endpoint Node does not exist.
	EdgeEndpointNotFoundReason = "EndpointNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="From",type="string",JSONPath=".spec.from.name"
// +kubebuilder:printcolumn:name="To",type="string",JSONPath=".spec.to.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Edge is a directed relationship between two graph Nodes. Its relationship
// class is given by spec.type and its descriptive data by spec.attributes. It
// supersedes the v1alpha1 Link/Cable/Circuit kinds and the inline parent refs.
// See RFC milo-os/inventory#43.
type Edge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec EdgeSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status EdgeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EdgeList contains a list of Edge.
type EdgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Edge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Edge{}, &EdgeList{})
}
