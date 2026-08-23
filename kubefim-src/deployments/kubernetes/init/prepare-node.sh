#!/bin/sh

set -eu

host_sys="${KUBEFIM_HOST_SYS:-/host/sys}"
tracefs="${host_sys}/kernel/tracing"
kernel_btf="${host_sys}/kernel/btf/vmlinux"

fail() {
    echo "KubeFIM node initialization failed: $*" >&2
    exit 1
}

[ -d "${host_sys}/kernel" ] || fail "${host_sys} is not a mounted host sysfs"

mkdir -p "${tracefs}"
if ! grep -qs " ${tracefs} tracefs " /proc/mounts; then
    if [ "${KUBEFIM_SKIP_MOUNT:-false}" = "true" ]; then
        echo "Skipping tracefs mount in test mode"
    else
        echo "Mounting tracefs at ${tracefs}"
        mount -t tracefs tracefs "${tracefs}" || fail "cannot mount tracefs"
    fi
fi

[ -d "${tracefs}/events/syscalls" ] || fail "syscall tracepoints are unavailable in tracefs"
[ -r "${kernel_btf}" ] || fail "kernel BTF is unavailable at ${kernel_btf}"
[ -s "${kernel_btf}" ] || fail "kernel BTF at ${kernel_btf} is empty"

echo "KubeFIM kernel prerequisites are ready"
echo "kernel=$(uname -r) architecture=$(uname -m)"
