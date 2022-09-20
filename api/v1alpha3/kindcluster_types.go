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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindClusterSpec defines the desired state of KindCluster
type KindClusterSpec struct {
	Nodes []v1alpha4.Node `json:"nodes,omitempty"`
}

// KindClusterStatus defines the observed state of KindCluster
type KindClusterStatus struct {
	// Ready status indicates whether the kind cluster is available
	Ready bool `json:"ready"`
	// FailureReason provides a programmatic error code describing a fatal reconciliation failure
	FailureReason string `json:"failureReason,omitempty"`
	// FailureMessage provides a detailed summary of the reconciliation failure
	FailureMessage string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=cluster-api,shortName=kc
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Whether the KIND cluster is currently running with the specified configuration"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// KindCluster is the Schema for the kindclusters API
type KindCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindClusterSpec   `json:"spec,omitempty"`
	Status KindClusterStatus `json:"status,omitempty"`
}

func (c *KindCluster) ToKindSpec() *v1alpha4.Cluster {
	cluster := &v1alpha4.Cluster{}
	cluster.Name = c.Name
	cluster.Nodes = c.Spec.Nodes
	v1alpha4.SetDefaultsCluster(cluster)

	return cluster
}

//+kubebuilder:object:root=true

// KindClusterList contains a list of KindCluster
type KindClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KindCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KindCluster{}, &KindClusterList{})
}
