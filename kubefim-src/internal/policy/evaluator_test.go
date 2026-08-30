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
	if decision.Action != ActionAudit || !decision.ExceptionApplied || decision.Suppressed {
		t.Fatalf("exception decision is %+v", decision)
	}
	if len(decision.MatchedRules) != 2 {
		t.Fatalf("matched rules are %v", decision.MatchedRules)
	}
}

func TestExclusionIsShadowedInObserveMode(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  mode: observe
  exclusions:
    - id: runtime-library-reads
      match:
        operations: [open]
        pathPrefixes: [/usr/lib/**]
        namespaces: [payments]
        containers: [api]
        images: [example/api:v1]
        success: true
      reason: routine runtime library read
      owner: platform-security
`)

	decision := evaluator.Decide(event.Event{
		Operation:   event.OperationOpen,
		Path:        "/usr/lib/libexample.so",
		ReturnValue: 3,
		Kubernetes: event.Kubernetes{
			Namespace: "payments", ContainerName: "api", Image: "example/api:v1",
		},
	})
	if !decision.WouldSuppress || decision.Suppressed {
		t.Fatalf("observe decision is %+v", decision)
	}
	if len(decision.MatchedRules) != 1 || decision.MatchedRules[0] != "runtime-library-reads" {
		t.Fatalf("matched rules are %v", decision.MatchedRules)
	}
}

func TestExclusionSuppressesOnlyExactEnrichedMatchInEnforceMode(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  mode: enforce
  exclusions:
    - id: runtime-library-reads
      match:
        operations: [open]
        pathPrefixes: [/usr/lib/**]
        namespaces: [payments]
      reason: routine runtime library read
      owner: platform-security
`)

	matched := evaluator.Decide(event.Event{
		Operation: event.OperationOpen, Path: "/usr/lib/a.so", ReturnValue: 3,
		Kubernetes: event.Kubernetes{Namespace: "payments"},
	})
	if !matched.WouldSuppress || !matched.Suppressed {
		t.Fatalf("matching decision is %+v", matched)
	}

	unmatched := evaluator.Decide(event.Event{
		Operation: event.OperationOpen, Path: "/usr/lib/a.so", ReturnValue: 3,
		Kubernetes: event.Kubernetes{Namespace: "orders"},
	})
	if unmatched.WouldSuppress || unmatched.Suppressed {
		t.Fatalf("non-matching decision is %+v", unmatched)
	}
}

func TestExclusionCannotSuppressProtectedOrAlertedAccess(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  mode: enforce
  protectedPaths: [/etc/**]
  rules:
    - id: failed-sensitive-access
      match:
        operations: [open]
        pathPrefixes: [/var/secrets/**]
        success: false
      action: alert
      reason: failed sensitive access
      owner: security
  exclusions:
    - id: broad-open-noise
      match:
        operations: [open]
        pathPrefixes: [/**]
      reason: routine open activity
      owner: platform-security
`)

	protected := evaluator.Decide(event.Event{
		Operation: event.OperationOpen, Path: "/etc/shadow", ReturnValue: 3,
	})
	if !protected.Protected || protected.WouldSuppress || protected.Suppressed {
		t.Fatalf("protected decision is %+v", protected)
	}

	alerted := evaluator.Decide(event.Event{
		Operation: event.OperationOpen, Path: "/var/secrets/token", ReturnValue: -13,
	})
	if alerted.Action != ActionAlert || alerted.WouldSuppress || alerted.Suppressed {
		t.Fatalf("alert decision is %+v", alerted)
	}
}

func TestExceptionDowngradeDoesNotMakeAlertEligibleForExclusion(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  mode: enforce
  rules:
    - id: failed-sensitive-access
      match:
        operations: [open]
        pathPrefixes: [/var/secrets/**]
        success: false
      action: alert
      reason: failed sensitive access
      owner: security
  exceptions:
    - id: approved-probe
      match:
        operations: [open]
        pathPrefixes: [/var/secrets/**]
        comms: [probe]
        uids: [1000]
        success: false
      reason: approved health probe
      owner: platform-security
      expires: "2027-01-01T00:00:00Z"
  exclusions:
    - id: failed-open-noise
      match:
        operations: [open]
        pathPrefixes: [/var/secrets/**]
        success: false
      reason: routine failed open
      owner: platform-security
`)

	decision := evaluator.Decide(event.Event{
		Operation: event.OperationOpen, Path: "/var/secrets/token",
		Comm: "probe", UID: 1000, ReturnValue: -13,
	})
	if decision.Action != ActionAudit || !decision.ExceptionApplied {
		t.Fatalf("exception was not applied: %+v", decision)
	}
	if decision.WouldSuppress || decision.Suppressed {
		t.Fatalf("downgraded alert was suppressed: %+v", decision)
	}
}

func TestRejectsMutationExclusion(t *testing.T) {
	_, err := Decode(strings.NewReader(`
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  mode: enforce
  exclusions:
    - id: unsafe-mutation-filter
      match:
        operations: [delete]
        pathPrefixes: [/tmp/**]
      reason: too broad
      owner: platform-security
`), testNow)
	if err == nil || !strings.Contains(err.Error(), "may only suppress open events") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectsUnscopedExclusion(t *testing.T) {
	_, err := Decode(strings.NewReader(`
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  exclusions:
    - id: all-opens
      match:
        operations: [open]
      reason: too broad
      owner: platform-security
`), testNow)
	if err == nil || !strings.Contains(err.Error(), "must match operation and path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecRule(t *testing.T) {
	evaluator := mustDecode(t, `
apiVersion: policy.kubefim.org/v1alpha1
kind: KubeFIMPolicy
spec:
  rules:
    - id: execution-from-temporary-directory
      match:
        operations: [exec]
        pathPrefixes: [/tmp/**]
        success: true
      action: alert
      reason: executable launched from temporary directory
      owner: security
`)

	decision := evaluator.Decide(event.Event{
		Operation: event.OperationExec, Path: "/tmp/kubefim-test", ReturnValue: 0,
	})
	if decision.Action != ActionAlert || decision.Suppressed {
		t.Fatalf("successful execution decision is %+v", decision)
	}
	if len(decision.MatchedRules) != 1 || decision.MatchedRules[0] != "execution-from-temporary-directory" {
		t.Fatalf("matched rules are %v", decision.MatchedRules)
	}
	if decision.Explanation != "executable launched from temporary directory" {
		t.Fatalf("explanation is %q", decision.Explanation)
	}

	failed := evaluator.Decide(event.Event{
		Operation: event.OperationExec, Path: "/tmp/kubefim-test", ReturnValue: -13,
	})
	if failed.Action != ActionAudit || len(failed.MatchedRules) != 0 {
		t.Fatalf("failed execution decision is %+v", failed)
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
