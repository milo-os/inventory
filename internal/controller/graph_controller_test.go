// SPDX-License-Identifier: AGPL-3.0-only

package controller_test

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha2 "go.miloapis.com/inventory/api/v1alpha2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func deleteGraphNode(name string) {
	n := &inventoryv1alpha2.Node{}
	n.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, n))).To(Succeed())
}

func deleteEdge(name string) {
	e := &inventoryv1alpha2.Edge{}
	e.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, e))).To(Succeed())
}

func deleteNodeType(name string) {
	nt := &inventoryv1alpha2.NodeType{}
	nt.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, nt))).To(Succeed())
}

func deleteEdgeType(name string) {
	et := &inventoryv1alpha2.EdgeType{}
	et.Name = name
	Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, et))).To(Succeed())
}

func makeSiteNodeType(name string) *inventoryv1alpha2.NodeType {
	return &inventoryv1alpha2.NodeType{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha2.NodeTypeSpec{
			Attributes: []inventoryv1alpha2.AttributeSchema{
				{Key: "displayName", Type: inventoryv1alpha2.AttributeString, Required: true},
				{Key: "siteType", Type: inventoryv1alpha2.AttributeString, Required: true,
					Enum: []string{"Datacenter", "Edge"}},
				{Key: "cpuCores", Type: inventoryv1alpha2.AttributeInteger},
			},
		},
	}
}

func makeHostNodeType(name string) *inventoryv1alpha2.NodeType {
	return &inventoryv1alpha2.NodeType{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha2.NodeTypeSpec{
			Attributes: []inventoryv1alpha2.AttributeSchema{
				{Key: "cpuCores", Type: inventoryv1alpha2.AttributeInteger, Required: true},
			},
		},
	}
}

func makeLocatedInEdgeType(name, fromType, toType string) *inventoryv1alpha2.EdgeType {
	return &inventoryv1alpha2.EdgeType{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha2.EdgeTypeSpec{
			Endpoints: inventoryv1alpha2.EndpointConstraint{
				FromTypes: []string{fromType},
				ToTypes:   []string{toType},
			},
		},
	}
}

func makeGraphNode(name, nodeType string, attrs map[string]string) *inventoryv1alpha2.Node {
	return &inventoryv1alpha2.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       inventoryv1alpha2.NodeSpec{Type: nodeType, Attributes: attrs},
	}
}

func makeEdge(name, edgeType, from, to string) *inventoryv1alpha2.Edge {
	return &inventoryv1alpha2.Edge{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: inventoryv1alpha2.EdgeSpec{
			Type: edgeType,
			From: inventoryv1alpha2.NodeReference{Name: from},
			To:   inventoryv1alpha2.NodeReference{Name: to},
		},
	}
}

var _ = Describe("Graph Node Controller", func() {
	var siteTypeName, hostTypeName, nodeName string

	BeforeEach(func() {
		siteTypeName = uniqueName("site")
		hostTypeName = uniqueName("host")
		nodeName = uniqueName("node")
	})

	AfterEach(func() {
		deleteGraphNode(nodeName)
		deleteNodeType(siteTypeName)
		deleteNodeType(hostTypeName)
	})

	It("becomes Ready when its NodeType exists and attributes are valid", func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())

		node := makeGraphNode(nodeName, siteTypeName, map[string]string{
			"displayName": "US East 1",
			"siteType":    "Datacenter",
		})
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha2.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha2.NodeReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is rejected at admission when the NodeType does not exist", func() {
		node := makeGraphNode(nodeName, siteTypeName, map[string]string{"displayName": "x", "siteType": "Edge"})
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("NodeType"))
		Expect(err.Error()).To(ContainSubstring("not found"))
	})

	It("is rejected when an attribute is unknown", func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())
		node := makeGraphNode(nodeName, siteTypeName, map[string]string{
			"displayName": "x", "siteType": "Edge", "bogus": "1",
		})
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not allow attribute"))
	})

	It("is rejected when a required attribute is missing", func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())
		node := makeGraphNode(nodeName, siteTypeName, map[string]string{"displayName": "x"})
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("requires attribute"))
	})

	It("is rejected when an Integer attribute does not parse", func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())
		node := makeGraphNode(nodeName, siteTypeName, map[string]string{
			"displayName": "x", "siteType": "Edge", "cpuCores": "lots",
		})
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("must be an integer"))
	})

	It("is rejected when a String attribute is outside its enum", func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())
		node := makeGraphNode(nodeName, siteTypeName, map[string]string{
			"displayName": "x", "siteType": "Orbital",
		})
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("must be one of"))
	})
})

var _ = Describe("Graph Edge Controller", func() {
	var siteTypeName, hostTypeName, edgeTypeName string
	var siteNode, hostNode, edgeName string

	BeforeEach(func() {
		siteTypeName = uniqueName("site")
		hostTypeName = uniqueName("host")
		edgeTypeName = uniqueName("locatedin")
		siteNode = uniqueName("site")
		hostNode = uniqueName("host")
		edgeName = uniqueName("edge")
	})

	AfterEach(func() {
		deleteEdge(edgeName)
		deleteGraphNode(hostNode)
		deleteGraphNode(siteNode)
		deleteEdgeType(edgeTypeName)
		deleteNodeType(siteTypeName)
		deleteNodeType(hostTypeName)
	})

	setup := func() {
		Expect(k8sClient.Create(testCtx, makeSiteNodeType(siteTypeName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeHostNodeType(hostTypeName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeLocatedInEdgeType(edgeTypeName, hostTypeName, siteTypeName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeGraphNode(siteNode, siteTypeName, map[string]string{
			"displayName": "US East 1", "siteType": "Datacenter",
		}))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeGraphNode(hostNode, hostTypeName, map[string]string{
			"cpuCores": "64",
		}))).To(Succeed())
	}

	It("becomes Ready with EndpointsResolved=True when both endpoints exist", func() {
		setup()
		edge := makeEdge(edgeName, edgeTypeName, hostNode, siteNode)
		Expect(k8sClient.Create(testCtx, edge)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha2.Edge
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(edge), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha2.EdgeReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))
			resolved := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha2.EdgeEndpointsResolved)
			g.Expect(resolved).NotTo(BeNil())
			g.Expect(resolved.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is rejected at admission when an endpoint Node does not exist", func() {
		setup()
		edge := makeEdge(edgeName, edgeTypeName, hostNode, uniqueName("ghost"))
		err := k8sClient.Create(testCtx, edge)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not found"))
	})

	It("is rejected when an endpoint violates the EdgeType endpoint-type constraint", func() {
		setup()
		// from/to swapped: a Site cannot be the `from` of this located-in type.
		edge := makeEdge(edgeName, edgeTypeName, siteNode, hostNode)
		err := k8sClient.Create(testCtx, edge)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not allow"))
	})

	It("blocks deletion of a Node while an Edge still references it", func() {
		setup()
		edge := makeEdge(edgeName, edgeTypeName, hostNode, siteNode)
		Expect(k8sClient.Create(testCtx, edge)).To(Succeed())

		n := &inventoryv1alpha2.Node{}
		n.Name = siteNode
		err := k8sClient.Delete(testCtx, n)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("still reference it"))
	})
})
