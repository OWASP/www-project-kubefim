package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"kubefim/internal/event"
	"kubefim/internal/policy"
)

const operationCount = int(event.OperationExec) + 1

// Registry stores fixed-cardinality Prometheus counters for one KubeFIM agent.
// Event-controlled strings such as paths and workload identities are never
// used as labels.
type Registry struct {
	events        [operationCount][2]atomic.Uint64
	emitted       atomic.Uint64
	suppressed    atomic.Uint64
	wouldSuppress atomic.Uint64
	protected     atomic.Uint64
	lostSamples   atomic.Uint64
	readErrors    atomic.Uint64
	outputErrors  atomic.Uint64
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) ObserveEvent(value event.Event, decision policy.Decision, emitted bool) {
	operation := int(value.Operation)
	if operation >= 0 && operation < operationCount {
		result := 0
		if value.Successful() {
			result = 1
		}
		r.events[operation][result].Add(1)
	}
	if emitted {
		r.emitted.Add(1)
	}
	if decision.Suppressed {
		r.suppressed.Add(1)
	}
	if decision.WouldSuppress {
		r.wouldSuppress.Add(1)
	}
	if decision.Protected {
		r.protected.Add(1)
	}
}

func (r *Registry) ObserveLostSamples(count uint64) { r.lostSamples.Add(count) }
func (r *Registry) ObserveReadError()               { r.readErrors.Add(1) }
func (r *Registry) ObserveOutputError()             { r.outputErrors.Add(1) }

func (r *Registry) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", r.serveMetrics)
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(writer, "ok\n")
	})
	return mux
}

func (r *Registry) serveMetrics(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	writeCounterHeader(writer, "kubefim_events_total", "Events processed by operation and syscall result.")
	for operation := event.OperationOpen; operation <= event.OperationExec; operation++ {
		for result, label := range []string{"failure", "success"} {
			_, _ = fmt.Fprintf(writer, "kubefim_events_total{operation=%q,result=%q} %d\n",
				operation.String(), label, r.events[int(operation)][result].Load())
		}
	}
	writeCounter(writer, "kubefim_events_emitted_total", "Events written to the configured output.", r.emitted.Load())
	writeCounter(writer, "kubefim_events_suppressed_total", "Events omitted by enforced policy exclusions.", r.suppressed.Load())
	writeCounter(writer, "kubefim_events_would_suppress_total", "Events matching exclusions in observe or enforce mode.", r.wouldSuppress.Load())
	writeCounter(writer, "kubefim_events_protected_total", "Events retained because their path is protected.", r.protected.Load())
	writeCounter(writer, "kubefim_lost_samples_total", "Kernel events reported lost by the eBPF transport.", r.lostSamples.Load())
	writeCounter(writer, "kubefim_collector_read_errors_total", "Errors returned while reading collector records.", r.readErrors.Load())
	writeCounter(writer, "kubefim_output_errors_total", "Errors returned while writing event output.", r.outputErrors.Load())
}

func writeCounterHeader(writer io.Writer, name, help string) {
	_, _ = fmt.Fprintf(writer, "# HELP %s %s\n# TYPE %s counter\n", name, help, name)
}

func writeCounter(writer io.Writer, name, help string, value uint64) {
	writeCounterHeader(writer, name, help)
	_, _ = fmt.Fprintf(writer, "%s %d\n", name, value)
}
