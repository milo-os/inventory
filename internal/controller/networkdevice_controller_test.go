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

var _ = Describe("NetworkDevice Controller", func() {
	var (
		regionName    string
		siteName      string
		clusterName   string
		deviceName    string
		otherSiteName string
	)

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		otherSiteName = uniqueName("site")
		clusterName = uniqueName("cluster")
		deviceName = uniqueName("netdev")
	})

	AfterEach(func() {
		deleteNetworkDevice(deviceName)
		deleteCluster(clusterName)
		deleteSite(siteName)
		deleteSite(otherSiteName)
		deleteRegion(regionName)
	})

	It("becomes Ready with topology labels propagated from its Cluster", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute))).To(Succeed())

		// Wait for the Cluster to have all its topology labels.
		Eventually(func(g Gomega) {
			var c inventoryv1alpha1.Cluster
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: clusterName}, &c)).To(Succeed())
			g.Expect(c.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(c.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyClusterLabel, clusterName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		device := makeNetworkDevice(deviceName, siteName, clusterName)
		Expect(k8sClient.Create(testCtx, device)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.NetworkDevice
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(device), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NetworkDeviceReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))

			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyClusterLabel, clusterName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is accepted when its siteRef differs from the Cluster's controlPlaneSiteRef (clusters span sites)", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(otherSiteName, regionName, inventoryv1alpha1.SiteTypeEdge))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute))).To(Succeed())

		device := makeNetworkDevice(deviceName, otherSiteName, clusterName)
		Expect(k8sClient.Create(testCtx, device)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.NetworkDevice
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(device), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.NetworkDeviceReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is rejected at admission when the referenced Cluster does not exist", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		device := makeNetworkDevice(deviceName, siteName, clusterName)
		err := k8sClient.Create(testCtx, device)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not found"))
	})
})
