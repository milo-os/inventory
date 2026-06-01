// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PortType classifies a Port by the kind of connection it terminates.
//
// +kubebuilder:validation:Enum=Power;Serial;Ethernet;Optical;PatchPanel
type PortType string

const (
	// PortTypePower is a power inlet / outlet (e.g. a PSU or PDU socket).
	PortTypePower PortType = "Power"
	// PortTypeSerial is a serial console port.
	PortTypeSerial PortType = "Serial"
	// PortTypeEthernet is a copper Ethernet interface.
	PortTypeEthernet PortType = "Ethernet"
	// PortTypeOptical is an optical (fiber) interface.
	PortTypeOptical PortType = "Optical"
	// PortTypePatchPanel is a patch-panel termination.
	PortTypePatchPanel PortType = "PatchPanel"
)

// PortSpec defines the desired state of Port.
//
// +kubebuilder:validation:XValidation:rule="self.deviceRef == oldSelf.deviceRef",message="deviceRef is immutable"
type PortSpec struct {
	// DeviceRef references the device that exposes this Port. This field is
	// immutable after creation.
	//
	// +kubebuilder:validation:Required
	DeviceRef PortDeviceReference `json:"deviceRef"`

	// Type classifies the Port.
	//
	// +kubebuilder:validation:Required
	Type PortType `json:"type"`

	// Name identifies the Port on its device (e.g. "eth0", "PSU1",
	// "pp-A-24").
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Speed is the nominal line rate of a network Port, expressed as a
	// Kubernetes Quantity (e.g. "10G", "400G"). Unset for non-network ports.
	//
	// +optional
	Speed *resource.Quantity `json:"speed,omitempty"`
}

// PortStatus defines the observed state of Port.
type PortStatus struct {
	// Represents the observations of a port's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// PortReady is the condition type indicating that the Port has been
	// accepted and its device reference resolves.
	PortReady = "Ready"
)

const (
	// PortReadyReason indicates the Port is accepted and ready for use.
	PortReadyReason = "Accepted"

	// PortPendingReason indicates the Port has not yet been reconciled.
	PortPendingReason = "Pending"

	// PortDeviceNotFoundReason indicates the referenced device does not exist.
	PortDeviceNotFoundReason = "DeviceNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Device-Kind",type="string",JSONPath=".spec.deviceRef.kind"
// +kubebuilder:printcolumn:name="Device",type="string",JSONPath=".spec.deviceRef.name"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"
// +kubebuilder:printcolumn:name="Speed",type="string",JSONPath=".spec.speed"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Port is the Schema for the ports API. A Port is a connection point exposed
// by a Node, NetworkDevice, or Rack-mounted device, and is the near/far-end
// identifier a Cable connects.
type Port struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec PortSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status PortStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PortList contains a list of Port.
type PortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Port `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Port{}, &PortList{})
}
