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

var _ = Describe("Provider Controller", func() {
	var providerName string

	BeforeEach(func() {
		providerName = uniqueName("provider")
	})

	AfterEach(func() {
		deleteProvider(providerName)
	})

	It("becomes Ready=True shortly after creation", func() {
		provider := makeProvider(providerName)
		Expect(k8sClient.Create(testCtx, provider)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Provider
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(provider), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.ProviderReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.ProviderReadyReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("allows deletion when no Site references it", func() {
		provider := makeProvider(providerName)
		Expect(k8sClient.Create(testCtx, provider)).To(Succeed())
		Expect(k8sClient.Delete(testCtx, provider)).To(Succeed())

		Eventually(func(g Gomega) {
			err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(provider), &inventoryv1alpha1.Provider{})
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
			g.Expect(err).To(HaveOccurred())
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects deletion while a Site references it", func() {
		regionName := uniqueName("region")
		siteName := uniqueName("site")
		// Region is cleaned up after the AfterEach deleteProvider runs; the
		// Site must be removed within this spec so that deleteProvider is no
		// longer blocked by the time AfterEach runs.
		DeferCleanup(func() { deleteRegion(regionName) })

		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		provider := makeProvider(providerName)
		Expect(k8sClient.Create(testCtx, provider)).To(Succeed())

		site := makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter)
		site.Spec.ProviderRef = &inventoryv1alpha1.LocalObjectReference{Name: providerName}
		Expect(k8sClient.Create(testCtx, site)).To(Succeed())

		err := k8sClient.Delete(testCtx, provider)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("still reference it"))

		deleteSite(siteName)
		Eventually(func(g Gomega) {
			err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &inventoryv1alpha1.Site{})
			g.Expect(err).To(HaveOccurred())
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})

var _ = Describe("Site providerRef", func() {
	var regionName, siteName, providerName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		providerName = uniqueName("provider")
		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
	})

	AfterEach(func() {
		deleteSite(siteName)
		deleteProvider(providerName)
		deleteRegion(regionName)
	})

	It("is NotReady with ProviderNotFound while the Provider is absent, then Ready once created", func() {
		site := makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter)
		site.Spec.ProviderRef = &inventoryv1alpha1.LocalObjectReference{Name: providerName}
		Expect(k8sClient.Create(testCtx, site)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.SiteProviderNotFoundReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		Expect(k8sClient.Create(testCtx, makeProvider(providerName))).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Site
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(site), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.SiteReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.SiteReadyReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})
