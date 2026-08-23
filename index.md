---

layout: col-sidebar
title: OWASP KubeFIM
tags: kubefim kubernetes eBPF FIM runtime-security
level: 2
type: Code
pitch: eBPF file integrity and process monitoring for Kubernetes nodes.

---

<p align="center">
  <img src="/www-project-kubefim/assets/images/kubefim-logo-v2.png" alt="OWASP KubeFIM logo" width="360">
</p>

# File integrity monitoring for Kubernetes

OWASP KubeFIM is an open source Linux node agent that uses eBPF to observe file
activity and process execution. In Kubernetes it connects each event to the
responsible process, container, Pod, namespace, and node.

[Source code](https://github.com/OWASP/www-project-kubefim/tree/main/kubefim-src){: .btn .btn-primary }
[Docker image](https://hub.docker.com/r/abhijitowasp/owasp-kubefim){: .btn .btn-outline-primary }
[Report an issue](https://github.com/OWASP/www-project-kubefim/issues){: .btn .btn-outline-primary }

> **Release status:** `v0.1.0-alpha.1` is intended for evaluation on Linux
> Kubernetes nodes. KubeFIM requires runtime kernel BTF and currently runs as a
> privileged DaemonSet.

## What it records

- file create/open, delete, rename, and permission-change syscalls
- successful and failed `execve` and `execveat` process execution
- syscall result, PID/TGID/PPID, UID/GID, and process command
- cgroup, mount namespace, and PID namespace identifiers
- full container ID and runtime for containerd, CRI-O, and Docker cgroups
- Kubernetes node, namespace, Pod UID/name, container name, image, and image ID

Events are written as JSON Lines for collection by existing log agents. A text
format remains available for development.

## Architecture

```text
syscall tracepoints
        │
        ▼
  eBPF programs ── entry/exit correlation
        │
        ▼
  KubeFIM agent ── cgroup and process identity
        │
        ▼
  Pod list/watch cache ── Kubernetes metadata
        │
        ▼
  policy decision ── JSON Lines to stdout
```

One privileged KubeFIM pod runs on each selected Linux node. The agent reads
`/proc/<pid>/cgroup` through a read-only host mount and maintains a node-filtered
Pod cache. It does not call the Kubernetes API for every event. Its service
account has read-only access to get, list, and watch Pods.

## Install

```sh
git clone https://github.com/OWASP/www-project-kubefim.git
cd www-project-kubefim/kubefim-src
kubectl apply -k deployments/kubernetes
kubectl rollout status daemonset/kubefim -n kubefim-system
kubectl logs -n kubefim-system daemonset/kubefim -c kubefim -f
```

The supplied manifests have been API-tested with Kubernetes 1.34, 1.35, and
1.36. The functional eBPF test runs on Ubuntu 24.04 ARM64 with kernel 6.8 and
K3s 1.36.1. Kernel compatibility and Kubernetes API compatibility are separate:
the node must expose `/sys/kernel/btf/vmlinux` and syscall tracepoints.

## Current boundaries

KubeFIM reports kernel activity but does not claim that every event is
malicious. The policy engine currently runs in observe mode and preserves
protected-path mutations. Workload-owner resolution, metrics, dashboards, and
terminated-Pod retention are planned work.

See the [agent documentation](https://github.com/OWASP/www-project-kubefim/blob/main/kubefim-src/README.md)
for configuration, tests, security notes, and implementation details.

## Contributing

Contributions are welcome in eBPF, Go, Kubernetes, detection engineering,
testing, and documentation. Read the [contribution guidelines](CONTRIBUTING)
and [security policy](SECURITY) before opening a pull request or report.

OWASP KubeFIM is an OWASP Incubator project licensed under Apache License 2.0.
