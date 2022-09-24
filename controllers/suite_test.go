/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"

	infrastructurev1alpha3 "github.com/phroggyy/cluster-api-provider-kind/api/v1alpha3"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var mock *KindMock

type KindMock struct {
	clusters []string
}

func (m *KindMock) List() ([]string, error) {
	return m.clusters, nil
}
func (m *KindMock) Delete(name, explicitKubeconfigPath string) error {
	index := -1
	for i, c := range m.clusters {
		if c == name {
			index = i
		}
	}

	if index == -1 {
		return errors.New("Cluster does not exist")
	}

	m.clusters = append(m.clusters[:index], m.clusters[index+1:]...)
	return nil
}
func (m *KindMock) Create(name string, options ...kindcluster.CreateOption) error {
	m.clusters = append(m.clusters, name)
	return nil
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = infrastructurev1alpha3.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	mock = &KindMock{clusters: []string{}}
	err = (&KindClusterReconciler{
		Client:       k8sManager.GetClient(),
		Scheme:       k8sManager.GetScheme(),
		KindProvider: mock,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
}, 60)

const (
	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("kindcluster controller", func() {
	const (
		clusterName      = "test-cluster"
		clusterNamespace = "default"
	)

	// after each test, we will delete all existing clusters
	var _ = AfterEach(func() {
		l := &infrastructurev1alpha3.KindClusterList{}
		k8sClient.List(context.Background(), l)

		By(fmt.Sprintf("removing existing clusters: %d", len(l.Items)))
		k8sClient.Delete(context.Background(), &infrastructurev1alpha3.KindCluster{
			ObjectMeta: v1.ObjectMeta{
				Name:      clusterName,
				Namespace: clusterNamespace,
			},
		})
		// Expect()
		// Expect(k8sClient.DeleteAllOf(context.Background(), &infrastructurev1alpha3.KindCluster{})).Should(Succeed())
		Eventually(func() int {
			l := &infrastructurev1alpha3.KindClusterList{}
			k8sClient.List(context.Background(), l)

			return len(l.Items)
		}, timeout, interval).Should(Equal(0))
		By("Finished removing existing clusters")
	})

	When("Updating KindCluster Status", func() {
		It("Should set Status.Ready when a cluster is ready", func() {
			By("creating a new KindCluster")
			ctx := context.Background()
			kindCluster := &infrastructurev1alpha3.KindCluster{
				TypeMeta: v1.TypeMeta{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
					Kind:       "KindCluster",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      clusterName,
					Namespace: clusterNamespace,
				},
				Spec: infrastructurev1alpha3.KindClusterSpec{
					Nodes: []v1alpha4.Node{{
						Role: "control-plane",
					}},
				},
			}

			Expect(k8sClient.Create(ctx, kindCluster)).Should(Succeed())

			// ensure our list of clusters is empty
			Expect(mock.clusters).Should(HaveLen(0))
			lookupKey := types.NamespacedName{Name: clusterName, Namespace: clusterNamespace}
			createdKindCluster := &infrastructurev1alpha3.KindCluster{}

			Eventually(func() int {
				return len(mock.clusters)
			}, timeout, interval).Should(Equal(1))

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdKindCluster)
				return err == nil && createdKindCluster.Status.Ready
			}, timeout, interval).Should(BeTrue())

			Expect(mock.clusters).Should(HaveLen(1))
			Expect(createdKindCluster.Status.Ready).To(BeTrue())
		})
	})

	When("Deleting KindCluster", func() {
		It("Should delete the underlying KIND cluster when deleting the resource", func() {
			By("creating a new KindCluster")
			ctx := context.Background()
			kindCluster := &infrastructurev1alpha3.KindCluster{
				TypeMeta: v1.TypeMeta{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
					Kind:       "KindCluster",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      clusterName,
					Namespace: clusterNamespace,
				},
				Spec: infrastructurev1alpha3.KindClusterSpec{
					Nodes: []v1alpha4.Node{{
						Role: "control-plane",
					}},
				},
			}

			Expect(k8sClient.Create(ctx, kindCluster)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: clusterName, Namespace: clusterNamespace}
			createdKindCluster := &infrastructurev1alpha3.KindCluster{}

			// the cluster has been created
			Eventually(func() int {
				return len(mock.clusters)
			}, timeout, interval).Should(Equal(1))

			err := k8sClient.Delete(ctx, kindCluster)
			Expect(err).To(Succeed())

			// once the cluster no longer exists, it should have been properly deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdKindCluster)
				return err == nil
			}, timeout, interval).Should(BeFalse())

			Expect(mock.clusters).Should(HaveLen(0))
		})
	})
})
