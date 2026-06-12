// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeTypeSpec describes the closed attribute schema for Nodes whose
// spec.type equals this NodeType's metadata.name.
type NodeTypeSpec struct {
	// DisplayName is a human-readable label for the node type.
	//
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Attributes is the closed set of attribute keys a Node of this type may
	// carry. Keys outside this set are rejected at admission; keys marked
	// Required must be present.
	//
	// +optional
	// +listType=map
	// +listMapKey=key
	Attributes []AttributeSchema `json:"attributes,omitempty"`
}

// NodeTypeStatus defines the observed state of a NodeType.
type NodeTypeStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// NodeTypeReady indicates the NodeType has been accepted.
	NodeTypeReady = "Ready"
	// NodeTypeReadyReason indicates the NodeType is accepted and ready.
	NodeTypeReadyReason = "Accepted"
	// NodeTypePendingReason indicates the NodeType has not yet reconciled.
	NodeTypePendingReason = "Pending"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// NodeType describes the attribute schema for a class of graph Nodes. It is
// the source of truth the admission webhook validates Node.spec.attributes
// against. See RFC milo-os/inventory#43.
type NodeType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec NodeTypeSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status NodeTypeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeTypeList contains a list of NodeType.
type NodeTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeType{}, &NodeTypeList{})
}
