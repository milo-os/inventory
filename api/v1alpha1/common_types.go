// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

// LocalObjectReference is a reference to another cluster-scoped object in the
// inventory.miloapis.com API group by name. All inventory cross-resource
// references use this type rather than raw strings so that CRD schemas are
// explicit about intent.
type LocalObjectReference struct {
	// Name of the referenced object.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// AssetReference is a typed reference to an inventory asset used as a Link
// endpoint. The Kind is restricted to the subset of inventory kinds that can
// participate in a Link.
type AssetReference struct {
	// Kind of the referenced asset. Only a fixed set of inventory kinds may be
	// referenced as a Link endpoint.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Site;Cluster;NetworkDevice
	Kind string `json:"kind"`

	// Name of the referenced asset.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// RackFace identifies which side of a rack a device is mounted on.
//
// +kubebuilder:validation:Enum=Front;Rear
type RackFace string

const (
	// RackFaceFront is the front (cold-aisle) face of a rack.
	RackFaceFront RackFace = "Front"
	// RackFaceRear is the rear (hot-aisle) face of a rack.
	RackFaceRear RackFace = "Rear"
)

// Placement records where a device is physically mounted within a Rack. The
// device occupies the contiguous unit range [StartUnit, StartUnit+UnitHeight-1]
// on the given Face. A validating webhook enforces that the range fits within
// the Rack's HeightU and does not overlap another device on the same face.
type Placement struct {
	// RackRef references the Rack this device is mounted in.
	//
	// +kubebuilder:validation:Required
	RackRef LocalObjectReference `json:"rackRef"`

	// StartUnit is the lowest rack unit (U) the device occupies, counting from
	// 1 at the bottom of the rack.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	StartUnit int32 `json:"startUnit"`

	// UnitHeight is the number of contiguous units the device occupies.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	UnitHeight int32 `json:"unitHeight"`

	// Face is the rack face the device is mounted on. Defaults to Front.
	//
	// +optional
	// +kubebuilder:default=Front
	Face RackFace `json:"face,omitempty"`
}

// PortDeviceReference is a typed reference to the device a Port belongs to.
// The Kind is restricted to the inventory kinds that expose ports (hosts,
// network devices, and rack-mounted PDUs / patch panels).
type PortDeviceReference struct {
	// Kind of the device exposing the port.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Node;NetworkDevice;Rack
	Kind string `json:"kind"`

	// Name of the referenced device.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// Coordinates describes a point on Earth using decimal degrees. Both latitude
// and longitude must fall within their standard ranges.
//
// +kubebuilder:validation:XValidation:rule="self.latitude >= -90.0 && self.latitude <= 90.0",message="latitude must be between -90 and 90 degrees"
// +kubebuilder:validation:XValidation:rule="self.longitude >= -180.0 && self.longitude <= 180.0",message="longitude must be between -180 and 180 degrees"
type Coordinates struct {
	// Latitude in decimal degrees. Must be between -90 and 90 inclusive.
	//
	// +kubebuilder:validation:Required
	Latitude *float64 `json:"latitude"`

	// Longitude in decimal degrees. Must be between -180 and 180 inclusive.
	//
	// +kubebuilder:validation:Required
	Longitude *float64 `json:"longitude"`
}

// Well-known topology label keys. Controllers propagate these onto inventory
// objects so that any client can discover assets by topology attributes using
// standard label selectors.
const (
	// TopologyRegionLabel names the Region an asset belongs to.
	TopologyRegionLabel = "topology.inventory.miloapis.com/region"

	// TopologySiteLabel names the Site an asset belongs to.
	TopologySiteLabel = "topology.inventory.miloapis.com/site"

	// TopologySiteTypeLabel carries the Site's type (Datacenter, Edge, ...).
	TopologySiteTypeLabel = "topology.inventory.miloapis.com/site-type"

	// TopologyClusterLabel names the Cluster an asset is assigned to, if any.
	TopologyClusterLabel = "topology.inventory.miloapis.com/cluster"

	// TopologyRackLabel names the Rack a device is placed in, if any.
	TopologyRackLabel = "topology.inventory.miloapis.com/rack"
)

// FinalizerPrefix is prepended to per-kind finalizer names. v0.1 does not use
// finalizers for deletion blocking (the validating webhook does that), but the
// prefix is reserved here for future cross-version upgrade safety.
const FinalizerPrefix = "inventory.miloapis.com/finalize-"
