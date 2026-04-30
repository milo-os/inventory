// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkDeviceRole classifies a NetworkDevice's function within the fabric.
//
// +kubebuilder:validation:Enum=BorderRouter;Spine;Leaf;Firewall
type NetworkDeviceRole string

const (
	// NetworkDeviceRoleBorderRouter connects the fabric to an external network.
	NetworkDeviceRoleBorderRouter NetworkDeviceRole = "BorderRouter"
	// NetworkDeviceRoleSpine is a spine switch in a Clos fabric.
	NetworkDeviceRoleSpine NetworkDeviceRole = "Spine"
	// NetworkDeviceRoleLeaf is a leaf / top-of-rack switch.
	NetworkDeviceRoleLeaf NetworkDeviceRole = "Leaf"
	// NetworkDeviceRoleFirewall is a firewall appliance.
	NetworkDeviceRoleFirewall NetworkDeviceRole = "Firewall"
)

// NetworkDeviceSpec defines the desired state of NetworkDevice.
//
// +kubebuilder:validation:XValidation:rule="self.clusterRef == oldSelf.clusterRef",message="clusterRef is immutable"
// +kubebuilder:validation:XValidation:rule="self.siteRef == oldSelf.siteRef",message="siteRef is immutable"
type NetworkDeviceSpec struct {
	// ClusterRef references the Cluster this device is part of. This field
	// is immutable after creation.
	//
	// +kubebuilder:validation:Required
	ClusterRef LocalObjectReference `json:"clusterRef"`

	// SiteRef references the Site where the device physically lives. A
	// validating webhook additionally requires that this Site matches the
	// referenced Cluster's SiteRef. This field is immutable after creation.
	//
	// +kubebuilder:validation:Required
	SiteRef LocalObjectReference `json:"siteRef"`

	// Role classifies the device's function.
	//
	// +kubebuilder:validation:Required
	Role NetworkDeviceRole `json:"role"`

	// ManagementAddress is an optional address at which operators can reach
	// the device's management plane. The inventory does not connect to it.
	//
	// +optional
	ManagementAddress string `json:"managementAddress,omitempty"`
}

// NetworkDeviceStatus defines the observed state of NetworkDevice.
type NetworkDeviceStatus struct {
	// Represents the observations of a network device's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// NetworkDeviceReady is the condition type indicating that the
	// NetworkDevice has been accepted and all of its references resolve.
	NetworkDeviceReady = "Ready"
)

const (
	// NetworkDeviceReadyReason indicates the NetworkDevice is accepted and
	// ready for use.
	NetworkDeviceReadyReason = "Accepted"

	// NetworkDevicePendingReason indicates the NetworkDevice has not yet
	// been reconciled.
	NetworkDevicePendingReason = "Pending"

	// NetworkDeviceClusterNotFoundReason indicates the referenced Cluster
	// does not exist.
	NetworkDeviceClusterNotFoundReason = "ClusterNotFound"

	// NetworkDeviceSiteNotFoundReason indicates the referenced Site does not
	// exist.
	NetworkDeviceSiteNotFoundReason = "SiteNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Site",type="string",JSONPath=".spec.siteRef.name"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name"
// +kubebuilder:printcolumn:name="Role",type="string",JSONPath=".spec.role"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// NetworkDevice is the Schema for the networkdevices API. A NetworkDevice is
// a switch, router, or firewall that is part of a Cluster and physically
// lives in a Site.
type NetworkDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec NetworkDeviceSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status NetworkDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkDeviceList contains a list of NetworkDevice.
type NetworkDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkDevice{}, &NetworkDeviceList{})
}
