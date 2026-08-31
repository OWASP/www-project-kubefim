# Changelog

Notable changes to OWASP KubeFIM are recorded here. The project uses semantic
versioning for published container images and marks evaluation builds as
pre-releases.

## [v0.1.0-alpha.3] - 2026-08-31

### Added

- Prometheus metrics for processed, emitted, suppressed, would-suppress,
  protected, lost-sample, collector-read-error, and output-error activity.
- `/metrics` and `/healthz` HTTP endpoints with Kubernetes readiness and
  liveness probes.
- A provisioned Grafana dashboard and Prometheus development stack.
- A Prometheus Operator `ServiceMonitor` deployment overlay.
- Fixed-cardinality operation and result labels to keep metric growth bounded.

### Changed

- Published Linux AMD64 and ARM64 agent and initializer images with SBOM and
  provenance attestations.
- Documented and verified the KubeFIM-to-Prometheus-to-Grafana path on K3s.

## [v0.1.0-alpha.2] - 2026-08-31

### Added

- Configurable policy rules for reducing known file-event noise.
- Observe-only policy evaluation before exclusions are enforced.
- Protected-path safeguards that take precedence over exclusion rules.
- Process execution monitoring for `execve` and `execveat`.
- Kubernetes workload enrichment through a node-filtered Pod list/watch cache.
- Container identity resolution for containerd, CRI-O, and Docker cgroups.

### Changed

- Published Linux AMD64 and ARM64 agent and initializer images with SBOM and
  provenance attestations.

[v0.1.0-alpha.3]: https://github.com/OWASP/www-project-kubefim/tree/v0.1.0-alpha.3
[v0.1.0-alpha.2]: https://github.com/OWASP/www-project-kubefim/tree/v0.1.0-alpha.2
