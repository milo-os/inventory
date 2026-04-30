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

var _ = Describe("Site Controller", func() {
	var (
		regionName string
		siteName   string
	)

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
	})

	AfterEach(func() {
		deleteSite(siteName)
		deleteRegion(regionName)
	})

	It("becomes Ready and gets topology labels when regionRef resolves", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		site := makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter)
		Expect(k8sClient.Create(testCtx, site)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.SiteReadyReason))

			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteTypeLabel, string(inventoryv1alpha1.SiteTypeDatacenter)))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("reports RegionNotFound when the regionRef is dangling, then recovers once the Region is created", func() {
		site := makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeEdge)
		Expect(k8sClient.Create(testCtx, site)).To(Succeed())

		By("observing Ready=False with reason RegionNotFound")
		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.SiteRegionNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		By("creating the missing Region")
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())

		By("observing Site transitions to Ready=True via the Region watch")
		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologyRegionLabel, regionName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects Region DELETE while a referencing Site still exists", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		site := makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter)
		Expect(k8sClient.Create(testCtx, site)).To(Succeed())

		// Wait for the Site controller to observe the ref so that the index is populated.
		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		region := &inventoryv1alpha1.Region{}
		region.Name = regionName
		err := k8sClient.Delete(testCtx, region)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot delete"))
		Expect(err.Error()).To(ContainSubstring(siteName))
	})
})
