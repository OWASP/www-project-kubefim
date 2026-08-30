package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kubefim/internal/event"
	"kubefim/internal/policy"
)

func TestMetricsExposeBoundedCounters(t *testing.T) {
	registry := NewRegistry()
	registry.ObserveEvent(event.Event{Operation: event.OperationOpen, ReturnValue: 3}, policy.Decision{
		WouldSuppress: true, Suppressed: true,
	}, false)
	registry.ObserveEvent(event.Event{Operation: event.OperationDelete, ReturnValue: -1}, policy.Decision{
		Protected: true,
	}, true)
	registry.ObserveLostSamples(7)
	registry.ObserveReadError()
	registry.ObserveOutputError()

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	response := httptest.NewRecorder()
	registry.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status is %d", response.Code)
	}
	body := response.Body.String()
	for _, want := range []string{
		`kubefim_events_total{operation="OPEN",result="success"} 1`,
		`kubefim_events_total{operation="DELETE",result="failure"} 1`,
		"kubefim_events_emitted_total 1",
		"kubefim_events_suppressed_total 1",
		"kubefim_events_would_suppress_total 1",
		"kubefim_events_protected_total 1",
		"kubefim_lost_samples_total 7",
		"kubefim_collector_read_errors_total 1",
		"kubefim_output_errors_total 1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics missing %q\n%s", want, body)
		}
	}
	for _, forbidden := range []string{"path=", "namespace=", "pod=", "container="} {
		if strings.Contains(body, forbidden) {
			t.Errorf("unbounded label %q found in metrics", forbidden)
		}
	}
	if strings.Contains(body, "UNKNOWN") {
		t.Errorf("unknown zero-value operation was exported\n%s", body)
	}
}

func TestHealthEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	NewRegistry().Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "ok\n" {
		t.Fatalf("health response is %d %q", response.Code, response.Body.String())
	}
}

func TestMetricsRejectPost(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	response := httptest.NewRecorder()
	NewRegistry().Handler().ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status is %d", response.Code)
	}
}
