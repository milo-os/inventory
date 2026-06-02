// SPDX-License-Identifier: AGPL-3.0-only

package controller_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	inventoryv1alpha1 "go.miloapis.com/inventory/api/v1alpha1"
	inventorycontroller "go.miloapis.com/inventory/internal/controller"
	webhookv1alpha1 "go.miloapis.com/inventory/internal/webhook/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Package-level vars consumed by the per-kind test files.
var (
	cfg        *rest.Config
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testCtx    context.Context
	cancel     context.CancelFunc
	testScheme = runtime.NewScheme()
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inventory Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testCtx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "base", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "base", "webhook")},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	utilruntime.Must(clientgoscheme.AddToScheme(testScheme))
	utilruntime.Must(inventoryv1alpha1.AddToScheme(testScheme))

	k8sClient, err = client.New(cfg, client.Options{Scheme: testScheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	whOpts := testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: testScheme,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    whOpts.LocalServingHost,
			Port:    whOpts.LocalServingPort,
			CertDir: whOpts.LocalServingCertDir,
		}),
		Metrics:        metricsserver.Options{BindAddress: "0"},
		LeaderElection: false,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(inventorycontroller.SetupIndexers(testCtx, mgr)).To(Succeed())

	Expect((&inventorycontroller.RegionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.ProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.SiteReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.RackReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.ClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.NetworkDeviceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())
	Expect((&inventorycontroller.LinkReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)).To(Succeed())

	Expect(webhookv1alpha1.SetupRegionWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupProviderWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupSiteWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupRackWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupClusterWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupNodeWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupNetworkDeviceWebhookWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupLinkWebhookWithManager(mgr)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		if err := mgr.Start(testCtx); err != nil {
			// Context cancellation on AfterSuite is expected; ignore.
			if testCtx.Err() == nil {
				Expect(err).NotTo(HaveOccurred())
			}
		}
	}()

	// Wait for the webhook server to be reachable.
	hostPort := net.JoinHostPort(whOpts.LocalServingHost, fmt.Sprintf("%d", whOpts.LocalServingPort))
	Eventually(func(g Gomega) {
		dialer := &net.Dialer{Timeout: time.Second}
		conn, dialErr := tls.DialWithDialer(dialer, "tcp", hostPort, &tls.Config{InsecureSkipVerify: true})
		g.Expect(dialErr).NotTo(HaveOccurred())
		_ = conn.Close()
	}).WithTimeout(30 * time.Second).WithPolling(500 * time.Millisecond).Should(Succeed())

	// Wait for manager cache to sync.
	Eventually(func() bool {
		return mgr.GetCache().WaitForCacheSync(testCtx)
	}, 10*time.Second, 100*time.Millisecond).Should(BeTrue())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if cancel != nil {
		cancel()
	}
	if testEnv != nil {
		Expect(testEnv.Stop()).To(Succeed())
	}
})
