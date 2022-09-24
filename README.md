# cluster-api-provider-kind

The Cluster-API (CAPI) Provider for KIND is a provider that creates KIND clusters using the CAPI.
Unlike other similar providers that launch the KIND image, this provider relies on a TCP-based
docker host rather than a mounted UNIX socket.

## Installation & Usage

### Prerequisites

In order to use this provider, you need the following installed on your machine:

1. Docker, to run containers
1. `socat`, to proxy the docker unix socket to TCP
1. `kind`, to spin up your management cluster
1. The `clusterctl` CLI, to interact with CAPI

### Installation

1. Create a new KIND cluster: `kind create cluster --name capi-test` (or skip to step 2 if you already have one)
1. Install the CAPI components: `clusterctl init`
1. Clone this repository: `git clone git@github.com:phroggyy/cluster-api-provider-kind.git`
1. Navigate to the project directory for the KIND provider: `cd cluster-api-provider-kind`
1. Install the KIND CRDs: `make install`
1. Deploy the controller: `make deploy IMG=phroggyy/cluster-api-provider-kind`
1. To check if it's installed, try to list your kindclusters: `kubectl get kindcluster` (or `kc` for short)

**Note:** whenever you want the controller to actually interact with your docker agent, you must run `socat`
on your local machine. See step 1 under Usage.

### Usage

1. Start `socat`: `socat TCP-LISTEN:2375,reuseaddr,fork UNIX-CONNECT:/var/run/docker.sock`
1. To create a new cluster, create a CAPI `Cluster` as well as a `KindCluster` resource:
   ```sh
   cat <<EOF | kubectl apply -f -
   apiVersion: cluster.x-k8s.io/v1beta1
   kind: Cluster
   metadata:
     name: capi-demo
   spec:
     clusterNetwork:
       pods:
         cidrBlocks: ["192.168.0.0/16"]
     infrastructureRef:
       apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
       kind: KindCluster
       name: capi-demo
   ---
   apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
   kind: KindCluster
   metadata:
     name: capi-demo
   spec:
     nodes:
     - role: control-plane
   ```
1. Wait for your cluster to be ready: `kubectl wait --for=jsonpath='{.status.ready}'=true kc capi-demo`

### Caveats & limitations

#### KIND node management
As KIND does not support managing nodes for a cluster independently, this provider is not entirely compliant
with CAPI in that there is no distinct `Machine` spec. Due to the same limitations imposed by KIND, this
provider comes with the caveat that any updates to a `KindCluster` resource will result in the underlying
cluster being _recreated_ rather than updated. As such, be very cautious about issuing updates, as your existing
cluster (and all its resources) will be removed.

#### Docker compatibility & requirements

_Note: this provider has only been tested on macOS, and no guarantees are provided for other operating systems._

This project relies specifically on Docker for Desktop, rather than just the Docker Engine. As such, if you
are attempting to run this in a linux environment, it may not work as expected.

This is because we rely on TCP rather than a UNIX socket for the connection to the docker daemon, for which
this provider uses the `host.docker.internal` hostname to connect. That hostname is only configured
on Docker for Desktop.

Additionally, because of this, you **must** use the Docker provider for KIND (rather than podman which is also supported). This provider relies on the networking implementation of Docker to ensure that we can reach the
host network from within our KIND cluster, which is what allows us to deploy the CAPI provider in a cluster,
rather than running it on the host.

## Contributing
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on your dev cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/cluster-api-provider-kind:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/cluster-api-provider-kind:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

