#!/usr/bin/env bash

set -euo pipefail

binary="${KUBEFIM_BINARY:-./kubefim}"
config="${KUBEFIM_CONFIG:-}"
output_format="${KUBEFIM_OUTPUT:-text}"
test_id="kubefim-e2e-$$"
source_path="/tmp/${test_id}-a"
destination_path="/tmp/${test_id}-b"
executable_path="/tmp/${test_id}-exec"
missing_executable_path="/tmp/${test_id}-missing"
execve_launcher="/tmp/${test_id}-execve"
execveat_launcher="/tmp/${test_id}-execveat"
log_file="$(mktemp)"
agent_pid=""

cc -Wall -Wextra -Werror scripts/testdata/execve.c -o "${execve_launcher}"
cc -Wall -Wextra -Werror scripts/testdata/execveat.c -o "${execveat_launcher}"

cleanup() {
    if [[ -n "${agent_pid}" ]] && kill -0 "${agent_pid}" 2>/dev/null; then
        kill -TERM "${agent_pid}" 2>/dev/null || true
        wait "${agent_pid}" 2>/dev/null || true
    fi
    rm -f "${source_path}" "${destination_path}" "${executable_path}" \
        "${execve_launcher}" "${execveat_launcher}" "${log_file}"
}
trap cleanup EXIT

command=("${binary}" --output "${output_format}")
if [[ -n "${config}" ]]; then
    command+=(--config "${config}")
fi

"${command[@]}" >"${log_file}" 2>&1 &
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
cp /bin/true "${executable_path}"
chmod 700 "${executable_path}"
"${execve_launcher}" "${executable_path}"
"${execveat_launcher}" "${executable_path}"
"${missing_executable_path}" 2>/dev/null || true
sleep 2

kill -TERM "${agent_pid}"
wait "${agent_pid}"
agent_pid=""

if [[ "${output_format}" == "json" ]]; then
    expected_events=(create open rename chmod delete exec)
    for expected in "${expected_events[@]}"; do
        if ! grep -F "\"operation\":\"${expected}\"" "${log_file}" | grep -F "${test_id}" >/dev/null; then
            echo "Missing ${expected} JSON event for ${test_id}" >&2
            cat "${log_file}" >&2
            exit 1
        fi
    done
    exec_count="$(grep -F "\"operation\":\"exec\"" "${log_file}" | \
        grep -F "\"path\":\"${executable_path}\"" | wc -l)"
else
    expected_events=("[CREATE]" "[OPEN]" "[RENAME]" "[CHMOD]" "[DELETE]" "[EXEC]")
    for expected in "${expected_events[@]}"; do
        if ! grep -F "${expected}" "${log_file}" | grep -F "${test_id}" >/dev/null; then
            echo "Missing ${expected} event for ${test_id}" >&2
            cat "${log_file}" >&2
            exit 1
        fi
    done
    exec_count="$(awk -v expected="PATH=${executable_path}" \
        '$1 == "[EXEC]" && $NF == expected { count++ } END { print count + 0 }' "${log_file}")"
fi
if [[ "${exec_count}" -lt 2 ]]; then
    echo "Expected execve and execveat events for ${executable_path}, got ${exec_count}" >&2
    cat "${log_file}" >&2
    exit 1
fi

if [[ -n "${config}" ]] && ! grep -F "rules=[execution-from-temporary-directory]" "${log_file}" >/dev/null; then
    echo "Missing temporary-directory execution policy decision" >&2
    cat "${log_file}" >&2
    exit 1
fi

grep -F "${test_id}" "${log_file}"
echo
echo "Agent lifecycle:"
grep -E "Attaching|Listening|Reader closed|unavailable|lost|error|fatal" "${log_file}" || true
if [[ -n "${config}" ]]; then
    echo
    echo "Execution policy:"
    grep -F "rules=[execution-from-temporary-directory]" "${log_file}"
fi
echo
echo "KubeFIM end-to-end smoke test passed."
