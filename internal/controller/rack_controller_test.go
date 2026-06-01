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

var _ = Describe("Rack Controller", func() {
	var regionName, siteName, rackName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		rackName = uniqueName("rack")
	})

	AfterEach(func() {
		deleteRack(rackName)
		deleteSite(siteName)
		deleteRegion(regionName)
	})

	It("becomes Ready=True once its Site exists", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		rack := makeRack(rackName, siteName, 42)
		Expect(k8sClient.Create(testCtx, rack)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Rack
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(rack), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.RackReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.RackReadyReason))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is NotReady with SiteNotFound while the Site is absent, then Ready once created", func() {
		rack := makeRack(rackName, siteName, 42)
		Expect(k8sClient.Create(testCtx, rack)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Rack
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(rack), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.RackReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.RackSiteNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Rack
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(rack), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.RackReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects siteRef mutation", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeRack(rackName, siteName, 42))).To(Succeed())

		// Re-fetch each attempt: the controller patches labels/status
		// concurrently, so a stale Update races to an optimistic-concurrency
		// conflict before CEL validation runs.
		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Rack
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: rackName}, &fetched)).To(Succeed())
			fetched.Spec.SiteRef.Name = "some-other-site"
			err := k8sClient.Update(testCtx, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("siteRef is immutable"))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})

var _ = Describe("Placement", func() {
	var regionName, siteName, rackName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		rackName = uniqueName("rack")
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeRack(rackName, siteName, 42))).To(Succeed())

		// Cleanup is registered via DeferCleanup so that it runs LIFO: the
		// per-It placed-device deletes (registered later) run before these,
		// leaving the rack empty by the time it is deleted. The rack delete
		// still retries because the delete-guard webhook reads from the
		// manager cache, which lags behind the device deletes.
		DeferCleanup(func() { deleteRegion(regionName) })
		DeferCleanup(func() { deleteSite(siteName) })
		DeferCleanup(func() {
			Eventually(func(g Gomega) {
				rk := &inventoryv1alpha1.Rack{}
				rk.Name = rackName
				g.Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, rk))).To(Succeed())
			}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
		})
	})

	It("admits a Node whose placement fits and propagates the rack label", func() {
		nodeName := uniqueName("node")
		DeferCleanup(func() { deleteNode(nodeName) })

		node := makeNode(nodeName, siteName)
		node.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  10,
			UnitHeight: 2,
			Face:       inventoryv1alpha1.RackFaceFront,
		}
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRackLabel, rackName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects a placement referencing a nonexistent Rack", func() {
		nodeName := uniqueName("node")
		node := makeNode(nodeName, siteName)
		node.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: "ghost-rack"},
			StartUnit:  1,
			UnitHeight: 1,
		}
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Rack ghost-rack not found"))
	})

	It("rejects a placement that does not fit within the rack height", func() {
		nodeName := uniqueName("node")
		node := makeNode(nodeName, siteName)
		node.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  41,
			UnitHeight: 4,
		}
		err := k8sClient.Create(testCtx, node)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not fit"))
	})

	It("rejects a Node overlapping another device on the same face", func() {
		firstName := uniqueName("node")
		secondName := uniqueName("node")
		DeferCleanup(func() { deleteNode(firstName); deleteNode(secondName) })

		first := makeNode(firstName, siteName)
		first.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  10,
			UnitHeight: 3,
			Face:       inventoryv1alpha1.RackFaceFront,
		}
		Expect(k8sClient.Create(testCtx, first)).To(Succeed())

		second := makeNode(secondName, siteName)
		second.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  12,
			UnitHeight: 2,
			Face:       inventoryv1alpha1.RackFaceFront,
		}
		err := k8sClient.Create(testCtx, second)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("overlaps"))
	})

	It("admits a device at the same units on the opposite face", func() {
		nodeName := uniqueName("node")
		devName := uniqueName("networkdevice")
		clusterName := uniqueName("cluster")
		DeferCleanup(func() {
			deleteNetworkDevice(devName)
			deleteNode(nodeName)
			deleteCluster(clusterName)
		})

		Expect(k8sClient.Create(testCtx, makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute))).To(Succeed())

		node := makeNode(nodeName, siteName)
		node.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  20,
			UnitHeight: 2,
			Face:       inventoryv1alpha1.RackFaceFront,
		}
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		dev := makeNetworkDevice(devName, siteName, clusterName)
		dev.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  20,
			UnitHeight: 2,
			Face:       inventoryv1alpha1.RackFaceRear,
		}
		Expect(k8sClient.Create(testCtx, dev)).To(Succeed())
	})

	It("rejects deletion of a Rack while a device is placed in it", func() {
		nodeName := uniqueName("node")
		DeferCleanup(func() { deleteNode(nodeName) })

		node := makeNode(nodeName, siteName)
		node.Spec.Placement = &inventoryv1alpha1.Placement{
			RackRef:    inventoryv1alpha1.LocalObjectReference{Name: rackName},
			StartUnit:  1,
			UnitHeight: 1,
		}
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		rack := &inventoryv1alpha1.Rack{}
		rack.Name = rackName
		err := k8sClient.Delete(testCtx, rack)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("still placed in it"))
	})
})
