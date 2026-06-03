// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VMAllocation describes the compute resources allocated to a VirtualMachine.
type VMAllocation struct {
	// VCPUs is the number of virtual CPUs allocated.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	VCPUs int32 `json:"vcpus"`

	// MemoryBytes is the amount of RAM allocated, in bytes.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MemoryBytes int64 `json:"memoryBytes"`

	// Disks is the list of virtual disks allocated to the VM. The disk shape
	// is shared with Node.
	//
	// +optional
	// +listType=atomic
	Disks []NodeDisk `json:"disks,omitempty"`
}

// VirtualMachineSpec defines the desired state of VirtualMachine. It models
// host assignment and allocation only; power state and health are explicit
// non-goals.
//
// +kubebuilder:validation:XValidation:rule="self.hostRef == oldSelf.hostRef",message="hostRef is immutable"
type VirtualMachineSpec struct {
	// HostRef references the Node this VM runs on. This field is immutable
	// after creation.
	//
	// +kubebuilder:validation:Required
	HostRef LocalObjectReference `json:"hostRef"`

	// ProviderRef optionally references the Provider that owns this VM.
	//
	// +optional
	ProviderRef *LocalObjectReference `json:"providerRef,omitempty"`

	// ProjectRef optionally links this VM to a resourcemanager Project (or
	// other platform resource) in another API group.
	//
	// +optional
	ProjectRef *ObjectReference `json:"projectRef,omitempty"`

	// Allocation describes the compute resources allocated to the VM.
	//
	// +kubebuilder:validation:Required
	Allocation VMAllocation `json:"allocation"`
}

// VirtualMachineStatus defines the observed state of VirtualMachine.
type VirtualMachineStatus struct {
	// Represents the observations of a virtual machine's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// VirtualMachineReady is the condition type indicating that the VM has
	// been accepted and its host (and optional provider) reference resolves.
	VirtualMachineReady = "Ready"
)

const (
	// VirtualMachineReadyReason indicates the VM is accepted and ready.
	VirtualMachineReadyReason = "Accepted"

	// VirtualMachinePendingReason indicates the VM has not yet been reconciled.
	VirtualMachinePendingReason = "Pending"

	// VirtualMachineHostNotFoundReason indicates the referenced host Node does
	// not exist.
	VirtualMachineHostNotFoundReason = "HostNotFound"

	// VirtualMachineProviderNotFoundReason indicates the referenced Provider
	// does not exist.
	VirtualMachineProviderNotFoundReason = "ProviderNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=vm
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.hostRef.name"
// +kubebuilder:printcolumn:name="vCPUs",type="integer",JSONPath=".spec.allocation.vcpus"
// +kubebuilder:printcolumn:name="Memory",type="integer",JSONPath=".spec.allocation.memoryBytes"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.providerRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// VirtualMachine is the Schema for the virtualmachines API. A VirtualMachine
// runs on a host Node and records its compute allocation and optional
// provider/project association.
type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec VirtualMachineSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status VirtualMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualMachineList contains a list of VirtualMachine.
type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachine{}, &VirtualMachineList{})
}
