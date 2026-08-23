# KubeFIM agent

KubeFIM is a privileged Linux node agent that collects file and process events
with eBPF. In Kubernetes it runs as a DaemonSet, resolves the event cgroup to a
full container ID, and enriches that ID from a node-scoped Pod list/watch cache.

## Requirements

- Linux on AMD64 or ARM64
- readable runtime kernel BTF at `/sys/kernel/btf/vmlinux`
- syscall tracepoints under `/sys/kernel/tracing/events/syscalls`
- permission to load eBPF programs; the supplied DaemonSet runs privileged
- Kubernetes 1.34, 1.35, or 1.36 for the tested manifest set

The init container checks these kernel interfaces and mounts tracefs when it is
not already mounted. It does not download kernel headers, install packages, or
modify `/usr/src` and `/lib/modules` on the node.

## Kubernetes installation

Clone the repository and apply the Kustomize deployment:

```sh
git clone https://github.com/OWASP/www-project-kubefim.git
cd www-project-kubefim/kubefim-src
kubectl apply -k deployments/kubernetes
kubectl rollout status daemonset/kubefim -n kubefim-system
kubectl logs -n kubefim-system daemonset/kubefim -c kubefim -f
```

The release uses these Docker Hub tags:

- agent: `abhijitowasp/owasp-kubefim:v0.1.0-alpha.1`
- node initializer: `abhijitowasp/owasp-kubefim:init-v0.1.0-alpha.1`

The service account can only get, list, and watch Pods. The API client uses the
mounted service-account CA and TLS; credentials are never written to events.

## Event output

The DaemonSet writes one JSON object per line. A pod event includes the kernel
identity, process, target, container, and Kubernetes fields:

```json
{"api_version":"events.kubefim.org/v1alpha1","schema_version":1,"operation":"exec","success":true,"process":{"pid":131337,"uid":0,"comm":"bash"},"target":{"path":"/tmp/tool"},"container":{"id":"<full-container-id>","runtime":"containerd","host":false},"kubernetes":{"node":"worker-1","namespace":"payments","pod_name":"api-6c89d","pod_uid":"<pod-uid>","container_name":"api","image":"example/api:v1","image_id":"<image-id>"}}
```

The default policy is mounted from
[`deployments/kubernetes/base/configmap.yaml`](deployments/kubernetes/base/configmap.yaml).
The complete configuration example is
[`configs/kubefim.example.yaml`](configs/kubefim.example.yaml).

## Local development

Run the Go checks on macOS or Linux:

```sh
go test -race ./...
go vet ./...
```

The eBPF integration test must run on Linux:

```sh
go generate ./bpf
go build -o kubefim ./cmd/kubefimd
sudo env KUBEFIM_OUTPUT=json scripts/smoke-test.sh
```

The Kubernetes initializer fixture test does not require root:

```sh
bash tests/kubernetes/prepare-node-test.sh
```

## Security and current limitations

KubeFIM observes kernel events; a single syscall is evidence of an action, not
proof of malicious intent. Policy rules should reduce noise without discarding
protected-path mutations. Failed container correlation is left unattributed
rather than guessed.

The current release does not resolve Deployment or Job ownership, retain
terminated Pod metadata, expose Prometheus metrics, or enforce policy in the
kernel. Review the privileged DaemonSet and host mounts before production use.

Report security issues through the repository [security policy](../SECURITY.md).
