// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Hosting;Colocation;Transit;InternetExchange;DarkFiber;Cloud
type ProviderType string

const (
	ProviderTypeHosting          ProviderType = "Hosting"
	ProviderTypeColocation       ProviderType = "Colocation"
	ProviderTypeTransit          ProviderType = "Transit"
	ProviderTypeInternetExchange ProviderType = "InternetExchange"
	ProviderTypeDarkFiber        ProviderType = "DarkFiber"
	ProviderTypeCloud            ProviderType = "Cloud"
)

type ProviderContract struct {
	// +optional
	ContractID string `json:"contractID,omitempty"`

	// +optional
	AccountID string `json:"accountID,omitempty"`

	// +optional
	// +kubebuilder:validation:Pattern=`^https?://.+$`
	PortalURL string `json:"portalURL,omitempty"`

	// +optional
	Notes string `json:"notes,omitempty"`
}

type ServiceIdentifier struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Identifier string `json:"identifier"`
}

type ProviderSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// +kubebuilder:validation:Required
	Type ProviderType `json:"type"`

	// +optional
	Contract *ProviderContract `json:"contract,omitempty"`

	// +optional
	// +listType=atomic
	ServiceIdentifiers []ServiceIdentifier `json:"serviceIdentifiers,omitempty"`
}

type ProviderStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	ProviderReady = "Ready"
)

const (
	ProviderReadyReason = "Accepted"

	ProviderPendingReason = "Pending"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec ProviderSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status ProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
