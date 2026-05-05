// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SiteType classifies a Site's physical or logical nature.
//
// +kubebuilder:validation:Enum=Datacenter;AvailabilityZone;Edge;Virtual
type SiteType string

const (
	// SiteTypeDatacenter is a traditional datacenter facility.
	SiteTypeDatacenter SiteType = "Datacenter"
	// SiteTypeAvailabilityZone is a cloud-provider-style failure domain.
	SiteTypeAvailabilityZone SiteType = "AvailabilityZone"
	// SiteTypeEdge is an edge / point-of-presence site.
	SiteTypeEdge SiteType = "Edge"
	// SiteTypeVirtual is a logical site that has no single physical location.
	SiteTypeVirtual SiteType = "Virtual"
)

// SiteSpec defines the desired state of Site.
//
// +kubebuilder:validation:XValidation:rule="self.regionRef == oldSelf.regionRef",message="regionRef is immutable"
type SiteSpec struct {
	// DisplayName is a human-readable name for the site.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// RegionRef references the Region this Site belongs to. This field is
	// immutable after creation.
	//
	// +kubebuilder:validation:Required
	RegionRef LocalObjectReference `json:"regionRef"`

	// Type classifies the Site.
	//
	// +kubebuilder:validation:Required
	Type SiteType `json:"type"`

	// Address is an optional free-form postal/street address for the site.
	//
	// +optional
	Address string `json:"address,omitempty"`
}

// SiteStatus defines the observed state of Site.
type SiteStatus struct {
	// Represents the observations of a site's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// SiteReady is the condition type indicating that the Site has been
	// accepted and its Region reference resolves.
	SiteReady = "Ready"
)

const (
	// SiteReadyReason indicates the Site is accepted and ready for use.
	SiteReadyReason = "Accepted"

	// SitePendingReason indicates the Site has not yet been reconciled.
	SitePendingReason = "Pending"

	// SiteRegionNotFoundReason indicates the referenced Region does not exist.
	SiteRegionNotFoundReason = "RegionNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.regionRef.name"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Site is the Schema for the sites API. A Site belongs to exactly one Region
// and is the parent of Nodes, Clusters, and NetworkDevices.
type Site struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec SiteSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status SiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SiteList contains a list of Site.
type SiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Site `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Site{}, &SiteList{})
}
