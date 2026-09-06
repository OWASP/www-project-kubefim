#!/bin/sh
set -eu

prometheus_url=${PROMETHEUS_URL:-http://127.0.0.1:9090}
grafana_url=${GRAFANA_URL:-http://127.0.0.1:3000}
loki_url=${LOKI_URL:-http://127.0.0.1:3100}

if ! curl -fsS "$loki_url/ready" >/dev/null; then
    echo "Loki is not ready" >&2
    exit 1
fi

prometheus_ready=$(curl -fsS "$prometheus_url/-/ready")
case $prometheus_ready in
    *Ready* | *ready*) ;;
    *)
        echo "Prometheus is not ready: $prometheus_ready" >&2
        exit 1
        ;;
esac

query() {
    curl -fsS -G --data-urlencode "query=$1" "$prometheus_url/api/v1/query"
}

up=$(query 'up{job="kubefim"}' | jq -r '.data.result[0].value[1] // ""')
if [ "$up" != "1" ]; then
    echo "Prometheus is not scraping KubeFIM (up=$up)" >&2
    exit 1
fi

events=$(query 'sum(kubefim_events_total)' | jq -r '.data.result[0].value[1] // ""')
if [ -z "$events" ]; then
    echo "KubeFIM event metrics are absent" >&2
    exit 1
fi

grafana_database=$(curl -fsS "$grafana_url/api/health" | jq -r '.database // ""')
if [ "$grafana_database" != "ok" ]; then
    echo "Grafana database health is $grafana_database" >&2
    exit 1
fi

dashboard_uid=$(curl -fsS "$grafana_url/api/dashboards/uid/kubefim-overview" | jq -r '.dashboard.uid // ""')
if [ "$dashboard_uid" != "kubefim-overview" ]; then
    echo "KubeFIM dashboard was not provisioned" >&2
    exit 1
fi

security_dashboard_uid=$(curl -fsS "$grafana_url/api/dashboards/uid/kubefim-security" | jq -r '.dashboard.uid // ""')
if [ "$security_dashboard_uid" != "kubefim-security" ]; then
    echo "KubeFIM security dashboard was not provisioned" >&2
    exit 1
fi

if ! curl -fsS "$grafana_url/api/datasources/proxy/uid/loki/loki/api/v1/labels" >/dev/null; then
    echo "Grafana cannot query its Loki datasource" >&2
    exit 1
fi

grafana_events=$(curl -fsS -G \
    --data-urlencode 'query=sum(kubefim_events_total)' \
    "$grafana_url/api/datasources/proxy/uid/prometheus/api/v1/query" |
    jq -r '.data.result[0].value[1] // ""')
if [ -z "$grafana_events" ]; then
    echo "Grafana cannot query its Prometheus datasource" >&2
    exit 1
fi

echo "KubeFIM Prometheus and Grafana verification passed (events=$events)."
