// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EndpointConstraint restricts which NodeTypes an Edge's endpoints may point
// at. An empty list on either side means "any node type".
type EndpointConstraint struct {
	// FromTypes is the set of allowed NodeType names for the edge's `from`
	// endpoint. Empty means any.
	//
	// +optional
	// +listType=set
	FromTypes []string `json:"fromTypes,omitempty"`

	// ToTypes is the set of allowed NodeType names for the edge's `to`
	// endpoint. Empty means any.
	//
	// +optional
	// +listType=set
	ToTypes []string `json:"toTypes,omitempty"`
}

// EdgeTypeSpec describes the closed attribute schema and endpoint constraints
// for Edges whose spec.type equals this EdgeType's metadata.name.
type EdgeTypeSpec struct {
	// DisplayName is a human-readable label for the edge type.
	//
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Endpoints constrains the NodeTypes the edge may connect.
	//
	// +optional
	Endpoints EndpointConstraint `json:"endpoints,omitempty"`

	// Attributes is the closed set of attribute keys an Edge of this type may
	// carry.
	//
	// +optional
	// +listType=map
	// +listMapKey=key
	Attributes []AttributeSchema `json:"attributes,omitempty"`
}

// EdgeTypeStatus defines the observed state of an EdgeType.
type EdgeTypeStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// EdgeTypeReady indicates the EdgeType has been accepted.
	EdgeTypeReady = "Ready"
	// EdgeTypeReadyReason indicates the EdgeType is accepted and ready.
	EdgeTypeReadyReason = "Accepted"
	// EdgeTypePendingReason indicates the EdgeType has not yet reconciled.
	EdgeTypePendingReason = "Pending"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// EdgeType describes the attribute schema and endpoint constraints for a
// class of graph Edges. See RFC milo-os/inventory#43.
type EdgeType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec EdgeTypeSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status EdgeTypeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EdgeTypeList contains a list of EdgeType.
type EdgeTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeType{}, &EdgeTypeList{})
}
