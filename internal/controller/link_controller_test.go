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

var _ = Describe("Link Controller", func() {
	var (
		regionName string
		siteAName  string
		siteBName  string
		linkName   string
	)

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteAName = uniqueName("site")
		siteBName = uniqueName("site")
		linkName = uniqueName("link")
	})

	AfterEach(func() {
		deleteLink(linkName)
		deleteSite(siteAName)
		deleteSite(siteBName)
		deleteRegion(regionName)
	})

	It("becomes Ready with EndpointsResolved=True when both endpoints exist", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteAName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteBName, regionName, inventoryv1alpha1.SiteTypeEdge))).To(Succeed())

		link := makeSiteToSiteLink(linkName, siteAName, siteBName)
		Expect(k8sClient.Create(testCtx, link)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Link
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(link), &fetched)).To(Succeed())

			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.LinkReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))

			resolved := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.LinkEndpointsResolved)
			g.Expect(resolved).NotTo(BeNil())
			g.Expect(resolved.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("is rejected at admission when an endpoint does not exist", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteAName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())

		link := makeSiteToSiteLink(linkName, siteAName, siteBName)
		err := k8sClient.Create(testCtx, link)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not found"))
		Expect(err.Error()).To(ContainSubstring(siteBName))
	})

	It("transitions Ready to False with reason EndpointNotFound when an endpoint is deleted", func() {
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteAName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteBName, regionName, inventoryv1alpha1.SiteTypeEdge))).To(Succeed())

		link := makeSiteToSiteLink(linkName, siteAName, siteBName)
		Expect(k8sClient.Create(testCtx, link)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Link
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(link), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.LinkReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		By("deleting one of the Link endpoints")
		deleteSite(siteBName)

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Link
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(link), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.LinkReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(ready.Reason).To(Equal(inventoryv1alpha1.LinkEndpointNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})
