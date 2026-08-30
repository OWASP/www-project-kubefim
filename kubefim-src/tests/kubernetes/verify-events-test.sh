#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
fixture=$(mktemp)
trap 'rm -f "$fixture"' EXIT HUP INT TERM

write_event() {
    operation=$1
    path=$2
    printf '%s\n' '{"operation":"'"$operation"'","target":{"path":"'"$path"'"},"kubernetes":{"namespace":"kubefim-test","pod_name":"kubefim-e2e","container_name":"workload"}}' >>"$fixture"
}

write_event create /tmp/kubefim-k8s-e2e-a
write_event rename /tmp/kubefim-k8s-e2e-a
write_event chmod /tmp/kubefim-k8s-e2e-b
write_event exec /tmp/kubefim-k8s-e2e-exec
write_event delete /tmp/kubefim-k8s-e2e-b
write_event open /etc/os-release
printf '%s\n' 'policy decision action=aggregate class=access protected=false exception=false would_suppress=true suppressed=true rules=[e2e-runtime-library-reads] reason="test"' >>"$fixture"

"$script_dir/verify-events.sh" "$fixture"

write_event open /usr/lib/os-release
if "$script_dir/verify-events.sh" "$fixture" >/dev/null 2>&1; then
    echo "verification accepted an excluded event" >&2
    exit 1
fi

echo "KubeFIM Kubernetes event verifier tests passed."
