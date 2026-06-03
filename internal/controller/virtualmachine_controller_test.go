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

var _ = Describe("VirtualMachine Controller", func() {
	var regionName, siteName, nodeName, providerName, vmName string

	BeforeEach(func() {
		regionName = uniqueName("region")
		siteName = uniqueName("site")
		nodeName = uniqueName("node")
		providerName = uniqueName("provider")
		vmName = uniqueName("vm")

		Expect(k8sClient.Create(testCtx, makeRegion(regionName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeSite(siteName, regionName, inventoryv1alpha1.SiteTypeDatacenter))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeNode(nodeName, siteName))).To(Succeed())
		Expect(k8sClient.Create(testCtx, makeProvider(providerName))).To(Succeed())

		// LIFO: the per-It VM delete (registered last) runs before these, so
		// the Node/Provider delete-guards see no VM by cleanup time. The
		// guarded deletes retry to absorb the manager-cache lag behind the VM
		// delete.
		DeferCleanup(func() { deleteRegion(regionName) })
		DeferCleanup(func() { deleteSite(siteName) })
		DeferCleanup(func() {
			Eventually(func(g Gomega) {
				p := &inventoryv1alpha1.Provider{}
				p.Name = providerName
				g.Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, p))).To(Succeed())
			}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
		})
		DeferCleanup(func() {
			Eventually(func(g Gomega) {
				n := &inventoryv1alpha1.Node{}
				n.Name = nodeName
				g.Expect(client.IgnoreNotFound(k8sClient.Delete(testCtx, n))).To(Succeed())
			}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
		})
		DeferCleanup(func() { deleteVirtualMachine(vmName) })
	})

	It("becomes Ready=True and inherits the host's topology labels", func() {
		vm := makeVirtualMachine(vmName, nodeName)
		Expect(k8sClient.Create(testCtx, vm)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.VirtualMachine
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(vm), &fetched)).To(Succeed())
			cond := meta.FindStatusCondition(fetched.Status.Conditions, inventoryv1alpha1.VirtualMachineReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal(inventoryv1alpha1.VirtualMachineReadyReason))
			g.Expect(fetched.Labels).To(HaveKeyWithValue(inventoryv1alpha1.TopologySiteLabel, siteName))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("round-trips a cross-group projectRef", func() {
		vm := makeVirtualMachine(vmName, nodeName)
		vm.Spec.ProjectRef = &inventoryv1alpha1.ObjectReference{
			APIGroup: "resourcemanager.miloapis.com",
			Kind:     "Project",
			Name:     "my-project",
		}
		Expect(k8sClient.Create(testCtx, vm)).To(Succeed())

		var fetched inventoryv1alpha1.VirtualMachine
		Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(vm), &fetched)).To(Succeed())
		Expect(fetched.Spec.ProjectRef).NotTo(BeNil())
		Expect(fetched.Spec.ProjectRef.APIGroup).To(Equal("resourcemanager.miloapis.com"))
		Expect(fetched.Spec.ProjectRef.Name).To(Equal("my-project"))
	})

	It("becomes Ready once an optional providerRef resolves", func() {
		vm := makeVirtualMachine(vmName, nodeName)
		vm.Spec.ProviderRef = &inventoryv1alpha1.LocalObjectReference{Name: providerName}
		Expect(k8sClient.Create(testCtx, vm)).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.VirtualMachine
			g.Expect(k8sClient.Get(testCtx, client.ObjectKeyFromObject(vm), &fetched)).To(Succeed())
			g.Expect(meta.IsStatusConditionTrue(fetched.Status.Conditions, inventoryv1alpha1.VirtualMachineReady)).To(BeTrue())
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects a VirtualMachine whose host Node does not exist", func() {
		vm := makeVirtualMachine(vmName, "ghost-node")
		err := k8sClient.Create(testCtx, vm)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("host Node ghost-node not found"))
	})

	It("rejects a VirtualMachine whose providerRef does not exist", func() {
		vm := makeVirtualMachine(vmName, nodeName)
		vm.Spec.ProviderRef = &inventoryv1alpha1.LocalObjectReference{Name: "ghost-provider"}
		err := k8sClient.Create(testCtx, vm)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Provider ghost-provider not found"))
	})

	It("rejects hostRef mutation", func() {
		Expect(k8sClient.Create(testCtx, makeVirtualMachine(vmName, nodeName))).To(Succeed())

		Eventually(func(g Gomega) {
			var fetched inventoryv1alpha1.VirtualMachine
			g.Expect(k8sClient.Get(testCtx, client.ObjectKey{Name: vmName}, &fetched)).To(Succeed())
			fetched.Spec.HostRef.Name = "another-node"
			err := k8sClient.Update(testCtx, &fetched)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("hostRef is immutable"))
		}).WithTimeout(defaultTimeout).WithPolling(defaultInterval).Should(Succeed())
	})

	It("rejects deleting the host Node while a VirtualMachine runs on it", func() {
		Expect(k8sClient.Create(testCtx, makeVirtualMachine(vmName, nodeName))).To(Succeed())

		err := k8sClient.Delete(testCtx, makeNode(nodeName, siteName))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("VirtualMachine(s) still hosted on it"))
	})

	It("rejects deleting a Provider while a VirtualMachine references it", func() {
		vm := makeVirtualMachine(vmName, nodeName)
		vm.Spec.ProviderRef = &inventoryv1alpha1.LocalObjectReference{Name: providerName}
		Expect(k8sClient.Create(testCtx, vm)).To(Succeed())

		err := k8sClient.Delete(testCtx, makeProvider(providerName))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("VirtualMachine(s) still reference it"))
	})
})
