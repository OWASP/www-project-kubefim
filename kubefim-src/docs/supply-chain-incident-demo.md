# Testing a compromised package install

This test shows what KubeFIM records when a package install behaves like a
supply-chain attack. The workload is harmless. It uses fake credentials, does
not make any network requests, and never writes to the Kubernetes node.

The test Pod does the following inside its own container:

1. Opens fake AWS, npm, and application credential files.
2. Copies `/bin/true` to `/tmp/.update-agent` as a stand-in for a dropped binary.
3. Makes the copied file executable and runs it.
4. Creates a file under `/etc/cron.d` to simulate a persistence attempt.
5. Deletes the copied file.

The cron file contains only a comment. Nothing is scheduled or executed from
it. The values in the credential files are all `DECOY_ONLY`.

## Run the test

This test uses the e2e policy overlay. Start or update KubeFIM first:

```sh
kubectl apply -k deployments/kubernetes/overlays/e2e
kubectl rollout status daemonset/kubefim -n kubefim-system
```

Run the workload:

```sh
kubectl apply -f tests/kubernetes/supply-chain-incident.yaml
```

Give the workload a few seconds to finish the test actions, then collect the
KubeFIM log:

```sh
kubectl logs -n kubefim-system daemonset/kubefim -c kubefim > kubefim.log
tests/kubernetes/verify-supply-chain-incident.sh kubefim.log
```

A successful run prints:

```text
KubeFIM supply-chain incident simulation passed.
```

To view the same evidence in Grafana, start the development stack and load the
log into Loki:

```sh
docker compose -f deployments/observability/docker-compose.yaml up -d
scripts/push-logs-to-loki.sh kubefim.log
```

Open `http://127.0.0.1:3000/d/kubefim-security` and select `kubefim-demo` from
the namespace filter.

## Events checked by the test

The verifier expects all of these events to come from the `supply-chain-demo`
Pod and its `compromised-app` container:

```text
OPEN    /root/.aws/credentials
OPEN    /root/.npmrc
OPEN    /app/.env
CREATE  /tmp/.update-agent
CHMOD   /tmp/.update-agent
EXEC    /tmp/.update-agent
CREATE  /etc/cron.d/system-update
DELETE  /tmp/.update-agent
```

The e2e policy has rules for credential access, temporary payload staging,
execution from `/tmp`, and cron persistence. The cron file also falls under the
built-in `/etc/**` protected path.

## What this tells us

These events are useful during an investigation because they answer the basic
questions: which file was touched, which process did it, and which Pod,
container, image, namespace, and node were involved.

KubeFIM currently reports the events and policy matches separately. It does not
yet combine this sequence into one incident or claim that the process is known
malware. That correlation work is the next part of the detection engine.

## Cleanup

```sh
kubectl delete -f tests/kubernetes/supply-chain-incident.yaml
```
