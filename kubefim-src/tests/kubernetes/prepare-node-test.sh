#!/usr/bin/env bash

set -euo pipefail

fixture="$(mktemp -d)"
cleanup() {
    rm -rf "${fixture}"
}
trap cleanup EXIT

mkdir -p "${fixture}/kernel/tracing/events/syscalls"
mkdir -p "${fixture}/kernel/btf"
printf 'test-btf' >"${fixture}/kernel/btf/vmlinux"

KUBEFIM_HOST_SYS="${fixture}" KUBEFIM_SKIP_MOUNT=true \
    deployments/kubernetes/init/prepare-node.sh >/dev/null

rm "${fixture}/kernel/btf/vmlinux"
if KUBEFIM_HOST_SYS="${fixture}" KUBEFIM_SKIP_MOUNT=true \
    deployments/kubernetes/init/prepare-node.sh >/dev/null 2>&1; then
    echo "initializer accepted a node without kernel BTF" >&2
    exit 1
fi

echo "KubeFIM node initializer tests passed."
