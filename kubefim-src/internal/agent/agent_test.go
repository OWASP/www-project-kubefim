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
	}), log.New(&logs, "", 0))

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

func TestRunReturnsOutputErrorAndClosesCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wantErr := errors.New("output unavailable")
	source := newFakeCollector(collector.Record{Event: event.Event{Operation: event.OperationOpen}})
	application := New(source, outputFunc(func(event.Event) error {
		return wantErr
	}), log.New(io.Discard, "", 0))

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
