// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeSpec defines the desired state of a graph Node.
//
// +kubebuilder:validation:XValidation:rule="self.type == oldSelf.type",message="type is immutable"
type NodeSpec struct {
	// Type names the node's asset class. The value must match the name of an
	// existing NodeType, which describes the attributes this node may carry.
	// Typical values are the former v1alpha1 kinds: Region, Site, Cluster,
	// NetworkDevice, Rack, Provider, Circuit, Cable, Port, VirtualMachine,
	// Host (the former compute Node), Fleet. This field is immutable.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`

	// Attributes is the node's key/value attribute bag. The admission webhook
	// validates these against the matching NodeType's attribute schema.
	//
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NodeStatus defines the observed state of a graph Node.
type NodeStatus struct {
	// Represents the observations of a node's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// NodeReady is the condition type indicating the Node has been accepted
	// and validated against its NodeType.
	NodeReady = "Ready"
)

const (
	// NodeReadyReason indicates the Node is accepted and ready for use.
	NodeReadyReason = "Accepted"

	// NodePendingReason indicates the Node has not yet been reconciled.
	NodePendingReason = "Pending"

	// NodeTypeNotFoundReason indicates the referenced NodeType does not exist.
	NodeTypeNotFoundReason = "NodeTypeNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=invnode
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Node is a vertex in the inventory property graph. Its asset class is given
// by spec.type and its descriptive data by spec.attributes. It supersedes the
// per-kind v1alpha1 inventory kinds. See RFC milo-os/inventory#43.
type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec NodeSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status NodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeList contains a list of Node.
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Node `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
