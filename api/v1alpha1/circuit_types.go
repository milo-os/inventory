// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CircuitType classifies a Circuit by its role in the network.
//
// +kubebuilder:validation:Enum=CrossConnect;ProviderCircuit;Transit;Peering
type CircuitType string

const (
	// CircuitTypeCrossConnect is an intra-facility cross-connect.
	CircuitTypeCrossConnect CircuitType = "CrossConnect"
	// CircuitTypeProviderCircuit is a provider-delivered circuit.
	CircuitTypeProviderCircuit CircuitType = "ProviderCircuit"
	// CircuitTypeTransit is an IP transit circuit.
	CircuitTypeTransit CircuitType = "Transit"
	// CircuitTypePeering is a peering circuit.
	CircuitTypePeering CircuitType = "Peering"
)

// CircuitEndpointKind names the inventory kind a Circuit end terminates on.
//
// +kubebuilder:validation:Enum=Site;Port
type CircuitEndpointKind string

const (
	// CircuitEndpointKindSite terminates the end at a Site.
	CircuitEndpointKindSite CircuitEndpointKind = "Site"
	// CircuitEndpointKindPort terminates the end at a Port.
	CircuitEndpointKindPort CircuitEndpointKind = "Port"
)

// CircuitEndpoint identifies where one end of a Circuit terminates. The Kind
// selects the inventory kind referenced by Name.
type CircuitEndpoint struct {
	// Kind of the inventory object this end terminates on.
	//
	// +kubebuilder:validation:Required
	Kind CircuitEndpointKind `json:"kind"`

	// Name of the referenced object.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// CircuitSpec defines the desired state of Circuit.
//
// +kubebuilder:validation:XValidation:rule="self.providerRef == oldSelf.providerRef",message="providerRef is immutable"
type CircuitSpec struct {
	// ProviderRef references the Provider that delivers this Circuit. This
	// field is immutable after creation.
	//
	// +kubebuilder:validation:Required
	ProviderRef LocalObjectReference `json:"providerRef"`

	// Type classifies the Circuit.
	//
	// +kubebuilder:validation:Required
	Type CircuitType `json:"type"`

	// CircuitID is the provider's circuit / LOA identifier.
	//
	// +optional
	CircuitID string `json:"circuitID,omitempty"`

	// BandwidthMbps is the Circuit's provisioned bandwidth in megabits per
	// second.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	BandwidthMbps *int64 `json:"bandwidthMbps,omitempty"`

	// AEnd is the A-side termination of the Circuit.
	//
	// +kubebuilder:validation:Required
	AEnd CircuitEndpoint `json:"aEnd"`

	// ZEnd is the Z-side termination of the Circuit.
	//
	// +kubebuilder:validation:Required
	ZEnd CircuitEndpoint `json:"zEnd"`

	// ServiceRef optionally links this Circuit to a platform resource in
	// another API group (e.g. a networking Galactic VPC uplink).
	//
	// +optional
	ServiceRef *ObjectReference `json:"serviceRef,omitempty"`
}

// CircuitStatus defines the observed state of Circuit.
type CircuitStatus struct {
	// Represents the observations of a circuit's current state.
	// Known condition types are: "Ready", "EndpointsResolved".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// CircuitReady is the condition type indicating that the Circuit has been
	// accepted and its Provider and Site endpoints resolve.
	CircuitReady = "Ready"

	// CircuitEndpointsResolved is set by the controller once it has verified
	// every Site-kind endpoint exists. Port-kind endpoints are not yet
	// existence-checked (deferred until the Port kind ships).
	CircuitEndpointsResolved = "EndpointsResolved"
)

const (
	// CircuitReadyReason indicates the Circuit is accepted and ready for use.
	CircuitReadyReason = "Accepted"

	// CircuitPendingReason indicates the Circuit has not yet been reconciled.
	CircuitPendingReason = "Pending"

	// CircuitProviderNotFoundReason indicates the referenced Provider does not
	// exist.
	CircuitProviderNotFoundReason = "ProviderNotFound"

	// CircuitEndpointNotFoundReason indicates a Site-kind endpoint does not
	// exist.
	CircuitEndpointNotFoundReason = "EndpointNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.providerRef.name"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="CircuitID",type="string",JSONPath=".spec.circuitID"
// +kubebuilder:printcolumn:name="Mbps",type="integer",JSONPath=".spec.bandwidthMbps"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Circuit is the Schema for the circuits API. A Circuit records a network
// circuit — cross-connect, provider circuit, transit, or peering — delivered
// by a Provider between two endpoints, optionally linked to a platform
// service in another API group.
type Circuit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec CircuitSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status CircuitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CircuitList contains a list of Circuit.
type CircuitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Circuit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Circuit{}, &CircuitList{})
}
