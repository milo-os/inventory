// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PowerPhase classifies the electrical phase configuration of a rack power
// feed.
//
// +kubebuilder:validation:Enum=SinglePhase;ThreePhase
type PowerPhase string

const (
	// PowerPhaseSingle is a single-phase feed.
	PowerPhaseSingle PowerPhase = "SinglePhase"
	// PowerPhaseThree is a three-phase feed.
	PowerPhaseThree PowerPhase = "ThreePhase"
)

// RackPowerFeed describes a single power feed delivered to a Rack.
type RackPowerFeed struct {
	// Name identifies the feed (e.g. "A", "B").
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Phase is the electrical phase configuration of the feed.
	//
	// +optional
	Phase PowerPhase `json:"phase,omitempty"`

	// Voltage is the nominal feed voltage in volts.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	Voltage int32 `json:"voltage,omitempty"`

	// AmpsRated is the rated current of the feed in amperes.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	AmpsRated int32 `json:"ampsRated,omitempty"`
}

// RackSpec defines the desired state of Rack. Cage and row are modeled as
// free-form attributes rather than their own kinds to avoid kind explosion.
//
// +kubebuilder:validation:XValidation:rule="self.siteRef == oldSelf.siteRef",message="siteRef is immutable"
type RackSpec struct {
	// SiteRef references the Site this Rack physically lives in. This field is
	// immutable after creation.
	//
	// +kubebuilder:validation:Required
	SiteRef LocalObjectReference `json:"siteRef"`

	// Cage is the free-form cage identifier within the Site.
	//
	// +optional
	Cage string `json:"cage,omitempty"`

	// Row is the free-form row identifier within the cage.
	//
	// +optional
	Row string `json:"row,omitempty"`

	// Name is a human-readable name for the rack (e.g. its asset label).
	//
	// +optional
	Name string `json:"name,omitempty"`

	// HeightU is the number of mountable rack units (U) the rack provides.
	// Devices placed in the rack occupy unit positions 1..HeightU.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	HeightU int32 `json:"heightU"`

	// PowerFeeds enumerates the power feeds delivered to the rack.
	//
	// +optional
	// +listType=atomic
	PowerFeeds []RackPowerFeed `json:"powerFeeds,omitempty"`
}

// RackStatus defines the observed state of Rack.
type RackStatus struct {
	// Represents the observations of a rack's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// RackReady is the condition type indicating that the Rack has been
	// accepted and its Site reference resolves.
	RackReady = "Ready"
)

const (
	// RackReadyReason indicates the Rack is accepted and ready for use.
	RackReadyReason = "Accepted"

	// RackPendingReason indicates the Rack has not yet been reconciled.
	RackPendingReason = "Pending"

	// RackSiteNotFoundReason indicates the referenced Site does not exist.
	RackSiteNotFoundReason = "SiteNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Site",type="string",JSONPath=".spec.siteRef.name"
// +kubebuilder:printcolumn:name="Cage",type="string",JSONPath=".spec.cage"
// +kubebuilder:printcolumn:name="Row",type="string",JSONPath=".spec.row"
// +kubebuilder:printcolumn:name="HeightU",type="integer",JSONPath=".spec.heightU"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Rack is the Schema for the racks API. A Rack belongs to exactly one Site and
// is the mount point for placed Nodes and NetworkDevices.
type Rack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec RackSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status RackStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RackList contains a list of Rack.
type RackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rack{}, &RackList{})
}
