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

var _ = Describe("Circuit Controller", func() {
	var regionName, siteA, siteZ, providerName, circuitName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteA = uniqueName("site")
		siteZ = uniqueName("site")
		providerName = uniqueName("provider")
		circuitName = uniqueName("circuit")

		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteA, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteZ, regionName, inventoryv1alpha1.SiteTypeEdge))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeProvider(providerName))).To(Succeed())

		DeferCleanup(func() { deleteRegion(regionName) })
		DeferCleanup(func() { deleteSite(siteA) })
		DeferCleanup(func() { deleteSite(siteZ) })
		DeferCleanup(func() { deleteProvider(providerName) })
		DeferCleanup(func() { deleteCircuit(circuitName) })
	})

	It("becomes Ready=True with EndpointsResolved once provider and sites resolve", func() {
		circuit := makeCircuit(circuitName, providerName, siteA, siteZ)
		Expect(k8sClient.Create(testCtx, circuit)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Circuit
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(circuit), &fetched)).To(Succeed())
			ready := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.CircuitReady)
			g.Expect(ready).NotTo(BeNil())
			g.Expect(ready.Status).To(Equal(metav1.ConditionTrue))
			resolved := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.CircuitEndpointsResolved)
			g.Expect(resolved).NotTo(BeNil())
			g.Expect(resolved.Status).To(Equal(metav1.ConditionTrue))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("round-trips a cross-group serviceRef", func() {
		circuit := makeCircuit(circuitName, providerName, siteA, siteZ)
		circuit.Spec.ServiceRef = &inventoryv1alpha1.ObjectReference{
			APIGroup:  "networking.miloapis.com",
			Kind:      "Network",
			Name:      "galactic-vpc-uplink",
			Namespace: "default",
		}
		Expect(k8sClient.Create(testCtx, circuit)).To(Succeed())

		var fetched inventoryv1alpha1.Circuit
		Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(circuit), &fetched)).To(Succeed())
		Expect(fetched.Spec.ServiceRef).NotTo(BeNil())
		Expect(fetched.Spec.ServiceRef.APIGroup).To(Equal("networking.miloapis.com"))
		Expect(fetched.Spec.ServiceRef.Name).To(Equal("galactic-vpc-uplink"))
	})

	It("rejects deleting a Provider while a Circuit references it", func() {
		Expect(k8sClient.Create(testCtx, makeCircuit(circuitName, providerName, siteA, siteZ))).To(Succeed())

		err := k8sClient.Delete(testCtx, makeProvider(providerName))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Circuit(s)"))
	})

	It("rejects a Circuit referencing a missing Provider", func() {
		circuit := makeCircuit(circuitName, "ghost-provider", siteA, siteZ)
		err := k8sClient.Create(testCtx, circuit)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Provider ghost-provider not found"))
	})

	It("rejects a Circuit whose Site endpoint does not exist", func() {
		circuit := makeCircuit(circuitName, providerName, siteA, "ghost-site")
		err := k8sClient.Create(testCtx, circuit)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Site ghost-site not found"))
	})

	It("rejects providerRef mutation", func() {
		Expect(k8sClient.Create(testCtx, makeCircuit(circuitName, providerName, siteA, siteZ))).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.Circuit
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: circuitName}, &fetched)).To(Succeed())
			fetched.Spec.ProviderRef.Name = "another-provider"
			err := k8sClient.Update(testCtx, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("providerRef is immutable"))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects deleting a Site while a Circuit terminates on it", func() {
		Expect(k8sClient.Create(testCtx, makeCircuit(circuitName, providerName, siteA, siteZ))).To(Succeed())

		err := k8sClient.Delete(testCtx, makeSite(siteA, regionName, inventoryv1alpha1.SiteTypeDatacenter))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Circuit(s) still reference it"))
	})
})
