**[简体中文](README_CN.md)** | [English](README.md)

---

# kubectl-cache

`kubectl-cache` 是 [kubectl](https://kubernetes.io/zh-cn/docs/reference/kubectl/) （一个 [Kubernetes](https://kubernetes.io) CLI 客户端）的插件。用于通过本地缓存获取或列出 Kubernetes 资源。

通过 `kubectl get ...` 可获取或列出 Kubernetes 资源，这是通过调用 Kubernetes APIServer 的 get 或 list 接口实现的。然而在对象数目很多的情况下 list 接口的性能较差，并且频繁调用 list 接口会给 APIServer 和 etcd （ APIServer 的存储后端）带来较大的压力，可能影响 APIServer 稳定性。 

该问题的一个常规解决方案是为需要查询的资源在本地构建缓存并通过 watch 接口检测集群中资源的变更（以更新缓存），见 [高效检测变更](https://kubernetes.io/zh-cn/docs/reference/using-api/api-concepts/#efficient-detection-of-changes) 。 Kubernetes 的 Go 客户端库 [k8s.io/client-go](https://pkg.go.dev/k8s.io/client-go) 中就包含实现该访问模式的工具，即 Informer （ [k8s.io/client-go/informers](https://pkg.go.dev/k8s.io/client-go/informers) ）。

**Informer 很好，因此它能不能被用于 `kubectl` 命令呢？**

答案就是 `kubectl-cache` ，它提供了一个通过 `kubectl` 命令使用 Informer 的方式。 `kubectl-cache` 基于本地缓存可以高效地查询 Kubernetes 资源而无需频繁的 list ，而且这对用户是透明的。它提供了 `get` 子命令，用法与 `kubectl get` 几乎完全一致。

## 示例

TODO: ...

## 原理

通过 `get` 子命令（ `kubectl cache get` 或 `kubectl-cache get` ）可获取或列出 Kubernetes 资源（类似于 `kubectl get` ）。见 [获取或列出 Kubernetes 资源](#获取或列出-kubernetes-资源-get-) 。

当执行 `kubectl-cache` 的 `get` 子命令，它会检测后台是否有正在运行的代理服务（如果没有它会启动一个新的），然后通过该代理查询 Kubernetes 资源。该代理通过 Informer 查询和管理本地缓存。若连续 10 分钟没有请求，代理会自动退出。

![kubectl-cache-proxy](docs/images/kubectl-cache-get.drawio.svg)

除了通过 `get` 子命令透明地启动代理，也可以通过 `proxy` 子命令（ `kubectl cache proxy` 或 `kubectl-cache proxy` ）显式地在本地运行一个 Kubernetes APIServer 的代理（类似于 `kubectl proxy` ），然后直接使用 kubectl 与之交互。见 [运行代理](#运行代理-proxy-) 。

对于部分读请求（ get 和 list ），代理会从本地缓存中查询并返回结果（基于 Informer ）；对于写请求（ create 、 update 、 patch 、 delete 、 deletecollection ）和 watch 请求，代理会直接将请求转发给 Kubernetes APIServer 。

![kubectl-cache-proxy](docs/images/kubectl-cache-proxy.drawio.svg)

## 安装

`kubectl-cache` 是一个 `kubectl` 的插件（见 [用插件扩展 kubectl](https://kubernetes.io/zh-cn/docs/tasks/extend-kubectl/kubectl-plugins/) ），因此安装 `kubectl-cache` 前建议先安装 `kubectl` （见 [安装 kubectl](https://kubernetes.io/zh-cn/docs/tasks/tools/#kubectl) ）。

### 通过 Krew 安装

[Krew](https://krew.sigs.k8s.io/) 是一个 kubectl 的插件管理工具，如果你已经安装了 Krew ，可执行以下命令安装 `kubectl-cache` ：

```bash
kubectl krew install --manifest-url https://raw.githubusercontent.com/yhlooo/kubectl-cache/master/cache.krew.yaml
```

### 通过二进制安装

通过 [Releases](https://github.com/yhlooo/kubectl-cache/releases) 页面下载可执行二进制，解压并将其中 `kubectl-cache` 文件放置到任意 `$PATH` 目录下。

### 从源码编译

要求 Go 1.22 ，执行以下命令下载源码并构建：

```bash
go install github.com/yhlooo/kubectl-cache/cmd/kubectl-cache@latest
```

构建的二进制默认将在 `${GOPATH}/bin` 目录下，需要确保该目录包含在 `$PATH` 中。

## 使用

`kubectl-cache` 可以作为一个独立的 CLI 工具使用（通过 `kubectl-cache` 命令），也可以作为 `kubectl` 的插件使用（通过 `kubectl cache` 命令）。

### 获取或列出 Kubernetes 资源（ `get` ）

通过 `get` 子命令（ `kubectl cache get` 或 `kubectl-cache get` ）可获取或列出 Kubernetes 资源。

一些示例：

```bash
# 列出默认命名空间下所有 Pod
kubectl cache get pod

# 列出默认命名空间下带有标签 app=test 的所有 Pod
kubectl cache get pod -l 'app=test'

# 列出默认命名空间下处于 Pending 阶段的所有 Pod
kubectl cache get pod --field-selector 'status.phase=Pending'

# 获取默认命名空间下名为 test-pod 的 Pod
kubectl cache get pod test-pod

# 获取默认命名空间下名为 test-pod 的 Pod ，并以 YAML 格式输出
kubectl cache get pod test-pod -o yaml

# 列出所有命名空间下所有 Pod
kubectl cache get pod -A
```

`kubectl cache get` 命令与 `kubectl get` 用法几乎完全一致，仅仅在 `get` 前加一个 `cache` 。

更多参数和用法参考 `kubectl cache get --help` 。

### 运行代理（ `proxy` ）

通过 `proxy` 子命令（ `kubectl cache proxy` 或 `kubectl-cache proxy` ）可在本地运行一个 Kubernetes APIServer 的代理：

```bash
# 在 8001 端口运行代理
kubectl cache proxy --port 8001
```

在另一个终端：

```bash
# 通过代理获取或列出 Kubernetes 资源
kubectl --server http://127.0.0.1:8001 get pod
```

`kubectl cache proxy` 命令与 `kubectl proxy` 用法几乎完全一致，仅仅在 `proxy` 前加一个 `cache` 。

更多参数和用法参考 `kubectl cache proxy --help` 。
