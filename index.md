---

layout: col-sidebar
title: OWASP KubeFIM
tags: kubefim kubernetes eBPF FIM runtime-security
level: 2
type: Code
pitch: eBPF-powered file integrity monitoring for Kubernetes workloads and nodes.

---

<p align="center">
  <img src="/assets/images/kubefim.png" alt="OWASP KubeFIM logo" width="420">
</p>

# File integrity visibility for Kubernetes

OWASP KubeFIM is an open source, eBPF-powered File Integrity Monitoring project
for Kubernetes. It observes filesystem activity at the Linux kernel, connects it
to the process responsible, and is being built to attribute each event to its
container, pod, namespace, and node.

[View the source code](https://github.com/OWASP/www-project-kubefim){: .btn .btn-primary }
[Report an issue](https://github.com/OWASP/www-project-kubefim/issues){: .btn .btn-outline-primary }

> **Project status:** KubeFIM is under active development as an OWASP Incubator
> project. The current Linux agent captures and tests core filesystem syscall
> events. Kubernetes enrichment, policy controls, metrics, and production
> packaging are being developed incrementally.

## Why KubeFIM?

Containers are lightweight and short-lived, but they still share the Linux
kernel of their Kubernetes node. Security teams need to understand when files are
created, removed, renamed, or have their permissions changed—and which workload
and process caused the activity.

KubeFIM is designed to provide that visibility from a node-level DaemonSet,
without modifying application images or injecting an agent into every workload.

## Current capabilities

| Capability | Status |
| --- | --- |
| Observe create and open activity through `openat` | Available |
| Observe deletion through `unlinkat` | Available |
| Observe rename activity and both paths through `renameat2` | Available |
| Observe path-based permission changes | Available |
| Capture PID, TGID, PPID, UID, GID, and process command | Available |
| Capture cgroup, mount namespace, and PID namespace identity | Available |
| Correlate syscall entry and exit, including failures | Available |
| Automated privileged Linux smoke testing | Available |
| Container and Kubernetes metadata enrichment | In development |
| Policy-based filtering and noise reduction | Planned |
| JSON logs and Prometheus metrics | Planned |
| Grafana, Elasticsearch, and CloudWatch examples | Planned |

## How it works

1. eBPF programs attach to supported filesystem syscall tracepoints on each Linux
   node.
2. Entry and exit events are correlated so KubeFIM can report the attempted
   operation and its result.
3. A node-local Go agent reads the kernel event stream and converts it into a
   stable event model.
4. Container and Kubernetes metadata will enrich that event from local caches.
5. Policy rules will retain security-relevant activity and suppress routine
   noise.
6. Structured logs and bounded-cardinality metrics will feed existing security
   and observability platforms.

```text
Linux syscalls
      ↓
eBPF entry/exit programs
      ↓
KubeFIM node agent
      ↓
Container and Kubernetes enrichment
      ↓
Policy filtering
      ↓
JSON logs · Prometheus metrics · security backends
```

## Example event

The current development-friendly output identifies the operation, process, user,
and path:

```text
[CREATE] PID=28886 UID=0 COMM=touch PATH=/tmp/kubefim-example
[RENAME] PID=28888 UID=0 COMM=mv PATH=/tmp/kubefim-example
[CHMOD] PID=28889 UID=0 COMM=chmod PATH=/tmp/kubefim-renamed
[DELETE] PID=28890 UID=0 COMM=rm PATH=/tmp/kubefim-renamed
```

The internal event contract already carries additional kernel identity and
syscall-result fields needed for container attribution and structured output.

## Project direction

KubeFIM is progressing through these major areas:

- Reliable filesystem event semantics and broader syscall coverage
- Versioned structured JSON events
- Container identification using cgroups and namespaces
- Kubernetes pod, namespace, container, and node enrichment
- Configurable FIM policies and read-access noise reduction
- Prometheus metrics and Grafana dashboards
- Elasticsearch and CloudWatch log-delivery examples
- Hardened Kubernetes manifests, Helm packaging, and release engineering

## Get involved

KubeFIM welcomes contributors interested in eBPF, Linux internals, Go,
Kubernetes, detection engineering, testing, documentation, and observability.

- Read the [contribution guidelines](CONTRIBUTING)
- Browse [open issues](https://github.com/OWASP/www-project-kubefim/issues)
- Review the [security policy](SECURITY)
- Join the [OWASP Slack community](https://owasp.org/slack/invite)
- Explore the [project repository](https://github.com/OWASP/www-project-kubefim)

## OWASP project

OWASP KubeFIM is an OWASP Incubator project created to advance open, practical
runtime file-integrity visibility for Kubernetes defenders. It is licensed under
the [Apache License 2.0](LICENSE).
