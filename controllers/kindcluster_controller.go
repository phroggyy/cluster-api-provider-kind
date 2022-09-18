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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"

	infrastructurev1alpha3 "github.com/phroggyy/cluster-api-provider-kind/api/v1alpha3"
)

// KindClusterReconciler reconciles a KindCluster object
type KindClusterReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	KindProvider *cluster.Provider
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KindCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KindClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	cluster := &infrastructurev1alpha3.KindCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !cluster.DeletionTimestamp.IsZero() {
		logger.Info("removing deleted cluster", "cluster_name", req.NamespacedName)
		return ctrl.Result{}, r.deleteCluster(cluster)
	}

	// check if it's create or update, probably using a finaliser?
	// r.replaceCluster(cluster)

	return ctrl.Result{}, r.replaceCluster(cluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KindClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha3.KindCluster{}).
		Complete(r)
}

func (r *KindClusterReconciler) deleteCluster(c *infrastructurev1alpha3.KindCluster) error {
	return nil
}

func (r *KindClusterReconciler) replaceCluster(cluster *infrastructurev1alpha3.KindCluster) error {
	clusters, err := r.KindProvider.List()

	if err != nil {
		return err
	}

	exists := false
	for _, c := range clusters {
		if c == cluster.Name {
			exists = true
			break
		}
	}

	// KIND doesn't support updating clusters, so we will delete the existing cluster
	// TODO: (future) some config option whether to replace-on-modify or enable a validator to prevent changes
	if exists {
		err = r.KindProvider.Delete(cluster.Name, "")
		if err != nil {
			return err
		}
	}

	return r.KindProvider.Create(
		cluster.Name,
		kindcluster.CreateWithV1Alpha4Config(cluster.ToKindSpec()),
	)
}
