// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LinkType classifies a Link's nature.
//
// +kubebuilder:validation:Enum=Physical;Logical;Internet
type LinkType string

const (
	// LinkTypePhysical is a physical cable / fiber between two assets.
	LinkTypePhysical LinkType = "Physical"
	// LinkTypeLogical is an overlay / tunnel between two assets.
	LinkTypeLogical LinkType = "Logical"
	// LinkTypeInternet represents transit over the public internet.
	LinkTypeInternet LinkType = "Internet"
)

// LinkSpec defines the desired state of Link.
//
// +kubebuilder:validation:XValidation:rule="self.endpoints[0] != self.endpoints[1]",message="link endpoints must be distinct"
type LinkSpec struct {
	// Endpoints are the two assets this Link connects. The two endpoints
	// must refer to different assets.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	// +listType=atomic
	Endpoints []AssetReference `json:"endpoints"`

	// Type classifies the Link.
	//
	// +kubebuilder:validation:Required
	Type LinkType `json:"type"`

	// CapacityMbps is the Link's nominal capacity in megabits per second.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	CapacityMbps *int64 `json:"capacityMbps,omitempty"`

	// LatencyMs is the Link's nominal one-way latency in milliseconds,
	// expressed as a dimensionless Kubernetes Quantity (e.g. "5", "250m"
	// for 0.25 ms). The field name encodes the unit; values are numeric.
	//
	// +optional
	LatencyMs *resource.Quantity `json:"latencyMs,omitempty"`
}

// LinkStatus defines the observed state of Link.
type LinkStatus struct {
	// Represents the observations of a link's current state.
	// Known condition types are: "Ready", "EndpointsResolved".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// LinkReady is the condition type indicating that the Link has been
	// accepted and both endpoints resolve.
	LinkReady = "Ready"

	// LinkEndpointsResolved is a condition type set by the controller once
	// it has verified that both endpoint objects exist. It is reported in
	// addition to Ready for observability.
	LinkEndpointsResolved = "EndpointsResolved"
)

const (
	// LinkReadyReason indicates the Link is accepted and ready for use.
	LinkReadyReason = "Accepted"

	// LinkPendingReason indicates the Link has not yet been reconciled.
	LinkPendingReason = "Pending"

	// LinkEndpointNotFoundReason indicates at least one endpoint object
	// does not exist.
	LinkEndpointNotFoundReason = "EndpointNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="A-Kind",type="string",JSONPath=".spec.endpoints[0].kind"
// +kubebuilder:printcolumn:name="A-Name",type="string",JSONPath=".spec.endpoints[0].name"
// +kubebuilder:printcolumn:name="B-Kind",type="string",JSONPath=".spec.endpoints[1].kind"
// +kubebuilder:printcolumn:name="B-Name",type="string",JSONPath=".spec.endpoints[1].name"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Mbps",type="integer",JSONPath=".spec.capacityMbps"
// +kubebuilder:printcolumn:name="Latency",type="string",JSONPath=".spec.latencyMs"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Link is the Schema for the links API. A Link records connectivity between
// two inventory assets (Sites, Clusters, or NetworkDevices).
type Link struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec LinkSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status LinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LinkList contains a list of Link.
type LinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Link `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Link{}, &LinkList{})
}
