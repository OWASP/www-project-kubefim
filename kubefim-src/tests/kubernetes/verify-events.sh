#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
    echo "usage: $0 KUBEFIM_LOG" >&2
    exit 2
fi

log_file=$1
test -r "$log_file"

require_event() {
    operation=$1
    path=$2
    if ! grep -F '"operation":"'"$operation"'"' "$log_file" |
        grep -F '"path":"'"$path"'"' |
        grep -F '"namespace":"kubefim-test"' |
        grep -F '"pod_name":"kubefim-e2e"' |
        grep -F '"container_name":"workload"' >/dev/null; then
        echo "missing enriched $operation event for $path" >&2
        exit 1
    fi
}

require_event create /tmp/kubefim-k8s-e2e-a
require_event rename /tmp/kubefim-k8s-e2e-a
require_event chmod /tmp/kubefim-k8s-e2e-b
require_event exec /tmp/kubefim-k8s-e2e-exec
require_event delete /tmp/kubefim-k8s-e2e-b
require_event open /etc/os-release

if grep -F '"path":"/usr/lib/os-release"' "$log_file" |
    grep -F '"namespace":"kubefim-test"' >/dev/null; then
    echo "excluded /usr/lib/os-release event reached output" >&2
    exit 1
fi

if ! grep -F 'suppressed=true rules=[e2e-runtime-library-reads]' "$log_file" >/dev/null; then
    echo "enforced exclusion decision was not logged" >&2
    exit 1
fi

decision_count=$(grep -F -c 'suppressed=true rules=[e2e-runtime-library-reads]' "$log_file")
if [ "$decision_count" -ne 1 ]; then
    echo "expected one bounded exclusion decision log, found $decision_count" >&2
    exit 1
fi

echo "KubeFIM Kubernetes event verification passed."
