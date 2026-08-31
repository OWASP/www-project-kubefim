# OWASP KubeFIM

OWASP KubeFIM is an open-source, eBPF-based file integrity and process execution
monitor for Kubernetes. It attaches eBPF programs to syscall tracepoints on each
Linux node and reports the operation, result, process identity, container
runtime identity, and Kubernetes workload metadata as JSON Lines.

KubeFIM is built for runtime security and incident investigation. It shows which
process changed a file, which container ran the process, and which Pod and
namespace owned the container—without modifying application images or injecting
sidecars.

The active agent source, container build, tests, and deployment manifests are
in [`kubefim-src`](kubefim-src/). The OWASP project page is generated from
[`index.md`](index.md).

## Current release scope

The current evaluation release is [`v0.1.0-alpha.3`](https://github.com/OWASP/www-project-kubefim/tree/v0.1.0-alpha.3).

- `openat`, `unlinkat`, `renameat2`, path-based `chmod`, `execve`, and `execveat`
- syscall success or failure and return value
- PID, TGID, PPID, UID, GID, cgroup ID, mount namespace, and PID namespace
- containerd, CRI-O, and Docker cgroup identification
- node-scoped Kubernetes Pod enrichment without an API call per event
- observe-mode policy classification and protected-path rules
- configurable noise exclusions with would-suppress and protected-event counters
- text output for development and versioned JSON Lines for log pipelines
- Prometheus metrics, health checks, and a provisioned Grafana dashboard
- tested Kubernetes manifests for versions 1.34, 1.35, and 1.36

KubeFIM is an early release. It requires a Linux kernel with syscall tracepoints
and runtime BTF available at `/sys/kernel/btf/vmlinux`. The DaemonSet is
privileged because it loads eBPF programs and observes host processes.

See the [agent README](kubefim-src/README.md) for installation, configuration,
testing, compatibility, and security details.

## Project links

- [OWASP project page](https://owasp.org/www-project-kubefim/)
- [Source and issues](https://github.com/OWASP/www-project-kubefim)
- [Container images](https://hub.docker.com/r/abhijitowasp/owasp-kubefim)
- [Releases](https://github.com/OWASP/www-project-kubefim/releases)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

Licensed under the Apache License 2.0.
