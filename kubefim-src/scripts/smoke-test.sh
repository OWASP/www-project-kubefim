#!/usr/bin/env bash

set -euo pipefail

binary="${KUBEFIM_BINARY:-./kubefim}"
test_id="kubefim-e2e-$$"
source_path="/tmp/${test_id}-a"
destination_path="/tmp/${test_id}-b"
log_file="$(mktemp)"
agent_pid=""

cleanup() {
    if [[ -n "${agent_pid}" ]] && kill -0 "${agent_pid}" 2>/dev/null; then
        kill -TERM "${agent_pid}" 2>/dev/null || true
        wait "${agent_pid}" 2>/dev/null || true
    fi
    rm -f "${source_path}" "${destination_path}" "${log_file}"
}
trap cleanup EXIT

"${binary}" >"${log_file}" 2>&1 &
agent_pid=$!
sleep 2

if ! kill -0 "${agent_pid}" 2>/dev/null; then
    echo "KubeFIM exited before the smoke test started" >&2
    cat "${log_file}" >&2
    exit 1
fi

touch "${source_path}"
cat "${source_path}" >/dev/null
mv "${source_path}" "${destination_path}"
chmod 600 "${destination_path}"
rm "${destination_path}"
sleep 2

kill -TERM "${agent_pid}"
wait "${agent_pid}"
agent_pid=""

expected_events=(
    "[CREATE]"
    "[OPEN]"
    "[RENAME]"
    "[CHMOD]"
    "[DELETE]"
)

for expected in "${expected_events[@]}"; do
    if ! grep -F "${expected}" "${log_file}" | grep -F "${test_id}" >/dev/null; then
        echo "Missing ${expected} event for ${test_id}" >&2
        cat "${log_file}" >&2
        exit 1
    fi
done

grep -F "${test_id}" "${log_file}"
echo
echo "Agent lifecycle:"
grep -E "Attaching|Listening|Reader closed|unavailable|lost|error|fatal" "${log_file}" || true
echo
echo "KubeFIM end-to-end smoke test passed."
