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
        grep -F '"namespace":"kubefim-demo"' |
        grep -F '"pod_name":"supply-chain-demo"' |
        grep -F '"container_name":"compromised-app"' >/dev/null; then
        echo "missing enriched $operation event for $path" >&2
        exit 1
    fi
}

require_rule() {
    if ! grep -F '"detection":{"action":"alert"' "$log_file" |
        grep -F "$1" >/dev/null; then
        echo "missing policy alert for $1" >&2
        exit 1
    fi
}

require_event open /root/.aws/credentials
require_event open /root/.npmrc
require_event open /app/.env
require_event create /tmp/.update-agent
require_event chmod /tmp/.update-agent
require_event exec /tmp/.update-agent
require_event create /etc/cron.d/system-update
require_event delete /tmp/.update-agent

require_rule cloud-credential-access
require_rule temporary-payload-staging
require_rule execution-from-temporary-directory
require_rule cron-persistence-attempt

echo "KubeFIM supply-chain incident simulation passed."
