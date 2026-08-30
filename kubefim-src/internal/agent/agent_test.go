package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"testing"

	"kubefim/internal/collector"
	"kubefim/internal/event"
	"kubefim/internal/policy"
)

type fakeCollector struct {
	records   []collector.Record
	closed    chan struct{}
	closeOnce sync.Once
}

func newFakeCollector(records ...collector.Record) *fakeCollector {
	return &fakeCollector{records: records, closed: make(chan struct{})}
}

func (f *fakeCollector) Read() (collector.Record, error) {
	if len(f.records) > 0 {
		record := f.records[0]
		f.records = f.records[1:]
		return record, nil
	}
	<-f.closed
	return collector.Record{}, collector.ErrClosed
}

func (f *fakeCollector) Close() error {
	f.closeOnce.Do(func() { close(f.closed) })
	return nil
}

type outputFunc func(event.Event) error

func (f outputFunc) Write(value event.Event) error { return f(value) }

type deciderFunc func(event.Event) policy.Decision

func (f deciderFunc) Decide(value event.Event) policy.Decision { return f(value) }

func TestRunWritesEventAndStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	want := event.Event{PID: 42, UID: 1000, Comm: "touch", Path: "/tmp/example", Operation: event.OperationCreate}
	source := newFakeCollector(
		collector.Record{LostSamples: 3},
		collector.Record{Event: want},
	)

	var got event.Event
	var logs strings.Builder
	application := New(source, outputFunc(func(value event.Event) error {
		got = value
		cancel()
		return nil
	}), policy.Default(), log.New(&logs, "", 0))

	if err := application.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("output event is %+v, want %+v", got, want)
	}
	if !strings.Contains(logs.String(), "lost 3 samples") {
		t.Fatalf("lost-sample log missing from %q", logs.String())
	}
}

func TestRunDoesNotWriteSuppressedEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	suppressedEvent := event.Event{PID: 1, Operation: event.OperationOpen, Path: "/usr/lib/a.so"}
	secondSuppressedEvent := event.Event{PID: 3, Operation: event.OperationOpen, Path: "/usr/lib/b.so"}
	emittedEvent := event.Event{PID: 2, Operation: event.OperationDelete, Path: "/tmp/a"}
	source := newFakeCollector(
		collector.Record{Event: suppressedEvent},
		collector.Record{Event: secondSuppressedEvent},
		collector.Record{Event: emittedEvent},
	)

	var written []event.Event
	var logs strings.Builder
	application := New(source, outputFunc(func(value event.Event) error {
		written = append(written, value)
		cancel()
		return nil
	}), deciderFunc(func(value event.Event) policy.Decision {
		if value.Operation == event.OperationOpen {
			return policy.Decision{
				Action: policy.ActionAggregate, Class: policy.ClassAccess,
				WouldSuppress: true, Suppressed: true, SuppressionRule: "library-reads",
				MatchedRules: []string{"library-reads"},
			}
		}
		return policy.Decision{Action: policy.ActionAudit, Class: policy.ClassMutation}
	}), log.New(&logs, "", 0))

	if err := application.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 || written[0] != emittedEvent {
		t.Fatalf("written events are %+v", written)
	}
	if !strings.Contains(logs.String(), "emitted=1 suppressed=2 would_suppress=2") {
		t.Fatalf("policy summary missing from %q", logs.String())
	}
	if count := strings.Count(logs.String(), "policy decision action="); count != 1 {
		t.Fatalf("suppression decision logged %d times in %q", count, logs.String())
	}
}

func TestRunReturnsOutputErrorAndClosesCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wantErr := errors.New("output unavailable")
	source := newFakeCollector(collector.Record{Event: event.Event{Operation: event.OperationOpen}})
	application := New(source, outputFunc(func(event.Event) error {
		return wantErr
	}), policy.Default(), log.New(io.Discard, "", 0))

	err := application.Run(ctx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run error is %v, want wrapped %v", err, wantErr)
	}

	select {
	case <-source.closed:
	default:
		t.Fatal("collector was not closed")
	}
}
