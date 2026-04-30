// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegionSpec defines the desired state of Region.
type RegionSpec struct {
	// DisplayName is a human-readable name for the region (e.g. "US East").
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// Coordinates optionally records a representative latitude/longitude for
	// the region. This is descriptive only and not used for routing.
	//
	// +optional
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}

// RegionStatus defines the observed state of Region.
type RegionStatus struct {
	// Represents the observations of a region's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// RegionReady is the condition type indicating that the Region has been
	// accepted and all references (if any) resolve.
	RegionReady = "Ready"
)

const (
	// RegionReadyReason indicates the Region is accepted and ready for use.
	RegionReadyReason = "Accepted"

	// RegionPendingReason indicates the Region has not yet been reconciled.
	RegionPendingReason = "Pending"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Region is the Schema for the regions API. A Region is the top-level
// geographic grouping in the inventory and has no parent references.
type Region struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec RegionSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status RegionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegionList contains a list of Region.
type RegionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Region `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Region{}, &RegionList{})
}
