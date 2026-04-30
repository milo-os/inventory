// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CPUArchitecture names a CPU instruction set architecture.
//
// +kubebuilder:validation:Enum=amd64;arm64
type CPUArchitecture string

const (
	// CPUArchitectureAMD64 is the x86_64 instruction set.
	CPUArchitectureAMD64 CPUArchitecture = "amd64"
	// CPUArchitectureARM64 is the AArch64 instruction set.
	CPUArchitectureARM64 CPUArchitecture = "arm64"
)

// DiskType classifies a Node's disk by underlying media / interface.
//
// +kubebuilder:validation:Enum=SSD;HDD;NVMe
type DiskType string

const (
	// DiskTypeSSD is a SATA/SAS solid-state disk.
	DiskTypeSSD DiskType = "SSD"
	// DiskTypeHDD is a spinning hard disk.
	DiskTypeHDD DiskType = "HDD"
	// DiskTypeNVMe is an NVMe solid-state disk.
	DiskTypeNVMe DiskType = "NVMe"
)

// NodeAddressType classifies a Node network address.
//
// +kubebuilder:validation:Enum=Internal;External;Hostname
type NodeAddressType string

const (
	// NodeAddressInternal is a private-network address.
	NodeAddressInternal NodeAddressType = "Internal"
	// NodeAddressExternal is a publicly-routable address.
	NodeAddressExternal NodeAddressType = "External"
	// NodeAddressHostname is a DNS hostname.
	NodeAddressHostname NodeAddressType = "Hostname"
)

// NodeRole is a Node's role within the Cluster it is assigned to.
//
// +kubebuilder:validation:Enum=ControlPlane;Worker
type NodeRole string

const (
	// NodeRoleControlPlane identifies a Kubernetes control-plane node.
	NodeRoleControlPlane NodeRole = "ControlPlane"
	// NodeRoleWorker identifies a Kubernetes worker node.
	NodeRoleWorker NodeRole = "Worker"
)

// NodePhase is a coarse lifecycle indicator for a Node.
//
// +kubebuilder:validation:Enum=Unassigned;Assigned;Unavailable
type NodePhase string

const (
	// NodePhaseUnassigned indicates the Node has no Cluster assignment.
	NodePhaseUnassigned NodePhase = "Unassigned"
	// NodePhaseAssigned indicates the Node is assigned to a Cluster.
	NodePhaseAssigned NodePhase = "Assigned"
	// NodePhaseUnavailable indicates the Node cannot currently be used.
	NodePhaseUnavailable NodePhase = "Unavailable"
)

// NodeDisk describes a single disk attached to a Node.
type NodeDisk struct {
	// Name identifies the disk (e.g. device name or asset label).
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// SizeBytes is the total raw capacity of the disk in bytes.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	SizeBytes int64 `json:"sizeBytes"`

	// Type classifies the disk media.
	//
	// +kubebuilder:validation:Required
	Type DiskType `json:"type"`
}

// NodeHardware describes the physical capabilities of a Node.
type NodeHardware struct {
	// CPUCores is the total number of logical CPU cores available on the node.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	CPUCores int32 `json:"cpuCores"`

	// CPUArchitecture identifies the CPU instruction set.
	//
	// +kubebuilder:validation:Required
	CPUArchitecture CPUArchitecture `json:"cpuArchitecture"`

	// MemoryBytes is the total RAM on the node in bytes.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MemoryBytes int64 `json:"memoryBytes"`

	// Disks is the list of disks attached to the node.
	//
	// +optional
	// +listType=atomic
	Disks []NodeDisk `json:"disks,omitempty"`
}

// NodeAddress is a single reachable address for a Node.
type NodeAddress struct {
	// Type classifies the address.
	//
	// +kubebuilder:validation:Required
	Type NodeAddressType `json:"type"`

	// Address is the IP address or hostname.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Address string `json:"address"`
}

// NodeAssignment records which Cluster the Node belongs to and its role
// within it.
type NodeAssignment struct {
	// ClusterRef references the Cluster the Node is assigned to.
	//
	// +kubebuilder:validation:Required
	ClusterRef LocalObjectReference `json:"clusterRef"`

	// Role is the Node's role within the referenced Cluster.
	//
	// +kubebuilder:validation:Required
	Role NodeRole `json:"role"`
}

// NodeSpec defines the desired state of Node.
//
// +kubebuilder:validation:XValidation:rule="self.siteRef == oldSelf.siteRef",message="siteRef is immutable"
type NodeSpec struct {
	// SiteRef references the Site where this Node physically lives. This
	// field is immutable after creation.
	//
	// +kubebuilder:validation:Required
	SiteRef LocalObjectReference `json:"siteRef"`

	// Hardware describes the Node's physical capabilities.
	//
	// +kubebuilder:validation:Required
	Hardware NodeHardware `json:"hardware"`

	// Addresses is the list of reachable addresses for the Node.
	//
	// +optional
	// +listType=atomic
	Addresses []NodeAddress `json:"addresses,omitempty"`

	// Assignment optionally records the Cluster this Node is a member of.
	// When unset the Node is considered unassigned.
	//
	// +optional
	Assignment *NodeAssignment `json:"assignment,omitempty"`
}

// NodeStatus defines the observed state of Node.
type NodeStatus struct {
	// Represents the observations of a node's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Phase is a coarse lifecycle indicator derived from the Node's
	// assignment and readiness.
	//
	// +optional
	Phase NodePhase `json:"phase,omitempty"`
}

const (
	// NodeReady is the condition type indicating that the Node has been
	// accepted and all of its references resolve.
	NodeReady = "Ready"
)

const (
	// NodeReadyReason indicates the Node is accepted and ready for use.
	NodeReadyReason = "Accepted"

	// NodePendingReason indicates the Node has not yet been reconciled.
	NodePendingReason = "Pending"

	// NodeSiteNotFoundReason indicates the referenced Site does not exist.
	NodeSiteNotFoundReason = "SiteNotFound"

	// NodeClusterNotFoundReason indicates the referenced Cluster (via the
	// Node's assignment) does not exist.
	NodeClusterNotFoundReason = "ClusterNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Site",type="string",JSONPath=".spec.siteRef.name"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.assignment.clusterRef.name"
// +kubebuilder:printcolumn:name="Role",type="string",JSONPath=".spec.assignment.role"
// +kubebuilder:printcolumn:name="Arch",type="string",JSONPath=".spec.hardware.cpuArchitecture"
// +kubebuilder:printcolumn:name="CPU",type="integer",JSONPath=".spec.hardware.cpuCores"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Node is the Schema for the nodes API. A Node is a physical or virtual
// machine that physically lives in a Site and may optionally be assigned to
// a Cluster.
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
