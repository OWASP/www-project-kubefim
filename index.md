---

layout: col-sidebar
title: OWASP KubeFIM
tags: kubefim kubernetes eBPF FIM runtime-security
level: 2
type: Code
pitch: File integrity and process monitoring for Kubernetes, built with eBPF.

---

<div class="text-center mb-4">
  <img src="/www-project-kubefim/assets/images/kubefim-logo-v2.png" alt="OWASP KubeFIM" width="280">
  <h1 class="mt-3">Runtime file visibility for Kubernetes</h1>
  <p class="lead">
    KubeFIM records file changes and process execution at the Linux kernel and
    connects each event to the process, container, Pod, namespace, and node.
  </p>
  <p>
    <a class="btn btn-primary" href="https://github.com/OWASP/www-project-kubefim/tree/main/kubefim-src">View source</a>
    <a class="btn btn-outline-primary" href="https://github.com/OWASP/www-project-kubefim/issues">Open an issue</a>
    <a class="btn btn-outline-primary" href="https://hub.docker.com/r/abhijitowasp/owasp-kubefim">Docker Hub</a>
  </p>
</div>

---

## One agent per node

KubeFIM runs as a DaemonSet on Linux worker nodes. It does not require a sidecar,
an injected library, or changes to application images. eBPF programs observe
syscall tracepoints, and a small Go agent converts the kernel records into
structured JSON events.

The current alpha records:

- create, open, delete, rename, and permission-change operations
- successful and failed process execution through `execve` and `execveat`
- process IDs, parent ID, user and group IDs, command, and syscall result
- cgroup, mount namespace, PID namespace, and full container ID
- Kubernetes node, namespace, Pod, container, image, and image ID

## How an event is built

| Stage | What happens |
| --- | --- |
| Kernel | eBPF entry and exit probes correlate the syscall, path, and result. |
| Agent | The Go process reads the ring buffer and decodes the stable event layout. |
| Container | The process cgroup is resolved through the node's read-only `/proc` mount. |
| Kubernetes | A node-filtered Pod list/watch cache adds workload metadata without an API call per event. |
| Policy and output | Rules classify the event and JSON Lines are written to standard output. |

KubeFIM uses a dedicated service account. Its ClusterRole permits only `get`,
`list`, and `watch` on Pods, and a ClusterRoleBinding grants that role to the
service account in `kubefim-system`.

## Install

```sh
git clone https://github.com/OWASP/www-project-kubefim.git
cd www-project-kubefim/kubefim-src
kubectl apply -k deployments/kubernetes
kubectl rollout status daemonset/kubefim -n kubefim-system
kubectl logs -n kubefim-system daemonset/kubefim -c kubefim -f
```

The release images support Linux AMD64 and ARM64. KubeFIM needs runtime kernel
BTF, tracefs syscall tracepoints, and permission to run a privileged DaemonSet.
The initializer checks those interfaces; it does not install kernel packages or
download headers onto the node.

[Read the installation and configuration guide](https://github.com/OWASP/www-project-kubefim/blob/main/kubefim-src/README.md).

## Kubernetes and cloud environments

The deployment is cloud-provider neutral and is intended for standard Linux
worker nodes on Amazon EKS, Google Kubernetes Engine, Azure Kubernetes Service,
and self-managed clusters. Compatibility is determined by the node kernel and
the cluster admission policy, not by a cloud API.

The manifests have been API-validated against Kubernetes 1.34.8, 1.35.5, and
1.36.1. End-to-end collection has been tested on Ubuntu 24.04 ARM64 with kernel
6.8, containerd, and K3s 1.36.1. Provider-specific EKS, GKE, and AKS testing is
still in progress. Serverless nodes, virtual nodes, and managed modes that reject
privileged DaemonSets are not currently supported.

## Output and integrations

KubeFIM writes one JSON object per line to stdout. Kubernetes log collectors can
forward these events to Elasticsearch, CloudWatch Logs, Loki, or another log
platform without a KubeFIM-specific exporter. Prometheus metrics, Grafana
dashboards, and documented backend pipelines are on the roadmap.

## Project status

`v0.1.0-alpha.1` is an evaluation release. KubeFIM observes activity; a single
syscall is not proof of malicious intent. Policy is observe-only by default, and
the agent currently requires privileged access to the host kernel.

Near-term work includes additional syscall coverage, path resolution, workload
owner metadata, retained metadata for terminated Pods, metrics, dashboards, and
a broader managed-Kubernetes test matrix.

## Help improve KubeFIM

Testing on real clusters is the most useful contribution right now. Please
[create an issue](https://github.com/OWASP/www-project-kubefim/issues) when you
find a missing event, unsupported kernel, unresolved container, noisy policy, or
integration problem. Compatibility reports should include the cloud provider,
Kubernetes version, node image, kernel, architecture, and container runtime.

Contributions in eBPF, Go, Kubernetes, detection engineering, testing, and
documentation are welcome. Read the [contribution guidelines](CONTRIBUTING) and
use the [security policy](SECURITY) for vulnerability reports.

OWASP KubeFIM is an OWASP Incubator project released under the Apache License 2.0.
