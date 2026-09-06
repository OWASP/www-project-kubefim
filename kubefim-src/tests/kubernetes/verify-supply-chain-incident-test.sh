#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
fixture=$(mktemp)
trap 'rm -f "$fixture"' EXIT HUP INT TERM

event() {
    printf '%s\n' '{"operation":"'"$1"'","target":{"path":"'"$2"'"},"kubernetes":{"namespace":"kubefim-demo","pod_name":"supply-chain-demo","container_name":"compromised-app"}}' >>"$fixture"
}

event open /root/.aws/credentials
event open /root/.npmrc
event open /app/.env
event create /tmp/.update-agent
event chmod /tmp/.update-agent
event exec /tmp/.update-agent
event create /etc/cron.d/system-update
event delete /tmp/.update-agent

for rule in cloud-credential-access temporary-payload-staging execution-from-temporary-directory cron-persistence-attempt; do
    printf '{"detection":{"action":"alert","matched_rules":["%s"]}}\n' "$rule" >>"$fixture"
done

"$script_dir/verify-supply-chain-incident.sh" "$fixture"

printf '%s\n' '{"operation":"delete","target":{"path":"/tmp/.update-agent"}}' >"$fixture"
if "$script_dir/verify-supply-chain-incident.sh" "$fixture" >/dev/null 2>&1; then
    echo "incident verifier accepted incomplete evidence" >&2
    exit 1
fi

echo "KubeFIM supply-chain incident verifier tests passed."
