// SPDX-License-Identifier: AGPL-3.0-only

package controller_test

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apirand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"

	. "github.com/onsi/gomega"
)

const (
	// defaultTimeout is the default timeout for Eventually assertions.
	defaultTimeout = 10 * time.Second

	// defaultInterval is the default polling interval for Eventually.
	defaultInterval = 250 * time.Millisecond
)

// uniqueName returns a random lowercase suffix unique within the suite run,
// prefixed by kind. All inventory objects are cluster-scoped, so tests must
// generate unique names to avoid collisions across specs.
func uniqueName(kind string) string {
	return fmt.Sprintf("%s-%s", kind, apirand.String(8))
}

// deleteRegion deletes a Region ignoring NotFound.
func deleteRegion(name string) {
	r := &inventoryv1alpha1.Region{}
	r.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, r))).To(Succeed())
}

// deleteSite deletes a Site ignoring NotFound.
func deleteSite(name string) {
	s := &inventoryv1alpha1.Site{}
	s.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, s))).To(Succeed())
}

func deleteProvider(name string) {
	p := &inventoryv1alpha1.Provider{}
	p.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, p))).To(Succeed())
}

// deleteRack deletes a Rack ignoring NotFound.
func deleteRack(name string) {
	rk := &inventoryv1alpha1.Rack{}
	rk.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, rk))).To(Succeed())
}

// deleteCluster deletes a Cluster ignoring NotFound.
func deleteCluster(name string) {
	c := &inventoryv1alpha1.Cluster{}
	c.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, c))).To(Succeed())
}

// deleteNode deletes a Node ignoring NotFound.
func deleteNode(name string) {
	n := &inventoryv1alpha1.Node{}
	n.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, n))).To(Succeed())
}

// deleteNetworkDevice deletes a NetworkDevice ignoring NotFound.
func deleteNetworkDevice(name string) {
	d := &inventoryv1alpha1.NetworkDevice{}
	d.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, d))).To(Succeed())
}

// deleteLink deletes a Link ignoring NotFound.
func deleteLink(name string) {
	l := &inventoryv1alpha1.Link{}
	l.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, l))).To(Succeed())
}

// deletePort deletes a Port ignoring NotFound.
func deletePort(name string) {
	p := &inventoryv1alpha1.Port{}
	p.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, p))).To(Succeed())
}

// deleteCable deletes a Cable ignoring NotFound.
func deleteCable(name string) {
	c := &inventoryv1alpha1.Cable{}
	c.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, c))).To(Succeed())
}

// deleteCircuit deletes a Circuit ignoring NotFound.
func deleteCircuit(name string) {
	c := &inventoryv1alpha1.Circuit{}
	c.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, c))).To(Succeed())
}

// makeRegion returns a Region with the given name that is ready to Create.
func makeRegion(name string) *inventoryv1alpha1.Region {
	return &inventoryv1alpha1.Region{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.RegionSpec{
			DisplayName: "Test Region " + name,
		},
	}
}

// makeSite returns a Site referencing the given region.
func makeSite(name, regionName string, siteType inventoryv1alpha1.SiteType) *inventoryv1alpha1.Site {
	return &inventoryv1alpha1.Site{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.SiteSpec{
			DisplayName: "Test Site " + name,
			RegionRef:   inventoryv1alpha1.LocalObjectReference{Name: regionName},
			Type:        siteType,
		},
	}
}

func makeProvider(name string) *inventoryv1alpha1.Provider {
	return &inventoryv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.ProviderSpec{
			DisplayName: "Test Provider " + name,
			Type:        inventoryv1alpha1.ProviderTypeColocation,
		},
	}
}

// makeRack returns a Rack referencing the given site with the given height.
func makeRack(name, siteName string, heightU int32) *inventoryv1alpha1.Rack {
	return &inventoryv1alpha1.Rack{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.RackSpec{
			SiteRef: inventoryv1alpha1.LocalObjectReference{Name: siteName},
			HeightU: heightU,
		},
	}
}

// makeCluster returns a Cluster whose control plane lives at the given site.
func makeCluster(name, siteName string, role inventoryv1alpha1.ClusterRole) *inventoryv1alpha1.Cluster {
	return &inventoryv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.ClusterSpec{
			DisplayName:         "Test Cluster " + name,
			ControlPlaneSiteRef: inventoryv1alpha1.LocalObjectReference{Name: siteName},
			Role:                role,
			Provider:            "test-provider",
		},
	}
}

// makeNode returns a Node with a minimal hardware spec referencing the given
// site. Assignment is left unset.
func makeNode(name, siteName string) *inventoryv1alpha1.Node {
	return &inventoryv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.NodeSpec{
			SiteRef: inventoryv1alpha1.LocalObjectReference{Name: siteName},
			Hardware: inventoryv1alpha1.NodeHardware{
				CPUCores:        8,
				CPUArchitecture: inventoryv1alpha1.CPUArchitectureAMD64,
				MemoryBytes:     32 * 1024 * 1024 * 1024,
			},
		},
	}
}

// makeNetworkDevice returns a NetworkDevice referencing the given site and
// cluster.
func makeNetworkDevice(name, siteName, clusterName string) *inventoryv1alpha1.NetworkDevice {
	return &inventoryv1alpha1.NetworkDevice{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.NetworkDeviceSpec{
			ClusterRef: inventoryv1alpha1.LocalObjectReference{Name: clusterName},
			SiteRef:    inventoryv1alpha1.LocalObjectReference{Name: siteName},
			Role:       inventoryv1alpha1.NetworkDeviceRoleLeaf,
		},
	}
}

// makePort returns a Port on the given device.
func makePort(name string, deviceKind, deviceName string, portType inventoryv1alpha1.PortType) *inventoryv1alpha1.Port {
	return &inventoryv1alpha1.Port{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.PortSpec{
			DeviceRef: inventoryv1alpha1.PortDeviceReference{Kind: deviceKind, Name: deviceName},
			Type:      portType,
			Name:      name,
		},
	}
}

// makeCable returns a Cable connecting the two named Ports.
func makeCable(name, portA, portB string, media inventoryv1alpha1.CableMedia) *inventoryv1alpha1.Cable {
	return &inventoryv1alpha1.Cable{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.CableSpec{
			Endpoints: []inventoryv1alpha1.LocalObjectReference{
				{Name: portA},
				{Name: portB},
			},
			Media: media,
		},
	}
}

// makeCircuit returns a Circuit delivered by the given Provider with both ends
// terminating at the named Sites.
func makeCircuit(name, providerName, siteA, siteZ string) *inventoryv1alpha1.Circuit {
	return &inventoryv1alpha1.Circuit{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.CircuitSpec{
			ProviderRef: inventoryv1alpha1.LocalObjectReference{Name: providerName},
			Type:        inventoryv1alpha1.CircuitTypeProviderCircuit,
			AEnd:        inventoryv1alpha1.CircuitEndpoint{Kind: inventoryv1alpha1.CircuitEndpointKindSite, Name: siteA},
			ZEnd:        inventoryv1alpha1.CircuitEndpoint{Kind: inventoryv1alpha1.CircuitEndpointKindSite, Name: siteZ},
		},
	}
}

// makeLink returns a Link whose endpoints both reference Sites with the given
// names.
func makeSiteToSiteLink(name, siteA, siteB string) *inventoryv1alpha1.Link {
	return &inventoryv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha1.LinkSpec{
			Endpoints: []inventoryv1alpha1.AssetReference{
				{Kind: "Site", Name: siteA},
				{Kind: "Site", Name: siteB},
			},
			Type: inventoryv1alpha1.LinkTypeLogical,
		},
	}
}
