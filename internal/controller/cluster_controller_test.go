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

var _ = Describe("Cluster Controller", func() {
	var (
		regionName  string
		siteName    string
		clusterName string
	)

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		clusterName = uniqueName("cluster")
	})

	AfterEach(func() {
		deleteCluster(clusterName)
		deleteSite(siteName)
		deleteRegion(regionName)
	})

	It("becomes Ready and carries region/site/site-type/cluster labels when siteRef resolves", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		// Wait for Site to have its topology labels so the Cluster can inherit them.
		Eventually(func(g Gomega) {
			var s inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: siteName}, &s)).To(Succeed())
			g.Expect(s.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		cluster := makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute)
		Expect(k8sClient.Create(testCtx, cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Cluster
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(cluster), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.ClusterReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.ClusterReadyReason))

			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteTypeLabel, string(inventoryv1alpha1.SiteTypeDatacenter)))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyClusterLabel, clusterName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("reports SiteNotFound when the siteRef is dangling", func() {
		cluster := makeCluster(clusterName, siteName, inventoryv1alpha1.ClusterRoleCompute)
		Expect(k8sClient.Create(testCtx, cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Cluster
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(cluster), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.ClusterReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.ClusterSiteNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})
