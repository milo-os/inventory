// SPDX-License-Identifier: AGPL-3.0-only

package controller_test

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// deleteEventually deletes obj retrying until a delete-guard webhook stops
// rejecting it (the guard reads the manager cache, which lags behind the
// deletes of the referencing objects).
func deleteEventually(obj client.Object) {
	Eventually(func(g Gomega) {
		g.Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, obj))).To(Succeed())
	}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
}

var _ = Describe("Port Controller", func() {
	var regionName, siteName, nodeName, portName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		nodeName = uniqueName("node")
		portName = uniqueName("port")
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeNode(nodeName, siteName))).To(Succeed())

		DeferCleanup(func() { deleteRegion(regionName) })
		DeferCleanup(func() { deleteSite(siteName) })
		DeferCleanup(func() { deleteNode(nodeName) })
		DeferCleanup(func() { deletePort(portName) })
	})

	It("becomes Ready=True and inherits the device's topology labels", func() {
		port := makePort(portName, "Node", nodeName, inventoryv1alpha1.PortTypeEthernet)
		Expect(k8sClient.Create(testCtx, port)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Port
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(port), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.PortReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.PortReadyReason))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is NotReady with DeviceNotFound while the device is absent", func() {
		port := makePort(portName, "Node", "ghost-node", inventoryv1alpha1.PortTypeEthernet)
		Expect(k8sClient.Create(testCtx, port)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Port
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(port), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.PortReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.PortDeviceNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects deviceRef mutation", func() {
		Expect(k8sClient.Create(testCtx, makePort(portName, "Node", nodeName, inventoryv1alpha1.PortTypeEthernet))).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Port
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: portName}, &fetched)).To(Succeed())
			fetched.Spec.DeviceRef.Name = "another-node"
			err := k8sClient.Update(testCtx, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("deviceRef is immutable"))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})

var _ = Describe("Cable Controller", func() {
	var regionName, siteName, nodeName, portA, portB string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		nodeName = uniqueName("node")
		portA = uniqueName("port")
		portB = uniqueName("port")
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeNode(nodeName, siteName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makePort(portA, "Node", nodeName, inventoryv1alpha1.PortTypeOptical))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makePort(portB, "Node", nodeName, inventoryv1alpha1.PortTypeOptical))).To(Succeed())

		// LIFO cleanup: per-It Cable/Link deletes (registered later) run first,
		// then the Ports (delete-guarded by Cable), then the Node/Site/Region.
		DeferCleanup(func() { deleteRegion(regionName) })
		DeferCleanup(func() { deleteSite(siteName) })
		DeferCleanup(func() { deleteNode(nodeName) })
		DeferCleanup(func() { deleteEventually(makePort(portA, "Node", nodeName, inventoryv1alpha1.PortTypeOptical)) })
		DeferCleanup(func() { deleteEventually(makePort(portB, "Node", nodeName, inventoryv1alpha1.PortTypeOptical)) })
	})

	It("becomes Ready=True with EndpointsResolved once both Ports exist", func() {
		cableName := uniqueName("cable")
		DeferCleanup(func() { deleteCable(cableName) })

		cable := makeCable(cableName, portA, portB, inventoryv1alpha1.CableMediaFiberSMF)
		Expect(k8sClient.Create(testCtx, cable)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Cable
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(cable), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.CableReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))
			resolved := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.CableEndpointsResolved)
			g.Expect(resolved).NotTo(BeNil())
			g.Expect(resolved.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects a Cable whose endpoint Port does not exist", func() {
		cable := makeCable(uniqueName("cable"), portA, "ghost-port", inventoryv1alpha1.CableMediaFiberSMF)
		err := k8sClient.Create(testCtx, cable)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ghost-port not found"))
	})

	It("rejects a Cable whose endpoints are not distinct", func() {
		cable := makeCable(uniqueName("cable"), portA, portA, inventoryv1alpha1.CableMediaFiberSMF)
		err := k8sClient.Create(testCtx, cable)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("endpoints must be distinct"))
	})

	It("rejects deletion of a Port while a Cable references it", func() {
		cableName := uniqueName("cable")
		DeferCleanup(func() { deleteCable(cableName) })
		Expect(k8sClient.Create(testCtx, makeCable(cableName, portA, portB, inventoryv1alpha1.CableMediaFiberSMF))).To(Succeed())

		port := &inventoryv1alpha1.Port{}
		port.Name = portA
		err := k8sClient.Delete(testCtx, port)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("still reference it"))
	})

	It("lets a Link record cableRefs and rejects deleting the referenced Cable", func() {
		cableName := uniqueName("cable")
		linkName := uniqueName("link")
		secondSite := uniqueName("site")

		// LIFO: the Link (guarding both the Cable and secondSite) must be
		// deleted first, so register it last. The guarded deletes retry to
		// absorb the manager-cache lag behind the Link delete.
		DeferCleanup(func() { deleteEventually(&inventoryv1alpha1.Site{ObjectMeta: metav1.ObjectMeta{Name: secondSite}}) })
		DeferCleanup(func() { deleteEventually(&inventoryv1alpha1.Cable{ObjectMeta: metav1.ObjectMeta{Name: cableName}}) })
		DeferCleanup(func() { deleteLink(linkName) })

		Expect(k8sClient.Create(testCtx, makeCable(cableName, portA, portB, inventoryv1alpha1.CableMediaFiberSMF))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(secondSite, regionName, inventoryv1alpha1.SiteTypeEdge))).To(Succeed())

		link := makeSiteToSiteLink(linkName, siteName, secondSite)
		link.Spec.CableRefs = []inventoryv1alpha1.LocalObjectReference{{Name: cableName}}
		Expect(k8sClient.Create(testCtx, link)).To(Succeed())

		cable := &inventoryv1alpha1.Cable{}
		cable.Name = cableName
		err := k8sClient.Delete(testCtx, cable)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("still reference it"))
	})
})
