package policy

import (
	"strings"
	"testing"
	"time"

	"kubefim/internal/event"
)

var testNow = time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

func TestDefaultClassification(t *testing.T) {
	evaluator := Default()

	access := evaluator.Decide(event.Event{Operation: event.OperationOpen})
	if access.Class != ClassAccess || access.Action != ActionAggregate {
		t.Fatalf("access decision is %+v", access)
	}

	mutation := evaluator.Decide(event.Event{Operation: event.OperationDelete})
	if mutation.Class != ClassMutation || mutation.Action != ActionAudit {
		t.Fatalf("mutation decision is %+v", mutation)
	}
}

func TestProtectedDestinationPathCannotBeSuppressed(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  protectedPaths: [/etc/**]
  exceptions:
    - id: maintenance
      match:
        operations: [rename]
        pathPrefixes: [/etc/**]
        comms: [mv]
        uids: [0]
      reason: maintenance
      owner: security
      expires: "2027-01-01T00:00:00Z"
`)

	decision := evaluator.Decide(event.Event{
		Operation: event.OperationRename, DestinationPath: "/etc/renamed", Comm: "mv", UID: 0,
	})
	if decision.Action != ActionAlert || !decision.Protected || decision.Suppressed {
		t.Fatalf("protected decision is %+v", decision)
	}
}

func TestRuleAndNarrowException(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  rules:
    - id: executable-change
      match:
        operations: [rename]
        pathPrefixes: [/opt/example/**]
      action: alert
      reason: executable changed
      owner: security
  exceptions:
    - id: maintenance
      match:
        operations: [rename]
        pathPrefixes: [/opt/example/**]
        comms: [maintenance]
        uids: [0]
      reason: approved maintenance
      owner: platform
      expires: "2027-01-01T00:00:00Z"
`)

	decision := evaluator.Decide(event.Event{
		Operation: event.OperationRename, Path: "/opt/example/a", Comm: "maintenance", UID: 0,
	})
	if decision.Action != ActionAudit || !decision.Suppressed {
		t.Fatalf("exception decision is %+v", decision)
	}
	if len(decision.MatchedRules) != 2 {
		t.Fatalf("matched rules are %v", decision.MatchedRules)
	}
}

func TestExceptionStopsApplyingAfterExpiry(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  rules:
    - id: executable-change
      match:
        operations: [rename]
        pathPrefixes: [/opt/example/**]
      action: alert
      reason: executable changed
      owner: security
  exceptions:
    - id: maintenance
      match:
        operations: [rename]
        pathPrefixes: [/opt/example/**]
        comms: [maintenance]
        uids: [0]
      reason: approved maintenance
      owner: platform
      expires: "2027-01-01T00:00:00Z"
`)
	evaluator.now = func() time.Time {
		return time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC)
	}

	decision := evaluator.Decide(event.Event{
		Operation: event.OperationRename, Path: "/opt/example/a", Comm: "maintenance", UID: 0,
	})
	if decision.Action != ActionAlert || decision.Suppressed {
		t.Fatalf("expired exception decision is %+v", decision)
	}
}

func TestRejectsUnknownField(t *testing.T) {
	_, err := Decode(strings.NewReader(`
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  surprise: true
`), testNow)
	if err == nil {
		t.Fatal("policy with unknown field was accepted")
	}
}

func TestRejectsUnsafeException(t *testing.T) {
	_, err := Decode(strings.NewReader(`
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  exceptions:
    - id: too-broad
      match:
        comms: [maintenance]
      reason: broad exception
      owner: platform
      expires: "2027-01-01T00:00:00Z"
`), testNow)
	if err == nil || !strings.Contains(err.Error(), "must match operation, path, comm, and UID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectsExpiredExceptionAtLoad(t *testing.T) {
	_, err := Decode(strings.NewReader(`
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  exceptions:
    - id: expired
      match:
        operations: [rename]
        pathPrefixes: [/opt/example/**]
        comms: [maintenance]
        uids: [0]
      reason: expired exception
      owner: platform
      expires: "2026-01-01T00:00:00Z"
`), testNow)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustDecode(t *testing.T, input string) *Evaluator {
	t.Helper()
	evaluator, err := Decode(strings.NewReader(input), testNow)
	if err != nil {
		t.Fatal(err)
	}
	return evaluator
}
