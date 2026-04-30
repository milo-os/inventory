// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRole classifies a Cluster's function within the fleet.
//
// +kubebuilder:validation:Enum=Compute;Management;Edge;Gateway
type ClusterRole string

const (
	// ClusterRoleCompute runs customer workloads.
	ClusterRoleCompute ClusterRole = "Compute"
	// ClusterRoleManagement runs internal control-plane workloads.
	ClusterRoleManagement ClusterRole = "Management"
	// ClusterRoleEdge runs workloads at an edge site.
	ClusterRoleEdge ClusterRole = "Edge"
	// ClusterRoleGateway serves as a network / traffic gateway.
	ClusterRoleGateway ClusterRole = "Gateway"
)

// ClusterSpec defines the desired state of Cluster.
//
// +kubebuilder:validation:XValidation:rule="self.controlPlaneSiteRef == oldSelf.controlPlaneSiteRef",message="controlPlaneSiteRef is immutable"
type ClusterSpec struct {
	// DisplayName is a human-readable name for the cluster.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// ControlPlaneSiteRef references the Site hosting this cluster's
	// Kubernetes control plane (i.e. where the API server lives). Worker
	// nodes and network devices belonging to this cluster may be at other
	// sites; a cluster's overall footprint is derived from its Nodes and
	// NetworkDevices, not from this field. Immutable after creation.
	//
	// +kubebuilder:validation:Required
	ControlPlaneSiteRef LocalObjectReference `json:"controlPlaneSiteRef"`

	// Role classifies the Cluster.
	//
	// +kubebuilder:validation:Required
	Role ClusterRole `json:"role"`

	// Provider is a free-form identifier for the system that provisioned or
	// manages the cluster, for example "sidero-omni" or "gke". The inventory
	// does not interpret this string.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Provider string `json:"provider"`

	// Endpoint is an optional API server URL for the cluster. When set it
	// must be an http:// or https:// URL.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^https?://.+$`
	Endpoint string `json:"endpoint,omitempty"`
}

// ClusterStatus defines the observed state of Cluster.
type ClusterStatus struct {
	// Represents the observations of a cluster's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// ClusterReady is the condition type indicating that the Cluster has been
	// accepted and its control-plane Site reference resolves.
	ClusterReady = "Ready"
)

const (
	// ClusterReadyReason indicates the Cluster is accepted and ready for use.
	ClusterReadyReason = "Accepted"

	// ClusterPendingReason indicates the Cluster has not yet been reconciled.
	ClusterPendingReason = "Pending"

	// ClusterSiteNotFoundReason indicates the referenced control-plane Site
	// does not exist.
	ClusterSiteNotFoundReason = "SiteNotFound"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="CP-Site",type="string",JSONPath=".spec.controlPlaneSiteRef.name"
// +kubebuilder:printcolumn:name="Role",type="string",JSONPath=".spec.role"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// Cluster is the Schema for the clusters API. A Cluster's control plane
// lives at exactly one Site (controlPlaneSiteRef). Workers (Nodes) and
// NetworkDevices belonging to the cluster may be at other Sites; the
// cluster's full geographic footprint is derived from those child assets.
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec ClusterSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster.
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
