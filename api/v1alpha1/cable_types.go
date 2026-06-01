// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CableMedia classifies the physical medium of a Cable.
//
// +kubebuilder:validation:Enum=Copper;FiberSMF;FiberMMF;Power;DAC
type CableMedia string

const (
	// CableMediaCopper is a copper twisted-pair run.
	CableMediaCopper CableMedia = "Copper"
	// CableMediaFiberSMF is single-mode fiber.
	CableMediaFiberSMF CableMedia = "FiberSMF"
	// CableMediaFiberMMF is multi-mode fiber.
	CableMediaFiberMMF CableMedia = "FiberMMF"
	// CableMediaPower is a power cable.
	CableMediaPower CableMedia = "Power"
	// CableMediaDAC is a direct-attach copper cable.
	CableMediaDAC CableMedia = "DAC"
)

// CableSpec defines the desired state of Cable. A Cable records the physical
// run between exactly two Ports — the near-end and far-end. It is distinct
// from the logical Link, which records connectivity/capacity between assets;
// a Link may reference the Cable(s) that realize it.
//
// +kubebuilder:validation:XValidation:rule="self.endpoints[0].name != self.endpoints[1].name",message="cable endpoints must be distinct"
type CableSpec struct {
	// Endpoints are the two Ports this Cable connects. The two endpoints must
	// reference different Ports.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	// +listType=atomic
	Endpoints []LocalObjectReference `json:"endpoints"`

	// Media classifies the physical medium.
	//
	// +kubebuilder:validation:Required
	Media CableMedia `json:"media"`

	// LengthM is the cable length in meters, expressed as a Kubernetes
	// Quantity (e.g. "3", "0.5").
	//
	// +optional
	LengthM *resource.Quantity `json:"lengthM,omitempty"`

	// Label is a free-form operator label for the run (e.g. its sticker).
	//
	// +optional
	Label string `json:"label,omitempty"`
}

// CableStatus defines the observed state of Cable.
type CableStatus struct {
	// Represents the observations of a cable's current state.
	// Known condition types are: "Ready", "EndpointsResolved".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// CableReady is the condition type indicating that the Cable has been
	// accepted and both endpoint Ports resolve.
	CableReady = "Ready"

	// CableEndpointsResolved is set by the controller once it has verified
	// that both endpoint Ports exist. It is reported alongside Ready for
	// observability.
	CableEndpointsResolved = "EndpointsResolved"
)

const (
	// CableReadyReason indicates the Cable is accepted and ready for use.
	CableReadyReason = "Accepted"

	// CablePendingReason indicates the Cable has not yet been reconciled.
	CablePendingReason = "Pending"

	// CableEndpointNotFoundReason indicates at least one endpoint Port does
	// not exist.
	CableEndpointNotFoundReason = "EndpointNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Media",type="string",JSONPath=".spec.media"
// +kubebuilder:printcolumn:name="A",type="string",JSONPath=".spec.endpoints[0].name"
// +kubebuilder:printcolumn:name="Z",type="string",JSONPath=".spec.endpoints[1].name"
// +kubebuilder:printcolumn:name="Length",type="string",JSONPath=".spec.lengthM"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Cable is the Schema for the cables API. A Cable is the physical run between
// two Ports.
type Cable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec CableSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status CableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CableList contains a list of Cable.
type CableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cable `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cable{}, &CableList{})
}
