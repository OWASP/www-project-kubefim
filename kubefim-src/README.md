# OWASP KubeFIM

OWASP KubeFIM is a Linux runtime monitor for Kubernetes. It uses eBPF syscall
tracepoints to record file activity and process execution on a node, then adds
the container and Kubernetes identity responsible for the event.

KubeFIM runs once per Linux node as a DaemonSet. Applications do not need a
sidecar, an injected library, or changes to their container images. Events are
written as JSON Lines to standard output, where an existing log collector can
route them to Elasticsearch, CloudWatch, Loki, or another backend.

The current release is `v0.1.0-alpha.1`. It is suitable for evaluation and
integration work. Read the [current limitations](#current-limitations) before
using it on production nodes.

## What KubeFIM records

- file create and open through `openat`
- delete through `unlinkat`
- rename through `renameat2`, including source and destination paths
- permission changes through `chmod`, `fchmodat`, and `fchmodat2`
- successful and failed `execve` and `execveat` process execution
- syscall result, PID, TGID, PPID, UID, GID, and process command
- cgroup ID, mount namespace ID, and PID namespace ID
- container runtime and full container ID
- node, namespace, Pod, Pod UID, container, image, and image ID

KubeFIM reports activity; it does not label every syscall as an attack. The
policy engine classifies events, protects configured paths, and explains why a
rule matched. The default policy is observe-only so an early configuration
cannot silently remove evidence.

## How it works

The kernel programs attach to syscall entry and exit tracepoints. Entry probes
copy the operation and pathname into a bounded map. Exit probes add the return
value and submit a completed event through an eBPF ring buffer. Process
execution uses the scheduler execution tracepoint to capture the successful
image transition while retaining failed syscall results.

The Go agent reads the ring buffer and resolves the process cgroup from the
host's `/proc`. Containerd, CRI-O, and Docker cgroup paths are parsed into a full
container ID. In Kubernetes, a node-filtered Pod list/watch cache maps that ID
to Pod status. This avoids an API request for every kernel event and keeps API
traffic bounded as event volume grows.

The final path is:

```text
Linux syscall
  -> eBPF entry/exit correlation
  -> ring buffer
  -> Go event decoder
  -> cgroup and container resolution
  -> Kubernetes Pod cache
  -> policy decision
  -> JSON Lines on stdout
```

## Requirements

- Linux on AMD64 or ARM64
- runtime kernel BTF at `/sys/kernel/btf/vmlinux`
- syscall tracepoints under `/sys/kernel/tracing/events/syscalls`
- permission to load eBPF programs; the supplied DaemonSet runs privileged
- a Kubernetes cluster with Linux worker nodes

The node initializer verifies BTF and tracefs and mounts tracefs when required.
It does not install packages, download kernel headers, or write to
`/usr/src` or `/lib/modules`.

## Install on Kubernetes

```sh
git clone https://github.com/OWASP/www-project-kubefim.git
cd www-project-kubefim/kubefim-src
kubectl apply -k deployments/kubernetes
kubectl rollout status daemonset/kubefim -n kubefim-system
kubectl logs -n kubefim-system daemonset/kubefim -c kubefim -f
```

The installation creates:

- the `kubefim-system` namespace
- a `kubefim` service account
- a `kubefim-pod-reader` ClusterRole with only `get`, `list`, and `watch` on Pods
- a ClusterRoleBinding from that role to the service account
- a ConfigMap containing the default policy
- one privileged KubeFIM pod per selected Linux node

The API client uses the mounted service-account token, cluster CA, and HTTPS.
Credentials and environment variables are not included in events.

Release images:

- agent: `abhijitowasp/owasp-kubefim:v0.1.0-alpha.1`
- initializer: `abhijitowasp/owasp-kubefim:init-v0.1.0-alpha.1`

Both tags contain `linux/amd64` and `linux/arm64` images.

## Cloud compatibility

KubeFIM does not depend on a cloud-provider API. The same manifests can be used
on Amazon EKS, Google Kubernetes Engine, Azure Kubernetes Service, and
self-managed Kubernetes when the worker node permits privileged pods and
provides the required BTF and tracefs interfaces.

Managed Kubernetes compatibility depends mainly on the node operating system
and the cluster's admission policy. Autopilot, serverless, virtual-node, and
Fargate-style products commonly restrict privileged DaemonSets or do not expose
a conventional host kernel; those environments should be treated as unsupported
until tested. Admission controls such as Pod Security Admission, Gatekeeper, or
Kyverno must explicitly allow the KubeFIM namespace and workload.

Current validation:

| Area | Status |
| --- | --- |
| Kubernetes API | Manifests validated against 1.34.8, 1.35.5, and 1.36.1 |
| Linux runtime | End-to-end test on Ubuntu 24.04 ARM64, kernel 6.8 |
| Distribution | Functional test on K3s 1.36.1 |
| Container runtime | Functional test with containerd; cgroup parsers cover CRI-O and Docker |
| EKS, GKE, AKS | Provider-neutral design; provider-specific test matrix is still in progress |

If you run KubeFIM on a managed cluster, please open a compatibility issue with
the provider, Kubernetes version, node image, kernel version, architecture, and
container runtime. Successful reports are as useful as failures.

## Configuration

The DaemonSet policy is in
[`deployments/kubernetes/base/configmap.yaml`](deployments/kubernetes/base/configmap.yaml).
A larger annotated example is available at
[`configs/kubefim.example.yaml`](configs/kubefim.example.yaml).

Policy rules can match operations, path prefixes, success or failure, UID,
process command, namespace, Pod, container, and image. Exclusions remain visible
in the configuration and decisions include the matched rule and reason.

Noise exclusions use the same match fields as detection rules. In `observe`
mode, matching events are still written and the decision log contains
`would_suppress=true`. In `enforce` mode, a matching event is omitted from the
event output and counted in the shutdown policy summary. The first API version
accepts exclusions only for `open`, and each exclusion must include an operation
and path scope. Protected paths and events promoted to `alert` cannot be
suppressed.

Start with `mode: observe`, exercise representative workloads, and review the
matched exclusion IDs before enabling `mode: enforce`. The annotated
configuration contains a namespace- and container-scoped example.

## Event format

Each stdout line is an independent JSON object:

```json
{"api_version":"events.kubefim.org/v1alpha1","schema_version":1,"operation":"exec","success":true,"return_value":0,"process":{"pid":131337,"uid":0,"comm":"bash"},"target":{"path":"/tmp/tool"},"container":{"id":"<container-id>","runtime":"containerd","host":false},"kubernetes":{"node":"worker-1","namespace":"payments","pod_name":"api-6c89d","pod_uid":"<pod-uid>","container_name":"api","image":"example/api:v1","image_id":"<image-id>"}}
```

`api_version` identifies the public JSON contract. `schema_version` identifies
the kernel/userspace binary event layout. Consumers should use `api_version` and
ignore unfamiliar optional fields.

## Repository layout

| Path | Purpose |
| --- | --- |
| `cmd/kubefimd` | command-line entry point and dependency wiring |
| `bpf/src` | eBPF C programs, shared kernel structures, and CO-RE type definitions |
| `bpf` | generated Go bindings and embedded eBPF objects |
| `internal/collector` | collector interface and eBPF tracepoint attachment |
| `internal/event` | stable userspace event model and binary decoder |
| `internal/enrichment` | cgroup parsing and Kubernetes Pod-cache enrichment |
| `internal/policy` | configuration loading, rule matching, and decisions |
| `internal/output` | JSON Lines and text output encoders |
| `configs` | complete policy example |
| `deployments/kubernetes` | Kustomize base, initializer, RBAC, and test overlay |
| `scripts` | Linux build and privileged smoke-test helpers |
| `tests/kubernetes` | node-initializer test and controlled workload fixture |

## Languages and dependencies

The userspace agent is written in Go. Kernel programs are written in restricted
C and compiled to eBPF bytecode. Kubernetes resources and policy configuration
are YAML; the node initializer and integration tests use POSIX shell.

The runtime Go dependency set is deliberately small:

- [`github.com/cilium/ebpf`](https://github.com/cilium/ebpf) loads programs,
  attaches tracepoints, reads maps, and consumes the ring buffer
- [`gopkg.in/yaml.v3`](https://github.com/go-yaml/yaml) decodes policy files
- [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) provides low-level
  operating-system interfaces used transitively

`bpf2go`, supplied by `cilium/ebpf`, generates the architecture-specific Go
bindings. Docker builds use pinned Debian/Go and BusyBox image digests. The
release agent image is built from `scratch` and contains only the KubeFIM binary.

## Development and tests

Go checks can run on macOS or Linux:

```sh
go test -race -count=1 ./...
go vet ./...
```

The kernel integration test requires Linux and root:

```sh
go generate ./bpf
go build -o kubefim ./cmd/kubefimd
sudo env KUBEFIM_OUTPUT=json scripts/smoke-test.sh
```

The initializer test does not require root:

```sh
bash tests/kubernetes/prepare-node-test.sh
```

Render the release manifests before changing Kubernetes resources:

```sh
kubectl kustomize deployments/kubernetes
```

## Current limitations

- KubeFIM currently requires a privileged security context.
- Paths are syscall arguments and are not yet canonicalized against directory
  file descriptors, mount roots, or symbolic links.
- Open events are intentionally available but noisy; production policy tuning
  remains environment-specific.
- Deployment, StatefulSet, Job, and other workload ownership is not resolved.
- Metadata for a Pod that terminates before correlation is not retained.
- Prometheus metrics, dashboards, and direct backend exporters are planned.
- Enforcement does not occur in the kernel; KubeFIM is an observation and
  detection system.

## Open an issue

Real cluster feedback drives this project. Open an
[issue](https://github.com/OWASP/www-project-kubefim/issues) for:

- a missing syscall or incomplete event
- a kernel or Kubernetes compatibility result
- container metadata that could not be resolved
- policy noise that can be reduced without hiding security-relevant activity
- an output integration, dashboard, or deployment improvement
- a reproducible bug or a focused feature proposal

Include logs only after removing tokens, credentials, internal image names, and
other sensitive data. Security vulnerabilities should follow the repository
[security policy](../SECURITY.md), not a public issue.

## License

OWASP KubeFIM is licensed under the [Apache License 2.0](../LICENSE).
