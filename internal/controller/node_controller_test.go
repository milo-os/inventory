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

var _ = Describe("Node Controller", func() {
	var (
		regionName  string
		siteName    string
		clusterName string
		nodeName    string
	)

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		clusterName = uniqueName("cluster")
		nodeName = uniqueName("node")
	})

	AfterEach(func() {
		deleteNode(nodeName)
		deleteCluster(clusterName)
		deleteSite(siteName)
		deleteRegion(regionName)
	})

	It("marks an unassigned Node Ready=True, Phase=Unassigned, with site labels but no cluster label", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		// Ensure the Site has region label before creating the Node so that the
		// Node's first reconciliation inherits it.
		Eventually(func(g Gomega) {
			var s inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: siteName}, &s)).To(Succeed())
			g.Expect(s.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		node := makeNode(nodeName, siteName)
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			g.Expect(fetched.Status.Phase).To(Equal(inventoryv1alpha1.NodePhaseUnassigned))

			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NodeReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))

			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteTypeLabel, string(inventoryv1alpha1.SiteTypeDatacenter)))
			g.Expect(fetched.Labels).NotTo(HaveKey(inventoryv1alpha1.TopologyClusterLabel))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("transitions to Phase=Assigned with a cluster label after the assignment is set", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute))).To(Succeed())

		// Wait for the Cluster to inherit all its topology labels.
		Eventually(func(g Gomega) {
			var c inventoryv1alpha1.Cluster
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: clusterName}, &c)).To(Succeed())
			g.Expect(c.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(c.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyClusterLabel, clusterName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		node := makeNode(nodeName, siteName)
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			g.Expect(fetched.Status.Phase).To(Equal(inventoryv1alpha1.NodePhaseUnassigned))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		By("patching the Node's assignment to point at the Cluster")
		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			original := fetched.DeepCopy()
			fetched.Spec.Assignment = &inventoryv1alpha1.NodeAssignment{
				ClusterRef: inventoryv1alpha1.LocalObjectReference{Name: clusterName},
				Role:       inventoryv1alpha1.NodeRoleWorker,
			}
			g.Expect(k8sClient.Patch(testCtx, &fetched, client.MergeFrom(original))).To(Succeed())
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			g.Expect(fetched.Status.Phase).To(Equal(inventoryv1alpha1.NodePhaseAssigned))

			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NodeReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))

			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyClusterLabel, clusterName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("reports SiteNotFound with Phase=Unavailable when siteRef is dangling", func() {
		node := makeNode(nodeName, siteName)
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NodeReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.NodeSiteNotFoundReason))
			g.Expect(fetched.Status.Phase).To(Equal(inventoryv1alpha1.NodePhaseUnavailable))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("reports ClusterNotFound with Phase=Unavailable when assignment points at a missing Cluster", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		node := makeNode(nodeName, siteName)
		node.Spec.Assignment = &inventoryv1alpha1.NodeAssignment{
			ClusterRef: inventoryv1alpha1.LocalObjectReference{Name: clusterName},
			Role:       inventoryv1alpha1.NodeRoleWorker,
		}
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Node
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(node), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NodeReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.NodeClusterNotFoundReason))
			g.Expect(fetched.Status.Phase).To(Equal(inventoryv1alpha1.NodePhaseUnavailable))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})
