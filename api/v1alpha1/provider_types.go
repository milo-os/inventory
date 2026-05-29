// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderType classifies the relationship an infrastructure provider has
// with the operator.
//
// +kubebuilder:validation:Enum=Hosting;Colocation;Transit;InternetExchange;DarkFiber;Cloud
type ProviderType string

const (
	// ProviderTypeHosting is a managed/bare-metal hosting provider.
	ProviderTypeHosting ProviderType = "Hosting"
	// ProviderTypeColocation is a colocation/datacenter space provider.
	ProviderTypeColocation ProviderType = "Colocation"
	// ProviderTypeTransit is an IP transit provider.
	ProviderTypeTransit ProviderType = "Transit"
	// ProviderTypeInternetExchange is an internet exchange (IX).
	ProviderTypeInternetExchange ProviderType = "InternetExchange"
	// ProviderTypeDarkFiber is a dark fiber provider.
	ProviderTypeDarkFiber ProviderType = "DarkFiber"
	// ProviderTypeCloud is a cloud provider.
	ProviderTypeCloud ProviderType = "Cloud"
)

// ProviderContract records descriptive contract metadata for a Provider.
// The inventory does not interpret these values; they exist so operators
// can answer "which account/contract is this" without leaving the API.
type ProviderContract struct {
	// ContractID is the provider's contract or agreement identifier.
	//
	// +optional
	ContractID string `json:"contractID,omitempty"`

	// AccountID is the operator's account identifier with the provider.
	//
	// +optional
	AccountID string `json:"accountID,omitempty"`

	// PortalURL is an optional link to the provider's management portal.
	// When set it must be an http:// or https:// URL.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^https?://.+$`
	PortalURL string `json:"portalURL,omitempty"`

	// Notes is free-form contract context.
	//
	// +optional
	Notes string `json:"notes,omitempty"`
}

// ServiceIdentifier is a named identifier the provider assigns to a service
// (e.g. {"ASN","64512"}, {"LOA-CFA","..."}). The inventory does not
// interpret these; they are queryable metadata.
type ServiceIdentifier struct {
	// Name is the identifier's label (e.g. "ASN", "LOA-CFA").
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Identifier is the value the provider assigned.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Identifier string `json:"identifier"`
}

// ProviderSpec defines the desired state of Provider.
type ProviderSpec struct {
	// DisplayName is a human-readable name for the provider.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// Type classifies the provider relationship.
	//
	// +kubebuilder:validation:Required
	Type ProviderType `json:"type"`

	// Contract optionally records contract metadata.
	//
	// +optional
	Contract *ProviderContract `json:"contract,omitempty"`

	// ServiceIdentifiers is an optional list of named identifiers the
	// provider assigns to services.
	//
	// +optional
	// +listType=atomic
	ServiceIdentifiers []ServiceIdentifier `json:"serviceIdentifiers,omitempty"`
}

// ProviderStatus defines the observed state of Provider.
type ProviderStatus struct {
	// Represents the observations of a provider's current state.
	// Known condition types are: "Ready".
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

const (
	// ProviderReady is the condition type indicating that the Provider has
	// been accepted.
	ProviderReady = "Ready"
)

const (
	// ProviderReadyReason indicates the Provider is accepted and ready for use.
	ProviderReadyReason = "Accepted"

	// ProviderPendingReason indicates the Provider has not yet been reconciled.
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

// Provider is the Schema for the providers API. A Provider records an
// infrastructure provider relationship (hosting, colo, transit, IX, dark
// fiber, cloud) and has no parent references.
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec ProviderSpec `json:"spec,omitempty"`

	// +kubebuilder:default={conditions:{{type:"Ready",status:"False",reason:"Pending",message:"Waiting for reconciliation",lastTransitionTime:"1970-01-01T00:00:00Z"}}}
	Status ProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider.
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
