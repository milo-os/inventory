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

var _ = Describe("Region Controller", func() {
	var regionName string

	BeforeEach(func() {
		regionName = uniqueName("region")
	})

	AfterEach(func() {
		deleteRegion(regionName)
	})

	It("becomes Ready=True shortly after creation", func() {
		region := makeRegion(regionName)
		Expect(k8sClient.Create(testCtx, region)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Region
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(region), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.RegionReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.RegionReadyReason))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("allows deletion when no Site references it", func() {
		region := makeRegion(regionName)
		Expect(k8sClient.Create(testCtx, region)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Region
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(region), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.RegionReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())

		Expect(k8sClient.Delete(testCtx, region)).To(Succeed())

		Eventually(func(g Gomega) {
			err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(region), &inventoryv1alpha1.Region{})
			g.Expect(client.IgnoreNotFound(err)).To(Succeed())
			g.Expect(err).To(HaveOccurred())
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})
})
