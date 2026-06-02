# API Reference

Packages:

- [inventory.miloapis.com/v1alpha1](#inventorymiloapiscomv1alpha1)

# inventory.miloapis.com/v1alpha1

Resource Types:

- [Cable](#cable)

- [Cluster](#cluster)

- [Link](#link)

- [NetworkDevice](#networkdevice)

- [Node](#node)

- [Port](#port)

- [Provider](#provider)

- [Rack](#rack)

- [Region](#region)

- [Site](#site)




## Cable
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Cable is the Schema for the cables API. A Cable is the physical run between
two Ports.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Cable</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#cablespec">spec</a></b></td>
        <td>object</td>
        <td>
          CableSpec defines the desired state of Cable. A Cable records the physical
run between exactly two Ports — the near-end and far-end. It is distinct
from the logical Link, which records connectivity/capacity between assets;
a Link may reference the Cable(s) that realize it.<br/>
          <br/>
            <i>Validations</i>:<li>self.endpoints[0].name != self.endpoints[1].name: cable endpoints must be distinct</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#cablestatus">status</a></b></td>
        <td>object</td>
        <td>
          CableStatus defines the observed state of Cable.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cable.spec
<sup><sup>[↩ Parent](#cable)</sup></sup>



CableSpec defines the desired state of Cable. A Cable records the physical
run between exactly two Ports — the near-end and far-end. It is distinct
from the logical Link, which records connectivity/capacity between assets;
a Link may reference the Cable(s) that realize it.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#cablespecendpointsindex">endpoints</a></b></td>
        <td>[]object</td>
        <td>
          Endpoints are the two Ports this Cable connects. The two endpoints must
reference different Ports.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>media</b></td>
        <td>enum</td>
        <td>
          Media classifies the physical medium.<br/>
          <br/>
            <i>Enum</i>: Copper, FiberSMF, FiberMMF, Power, DAC<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>label</b></td>
        <td>string</td>
        <td>
          Label is a free-form operator label for the run (e.g. its sticker).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lengthM</b></td>
        <td>int or string</td>
        <td>
          LengthM is the cable length in meters, expressed as a Kubernetes
Quantity (e.g. "3", "0.5").<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cable.spec.endpoints[index]
<sup><sup>[↩ Parent](#cablespec)</sup></sup>



LocalObjectReference is a reference to another cluster-scoped object in the
inventory.miloapis.com API group by name. All inventory cross-resource
references use this type rather than raw strings so that CRD schemas are
explicit about intent.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Cable.status
<sup><sup>[↩ Parent](#cable)</sup></sup>



CableStatus defines the observed state of Cable.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#cablestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a cable's current state.
Known condition types are: "Ready", "EndpointsResolved".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cable.status.conditions[index]
<sup><sup>[↩ Parent](#cablestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Cluster
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Cluster is the Schema for the clusters API. A Cluster's control plane
lives at exactly one Site (controlPlaneSiteRef). Workers (Nodes) and
NetworkDevices belonging to the cluster may be at other Sites; the
cluster's full geographic footprint is derived from those child assets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Cluster</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#clusterspec">spec</a></b></td>
        <td>object</td>
        <td>
          ClusterSpec defines the desired state of Cluster.<br/>
          <br/>
            <i>Validations</i>:<li>self.controlPlaneSiteRef == oldSelf.controlPlaneSiteRef: controlPlaneSiteRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#clusterstatus">status</a></b></td>
        <td>object</td>
        <td>
          ClusterStatus defines the observed state of Cluster.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cluster.spec
<sup><sup>[↩ Parent](#cluster)</sup></sup>



ClusterSpec defines the desired state of Cluster.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#clusterspeccontrolplanesiteref">controlPlaneSiteRef</a></b></td>
        <td>object</td>
        <td>
          ControlPlaneSiteRef references the Site hosting this cluster's
Kubernetes control plane (i.e. where the API server lives). Worker
nodes and network devices belonging to this cluster may be at other
sites; a cluster's overall footprint is derived from its Nodes and
NetworkDevices, not from this field. Immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>displayName</b></td>
        <td>string</td>
        <td>
          DisplayName is a human-readable name for the cluster.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>provider</b></td>
        <td>string</td>
        <td>
          Provider is a free-form identifier for the system that provisioned or
manages the cluster, for example "sidero-omni" or "gke". The inventory
does not interpret this string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>enum</td>
        <td>
          Role classifies the Cluster.<br/>
          <br/>
            <i>Enum</i>: Compute, Management, Edge, Gateway<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>endpoint</b></td>
        <td>string</td>
        <td>
          Endpoint is an optional API server URL for the cluster. When set it
must be an http:// or https:// URL.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cluster.spec.controlPlaneSiteRef
<sup><sup>[↩ Parent](#clusterspec)</sup></sup>



ControlPlaneSiteRef references the Site hosting this cluster's
Kubernetes control plane (i.e. where the API server lives). Worker
nodes and network devices belonging to this cluster may be at other
sites; a cluster's overall footprint is derived from its Nodes and
NetworkDevices, not from this field. Immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Cluster.status
<sup><sup>[↩ Parent](#cluster)</sup></sup>



ClusterStatus defines the observed state of Cluster.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#clusterstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a cluster's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Cluster.status.conditions[index]
<sup><sup>[↩ Parent](#clusterstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Link
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Link is the Schema for the links API. A Link records connectivity between
two inventory assets (Sites, Clusters, or NetworkDevices).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Link</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#linkspec">spec</a></b></td>
        <td>object</td>
        <td>
          LinkSpec defines the desired state of Link.<br/>
          <br/>
            <i>Validations</i>:<li>self.endpoints[0] != self.endpoints[1]: link endpoints must be distinct</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#linkstatus">status</a></b></td>
        <td>object</td>
        <td>
          LinkStatus defines the observed state of Link.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Link.spec
<sup><sup>[↩ Parent](#link)</sup></sup>



LinkSpec defines the desired state of Link.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#linkspecendpointsindex">endpoints</a></b></td>
        <td>[]object</td>
        <td>
          Endpoints are the two assets this Link connects. The two endpoints
must refer to different assets.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type classifies the Link.<br/>
          <br/>
            <i>Enum</i>: Physical, Logical, Internet<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#linkspeccablerefsindex">cableRefs</a></b></td>
        <td>[]object</td>
        <td>
          CableRefs optionally records the physical Cable(s) that realize this
logical Link.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>capacityMbps</b></td>
        <td>integer</td>
        <td>
          CapacityMbps is the Link's nominal capacity in megabits per second.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>latencyMs</b></td>
        <td>int or string</td>
        <td>
          LatencyMs is the Link's nominal one-way latency in milliseconds,
expressed as a dimensionless Kubernetes Quantity (e.g. "5", "250m"
for 0.25 ms). The field name encodes the unit; values are numeric.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Link.spec.endpoints[index]
<sup><sup>[↩ Parent](#linkspec)</sup></sup>



AssetReference is a typed reference to an inventory asset used as a Link
endpoint. The Kind is restricted to the subset of inventory kinds that can
participate in a Link.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          Kind of the referenced asset. Only a fixed set of inventory kinds may be
referenced as a Link endpoint.<br/>
          <br/>
            <i>Enum</i>: Site, Cluster, NetworkDevice<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced asset.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Link.spec.cableRefs[index]
<sup><sup>[↩ Parent](#linkspec)</sup></sup>



LocalObjectReference is a reference to another cluster-scoped object in the
inventory.miloapis.com API group by name. All inventory cross-resource
references use this type rather than raw strings so that CRD schemas are
explicit about intent.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Link.status
<sup><sup>[↩ Parent](#link)</sup></sup>



LinkStatus defines the observed state of Link.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#linkstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a link's current state.
Known condition types are: "Ready", "EndpointsResolved".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Link.status.conditions[index]
<sup><sup>[↩ Parent](#linkstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## NetworkDevice
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






NetworkDevice is the Schema for the networkdevices API. A NetworkDevice is
a switch, router, or firewall that is part of a Cluster and physically
lives in a Site.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>NetworkDevice</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#networkdevicespec">spec</a></b></td>
        <td>object</td>
        <td>
          NetworkDeviceSpec defines the desired state of NetworkDevice.<br/>
          <br/>
            <i>Validations</i>:<li>self.clusterRef == oldSelf.clusterRef: clusterRef is immutable</li><li>self.siteRef == oldSelf.siteRef: siteRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#networkdevicestatus">status</a></b></td>
        <td>object</td>
        <td>
          NetworkDeviceStatus defines the observed state of NetworkDevice.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### NetworkDevice.spec
<sup><sup>[↩ Parent](#networkdevice)</sup></sup>



NetworkDeviceSpec defines the desired state of NetworkDevice.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#networkdevicespecclusterref">clusterRef</a></b></td>
        <td>object</td>
        <td>
          ClusterRef references the Cluster this device is part of. This field
is immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>enum</td>
        <td>
          Role classifies the device's function.<br/>
          <br/>
            <i>Enum</i>: BorderRouter, Spine, Leaf, Firewall<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#networkdevicespecsiteref">siteRef</a></b></td>
        <td>object</td>
        <td>
          SiteRef references the Site where the device physically lives. A
validating webhook additionally requires that this Site matches the
referenced Cluster's SiteRef. This field is immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>managementAddress</b></td>
        <td>string</td>
        <td>
          ManagementAddress is an optional address at which operators can reach
the device's management plane. The inventory does not connect to it.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#networkdevicespecplacement">placement</a></b></td>
        <td>object</td>
        <td>
          Placement optionally records where the device is physically mounted in a
Rack. A validating webhook enforces fit and non-overlap.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### NetworkDevice.spec.clusterRef
<sup><sup>[↩ Parent](#networkdevicespec)</sup></sup>



ClusterRef references the Cluster this device is part of. This field
is immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### NetworkDevice.spec.siteRef
<sup><sup>[↩ Parent](#networkdevicespec)</sup></sup>



SiteRef references the Site where the device physically lives. A
validating webhook additionally requires that this Site matches the
referenced Cluster's SiteRef. This field is immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### NetworkDevice.spec.placement
<sup><sup>[↩ Parent](#networkdevicespec)</sup></sup>



Placement optionally records where the device is physically mounted in a
Rack. A validating webhook enforces fit and non-overlap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#networkdevicespecplacementrackref">rackRef</a></b></td>
        <td>object</td>
        <td>
          RackRef references the Rack this device is mounted in.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>startUnit</b></td>
        <td>integer</td>
        <td>
          StartUnit is the lowest rack unit (U) the device occupies, counting from
1 at the bottom of the rack.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>unitHeight</b></td>
        <td>integer</td>
        <td>
          UnitHeight is the number of contiguous units the device occupies.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>face</b></td>
        <td>enum</td>
        <td>
          Face is the rack face the device is mounted on. Defaults to Front.<br/>
          <br/>
            <i>Enum</i>: Front, Rear<br/>
            <i>Default</i>: Front<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### NetworkDevice.spec.placement.rackRef
<sup><sup>[↩ Parent](#networkdevicespecplacement)</sup></sup>



RackRef references the Rack this device is mounted in.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### NetworkDevice.status
<sup><sup>[↩ Parent](#networkdevice)</sup></sup>



NetworkDeviceStatus defines the observed state of NetworkDevice.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#networkdevicestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a network device's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### NetworkDevice.status.conditions[index]
<sup><sup>[↩ Parent](#networkdevicestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Node
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Node is the Schema for the nodes API. A Node is a physical or virtual
machine that physically lives in a Site and may optionally be assigned to
a Cluster.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Node</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#nodespec">spec</a></b></td>
        <td>object</td>
        <td>
          NodeSpec defines the desired state of Node.<br/>
          <br/>
            <i>Validations</i>:<li>self.siteRef == oldSelf.siteRef: siteRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#nodestatus">status</a></b></td>
        <td>object</td>
        <td>
          NodeStatus defines the observed state of Node.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Node.spec
<sup><sup>[↩ Parent](#node)</sup></sup>



NodeSpec defines the desired state of Node.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#nodespechardware">hardware</a></b></td>
        <td>object</td>
        <td>
          Hardware describes the Node's physical capabilities.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#nodespecsiteref">siteRef</a></b></td>
        <td>object</td>
        <td>
          SiteRef references the Site where this Node physically lives. This
field is immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#nodespecaddressesindex">addresses</a></b></td>
        <td>[]object</td>
        <td>
          Addresses is the list of reachable addresses for the Node.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#nodespecassignment">assignment</a></b></td>
        <td>object</td>
        <td>
          Assignment optionally records the Cluster this Node is a member of.
When unset the Node is considered unassigned.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#nodespecplacement">placement</a></b></td>
        <td>object</td>
        <td>
          Placement optionally records where the Node is physically mounted in a
Rack. A validating webhook enforces fit and non-overlap.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Node.spec.hardware
<sup><sup>[↩ Parent](#nodespec)</sup></sup>



Hardware describes the Node's physical capabilities.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>cpuArchitecture</b></td>
        <td>enum</td>
        <td>
          CPUArchitecture identifies the CPU instruction set.<br/>
          <br/>
            <i>Enum</i>: amd64, arm64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cpuCores</b></td>
        <td>integer</td>
        <td>
          CPUCores is the total number of logical CPU cores available on the node.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>memoryBytes</b></td>
        <td>integer</td>
        <td>
          MemoryBytes is the total RAM on the node in bytes.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#nodespechardwaredisksindex">disks</a></b></td>
        <td>[]object</td>
        <td>
          Disks is the list of disks attached to the node.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Node.spec.hardware.disks[index]
<sup><sup>[↩ Parent](#nodespechardware)</sup></sup>



NodeDisk describes a single disk attached to a Node.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name identifies the disk (e.g. device name or asset label).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sizeBytes</b></td>
        <td>integer</td>
        <td>
          SizeBytes is the total raw capacity of the disk in bytes.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type classifies the disk media.<br/>
          <br/>
            <i>Enum</i>: SSD, HDD, NVMe<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.spec.siteRef
<sup><sup>[↩ Parent](#nodespec)</sup></sup>



SiteRef references the Site where this Node physically lives. This
field is immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.spec.addresses[index]
<sup><sup>[↩ Parent](#nodespec)</sup></sup>



NodeAddress is a single reachable address for a Node.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is the IP address or hostname.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type classifies the address.<br/>
          <br/>
            <i>Enum</i>: Internal, External, Hostname<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.spec.assignment
<sup><sup>[↩ Parent](#nodespec)</sup></sup>



Assignment optionally records the Cluster this Node is a member of.
When unset the Node is considered unassigned.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#nodespecassignmentclusterref">clusterRef</a></b></td>
        <td>object</td>
        <td>
          ClusterRef references the Cluster the Node is assigned to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>enum</td>
        <td>
          Role is the Node's role within the referenced Cluster.<br/>
          <br/>
            <i>Enum</i>: ControlPlane, Worker<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.spec.assignment.clusterRef
<sup><sup>[↩ Parent](#nodespecassignment)</sup></sup>



ClusterRef references the Cluster the Node is assigned to.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.spec.placement
<sup><sup>[↩ Parent](#nodespec)</sup></sup>



Placement optionally records where the Node is physically mounted in a
Rack. A validating webhook enforces fit and non-overlap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#nodespecplacementrackref">rackRef</a></b></td>
        <td>object</td>
        <td>
          RackRef references the Rack this device is mounted in.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>startUnit</b></td>
        <td>integer</td>
        <td>
          StartUnit is the lowest rack unit (U) the device occupies, counting from
1 at the bottom of the rack.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>unitHeight</b></td>
        <td>integer</td>
        <td>
          UnitHeight is the number of contiguous units the device occupies.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>face</b></td>
        <td>enum</td>
        <td>
          Face is the rack face the device is mounted on. Defaults to Front.<br/>
          <br/>
            <i>Enum</i>: Front, Rear<br/>
            <i>Default</i>: Front<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Node.spec.placement.rackRef
<sup><sup>[↩ Parent](#nodespecplacement)</sup></sup>



RackRef references the Rack this device is mounted in.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Node.status
<sup><sup>[↩ Parent](#node)</sup></sup>



NodeStatus defines the observed state of Node.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#nodestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a node's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>phase</b></td>
        <td>enum</td>
        <td>
          Phase is a coarse lifecycle indicator derived from the Node's
assignment and readiness.<br/>
          <br/>
            <i>Enum</i>: Unassigned, Assigned, Unavailable<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Node.status.conditions[index]
<sup><sup>[↩ Parent](#nodestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Port
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Port is the Schema for the ports API. A Port is a connection point exposed
by a Node, NetworkDevice, or Rack-mounted device, and is the near/far-end
identifier a Cable connects.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Port</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#portspec">spec</a></b></td>
        <td>object</td>
        <td>
          PortSpec defines the desired state of Port.<br/>
          <br/>
            <i>Validations</i>:<li>self.deviceRef == oldSelf.deviceRef: deviceRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#portstatus">status</a></b></td>
        <td>object</td>
        <td>
          PortStatus defines the observed state of Port.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Port.spec
<sup><sup>[↩ Parent](#port)</sup></sup>



PortSpec defines the desired state of Port.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#portspecdeviceref">deviceRef</a></b></td>
        <td>object</td>
        <td>
          DeviceRef references the device that exposes this Port. This field is
immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name identifies the Port on its device (e.g. "eth0", "PSU1",
"pp-A-24").<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type classifies the Port.<br/>
          <br/>
            <i>Enum</i>: Power, Serial, Ethernet, Optical, PatchPanel<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>speed</b></td>
        <td>int or string</td>
        <td>
          Speed is the nominal line rate of a network Port, expressed as a
Kubernetes Quantity (e.g. "10G", "400G"). Unset for non-network ports.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Port.spec.deviceRef
<sup><sup>[↩ Parent](#portspec)</sup></sup>



DeviceRef references the device that exposes this Port. This field is
immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          Kind of the device exposing the port.<br/>
          <br/>
            <i>Enum</i>: Node, NetworkDevice, Rack<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced device.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Port.status
<sup><sup>[↩ Parent](#port)</sup></sup>



PortStatus defines the observed state of Port.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#portstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a port's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Port.status.conditions[index]
<sup><sup>[↩ Parent](#portstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Provider
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>








<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Provider</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#providerspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#providerstatus">status</a></b></td>
        <td>object</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Provider.spec
<sup><sup>[↩ Parent](#provider)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>displayName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: Hosting, Colocation, Transit, InternetExchange, DarkFiber, Cloud<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#providerspeccontract">contract</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#providerspecserviceidentifiersindex">serviceIdentifiers</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Provider.spec.contract
<sup><sup>[↩ Parent](#providerspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accountID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contractID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>notes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portalURL</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Provider.spec.serviceIdentifiers[index]
<sup><sup>[↩ Parent](#providerspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>identifier</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Provider.status
<sup><sup>[↩ Parent](#provider)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#providerstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Provider.status.conditions[index]
<sup><sup>[↩ Parent](#providerstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Rack
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Rack is the Schema for the racks API. A Rack belongs to exactly one Site and
is the mount point for placed Nodes and NetworkDevices.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Rack</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#rackspec">spec</a></b></td>
        <td>object</td>
        <td>
          RackSpec defines the desired state of Rack. Cage and row are modeled as
free-form attributes rather than their own kinds to avoid kind explosion.<br/>
          <br/>
            <i>Validations</i>:<li>self.siteRef == oldSelf.siteRef: siteRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#rackstatus">status</a></b></td>
        <td>object</td>
        <td>
          RackStatus defines the observed state of Rack.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Rack.spec
<sup><sup>[↩ Parent](#rack)</sup></sup>



RackSpec defines the desired state of Rack. Cage and row are modeled as
free-form attributes rather than their own kinds to avoid kind explosion.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>heightU</b></td>
        <td>integer</td>
        <td>
          HeightU is the number of mountable rack units (U) the rack provides.
Devices placed in the rack occupy unit positions 1..HeightU.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#rackspecsiteref">siteRef</a></b></td>
        <td>object</td>
        <td>
          SiteRef references the Site this Rack physically lives in. This field is
immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cage</b></td>
        <td>string</td>
        <td>
          Cage is the free-form cage identifier within the Site.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a human-readable name for the rack (e.g. its asset label).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#rackspecpowerfeedsindex">powerFeeds</a></b></td>
        <td>[]object</td>
        <td>
          PowerFeeds enumerates the power feeds delivered to the rack.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>row</b></td>
        <td>string</td>
        <td>
          Row is the free-form row identifier within the cage.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Rack.spec.siteRef
<sup><sup>[↩ Parent](#rackspec)</sup></sup>



SiteRef references the Site this Rack physically lives in. This field is
immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Rack.spec.powerFeeds[index]
<sup><sup>[↩ Parent](#rackspec)</sup></sup>



RackPowerFeed describes a single power feed delivered to a Rack.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name identifies the feed (e.g. "A", "B").<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>ampsRated</b></td>
        <td>integer</td>
        <td>
          AmpsRated is the rated current of the feed in amperes.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>phase</b></td>
        <td>enum</td>
        <td>
          Phase is the electrical phase configuration of the feed.<br/>
          <br/>
            <i>Enum</i>: SinglePhase, ThreePhase<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>voltage</b></td>
        <td>integer</td>
        <td>
          Voltage is the nominal feed voltage in volts.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Rack.status
<sup><sup>[↩ Parent](#rack)</sup></sup>



RackStatus defines the observed state of Rack.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#rackstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a rack's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Rack.status.conditions[index]
<sup><sup>[↩ Parent](#rackstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Region
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Region is the Schema for the regions API. A Region is the top-level
geographic grouping in the inventory and has no parent references.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Region</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#regionspec">spec</a></b></td>
        <td>object</td>
        <td>
          RegionSpec defines the desired state of Region.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#regionstatus">status</a></b></td>
        <td>object</td>
        <td>
          RegionStatus defines the observed state of Region.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Region.spec
<sup><sup>[↩ Parent](#region)</sup></sup>



RegionSpec defines the desired state of Region.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>displayName</b></td>
        <td>string</td>
        <td>
          DisplayName is a human-readable name for the region (e.g. "US East").<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#regionspeccoordinates">coordinates</a></b></td>
        <td>object</td>
        <td>
          Coordinates optionally records a representative latitude/longitude for
the region. This is descriptive only and not used for routing.<br/>
          <br/>
            <i>Validations</i>:<li>self.latitude >= -90.0 && self.latitude <= 90.0: latitude must be between -90 and 90 degrees</li><li>self.longitude >= -180.0 && self.longitude <= 180.0: longitude must be between -180 and 180 degrees</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Region.spec.coordinates
<sup><sup>[↩ Parent](#regionspec)</sup></sup>



Coordinates optionally records a representative latitude/longitude for
the region. This is descriptive only and not used for routing.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>latitude</b></td>
        <td>number</td>
        <td>
          Latitude in decimal degrees. Must be between -90 and 90 inclusive.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>longitude</b></td>
        <td>number</td>
        <td>
          Longitude in decimal degrees. Must be between -180 and 180 inclusive.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Region.status
<sup><sup>[↩ Parent](#region)</sup></sup>



RegionStatus defines the observed state of Region.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#regionstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a region's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Region.status.conditions[index]
<sup><sup>[↩ Parent](#regionstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Site
<sup><sup>[↩ Parent](#inventorymiloapiscomv1alpha1 )</sup></sup>






Site is the Schema for the sites API. A Site belongs to exactly one Region
and is the parent of Nodes, Clusters, and NetworkDevices.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>inventory.miloapis.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Site</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sitespec">spec</a></b></td>
        <td>object</td>
        <td>
          SiteSpec defines the desired state of Site.<br/>
          <br/>
            <i>Validations</i>:<li>self.regionRef == oldSelf.regionRef: regionRef is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#sitestatus">status</a></b></td>
        <td>object</td>
        <td>
          SiteStatus defines the observed state of Site.<br/>
          <br/>
            <i>Default</i>: map[conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for reconciliation reason:Pending status:False type:Ready]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Site.spec
<sup><sup>[↩ Parent](#site)</sup></sup>



SiteSpec defines the desired state of Site.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>displayName</b></td>
        <td>string</td>
        <td>
          DisplayName is a human-readable name for the site.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#sitespecregionref">regionRef</a></b></td>
        <td>object</td>
        <td>
          RegionRef references the Region this Site belongs to. This field is
immutable after creation.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type classifies the Site.<br/>
          <br/>
            <i>Enum</i>: Datacenter, AvailabilityZone, Edge, Virtual<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address is an optional free-form postal/street address for the site.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sitespecproviderref">providerRef</a></b></td>
        <td>object</td>
        <td>
          LocalObjectReference is a reference to another cluster-scoped object in the
inventory.miloapis.com API group by name. All inventory cross-resource
references use this type rather than raw strings so that CRD schemas are
explicit about intent.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Site.spec.regionRef
<sup><sup>[↩ Parent](#sitespec)</sup></sup>



RegionRef references the Region this Site belongs to. This field is
immutable after creation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Site.spec.providerRef
<sup><sup>[↩ Parent](#sitespec)</sup></sup>



LocalObjectReference is a reference to another cluster-scoped object in the
inventory.miloapis.com API group by name. All inventory cross-resource
references use this type rather than raw strings so that CRD schemas are
explicit about intent.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referenced object.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Site.status
<sup><sup>[↩ Parent](#site)</sup></sup>



SiteStatus defines the observed state of Site.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#sitestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Represents the observations of a site's current state.
Known condition types are: "Ready".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Site.status.conditions[index]
<sup><sup>[↩ Parent](#sitestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
