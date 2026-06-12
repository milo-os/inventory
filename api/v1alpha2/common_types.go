// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha2

// NodeReference is a reference to a graph Node by name. Both Edge endpoints
// and any future node-to-node pointers use this type.
type NodeReference struct {
	// Name of the referenced Node.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// AttributeValueType enumerates the scalar value kinds an attribute may hold.
// Attribute values are always stored as strings on the wire; the type drives
// how the admission webhook parses and validates them.
//
// +kubebuilder:validation:Enum=String;Integer;Float;Boolean
type AttributeValueType string

const (
	// AttributeString is an arbitrary UTF-8 string.
	AttributeString AttributeValueType = "String"
	// AttributeInteger is a base-10 signed integer.
	AttributeInteger AttributeValueType = "Integer"
	// AttributeFloat is a decimal floating-point number.
	AttributeFloat AttributeValueType = "Float"
	// AttributeBoolean is "true" or "false".
	AttributeBoolean AttributeValueType = "Boolean"
)

// AttributeSchema describes one allowed attribute key on a Node or Edge type.
// The set of AttributeSchemas on a NodeType/EdgeType is the closed schema the
// admission webhook validates each object's attributes against.
type AttributeSchema struct {
	// Key is the attribute name (the map key in spec.attributes).
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// Type is the scalar kind the value must parse as.
	//
	// +kubebuilder:validation:Required
	Type AttributeValueType `json:"type"`

	// Required marks the attribute as mandatory. Objects missing a required
	// attribute are rejected at admission.
	//
	// +optional
	Required bool `json:"required,omitempty"`

	// Enum optionally restricts a String attribute to a fixed set of values.
	//
	// +optional
	// +listType=atomic
	Enum []string `json:"enum,omitempty"`
}
