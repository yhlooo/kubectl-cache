[简体中文](README_CN.md) | **[English](README.md)**

**Most of this article is translated from [README_CN.md](README_CN.md) by [ChatGPT 3.5](https://chatgpt.com/). :)**

---

# kubectl-cache

`kubectl-cache` is a plugin for [kubectl](https://kubernetes.io/docs/reference/kubectl/) (the [Kubernetes](https://kubernetes.io) CLI client) used to get or list Kubernetes resources from local cache.

Traditionally, `kubectl get ...` get or lists Kubernetes resources by calling the Kubernetes APIServer's **get** or **list** endpoints. However, the **list** endpoint can be inefficient when dealing with a large number of objects. Frequent **list** calls can stress the APIServer and its storage backend (etcd), potentially affecting APIServer stability.

A common solution to this problem is to build a local cache of resources that need to be queried and detect changes in the cluster using the **watch** interface (to update the cache), as described in [Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes). Kubernetes' Go client library, [k8s.io/client-go](https://pkg.go.dev/k8s.io/client-go), includes tools such as Informers ([k8s.io/client-go/informers](https://pkg.go.dev/k8s.io/client-go/informers)), which implement this access pattern.

**Informer is great, but can it be used with `kubectl` commands?**

The answer is `kubectl-cache`, which provides a way to use Informers with `kubectl` commands. `kubectl-cache` efficiently queries Kubernetes resources based on a local cache without the need for frequent **list** operations, transparently to the user. It provides a `get` subcommand that operates almost identically to `kubectl get`.

## Examples

TODO: ...

## How It Works

Using the `get` subcommand (`kubectl cache get` or `kubectl-cache get`), you can get or list Kubernetes resources (similar to `kubectl get`). See [Get or List Kubernetes Resources](#get-or-list-kubernetes-resources-get).

When you execute `kubectl-cache`'s `get` subcommand, it checks if there is a running proxy service in the background (and starts a new one if not), then queries Kubernetes resources through this proxy. The proxy manages and queries the local cache using Informers. If there are no requests for 10 minutes, the proxy will exit automatically.

![kubectl-cache-proxy](docs/images/kubectl-cache-get.drawio.svg)

In addition to transparently starting a proxy through the `get` subcommand, you can also explicitly run a proxy for the Kubernetes APIServer locally using the `proxy` subcommand (`kubectl cache proxy` or `kubectl-cache proxy`), similar to `kubectl proxy`. See [Running a Proxy](#running-a-proxy-proxy).

For read requests (get and list), the proxy queries and returns results from the local cache (based on Informers). For write requests (create, update, patch, delete, deletecollection) and watch requests, the proxy forwards the requests directly to the Kubernetes APIServer.

![kubectl-cache-proxy](docs/images/kubectl-cache-proxy.drawio.svg)

## Installation

`kubectl-cache` is a plugin for `kubectl` (see [Extending kubectl with plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)), so it is recommended to install `kubectl` before installing `kubectl-cache` (see [Install kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)).

### Via Krew

[Krew](https://krew.sigs.k8s.io/) is a plugin manager for `kubectl`. If you have Krew installed, you can install `kubectl-cache` with the following command:

```bash
kubectl krew install --manifest-url https://raw.githubusercontent.com/yhlooo/kubectl-cache/master/cache.krew.yaml
```

### Binaries

Download the executable binary from the [Releases](https://github.com/yhlooo/kubectl-cache/releases) page, extract it, and place the `kubectl-cache` file into any `$PATH` directory.

### From Sources

Requires Go 1.22. Execute the following command to download the source code and build it:

```bash
go install github.com/yhlooo/kubectl-cache/cmd/kubectl-cache@latest
```

The built binary will be located in `${GOPATH}/bin` by default. Make sure this directory is included in your `$PATH`.

## Usage

`kubectl-cache` can be used as a standalone CLI tool (via the `kubectl-cache` command) or as a plugin for `kubectl` (via the `kubectl cache` command).

### Get or List Kubernetes Resources (`get`)

You can get or list Kubernetes resources using the `get` subcommand (`kubectl cache get` or `kubectl-cache get`).

Some examples:

```bash
# List all Pods in the default namespace
kubectl cache get pod

# List all Pods in the default namespace with label app=test
kubectl cache get pod -l 'app=test'

# List all Pods in the default namespace in the Pending phase
kubectl cache get pod --field-selector 'status.phase=Pending'

# Get a Pod named test-pod in the default namespace
kubectl cache get pod test-pod

# Get a Pod named test-pod in the default namespace and output in YAML format
kubectl cache get pod test-pod -o yaml

# List all Pods in all namespaces
kubectl cache get pod -A
```

The `kubectl cache get` command behaves almost identically to `kubectl get`, with `cache` added before `get`.

For more options and usage, refer to `kubectl cache get --help`.

### Running a Proxy (`proxy`)

You can run a proxy for the Kubernetes APIServer locally using the `proxy` subcommand (`kubectl cache proxy` or `kubectl-cache proxy`):

```bash
# Run the proxy on port 8001
kubectl cache proxy --port 8001
```

In another terminal:

```bash
# Get or list Kubernetes resources via the proxy
kubectl --server http://127.0.0.1:8001 get pod
```

The `kubectl cache proxy` command behaves almost identically to `kubectl proxy`, with `cache` added before `proxy`.

For more options and usage, refer to `kubectl cache proxy --help`.
