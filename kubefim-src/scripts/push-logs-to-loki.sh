#!/bin/sh
set -eu

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo "usage: $0 KUBEFIM_LOG [LOKI_URL]" >&2
    exit 2
fi

log_file=$1
loki_url=${2:-http://127.0.0.1:3100}
payload=$(mktemp)
trap 'rm -f "$payload"' EXIT HUP INT TERM

jq -Rsc '
  [split("\n")[] | fromjson?
    | select(.api_version == "events.kubefim.org/v1alpha1" and .kubernetes != null)
    | tojson]
  | to_entries
  | {streams: [{
      stream: {job: "kubefim", source: "kubernetes"},
      values: map([(((now * 1000000000 | floor) + .key) | tostring), .value])
    }]}
' "$log_file" >"$payload"

curl -fsS -H 'Content-Type: application/json' --data-binary @"$payload" \
    "$loki_url/loki/api/v1/push"
echo "KubeFIM logs loaded into Loki."
