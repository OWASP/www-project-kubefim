package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"kubefim/internal/collector"
	"kubefim/internal/enrichment"
	"kubefim/internal/output"
	"kubefim/internal/policy"
)

// Logger is the logging surface needed by the agent.
type Logger interface {
	Printf(format string, args ...any)
	Println(args ...any)
}

// Agent coordinates event collection and output.
type Agent struct {
	collector collector.Collector
	output    output.Output
	policy    policy.Decider
	enricher  enrichment.Enricher
	observer  Observer
	logger    Logger
}

func New(eventCollector collector.Collector, eventOutput output.Output, decider policy.Decider, logger Logger, enrichers ...enrichment.Enricher) *Agent {
	var eventEnricher enrichment.Enricher = enrichment.None{}
	if len(enrichers) > 0 && enrichers[0] != nil {
		eventEnricher = enrichers[0]
	}
	return NewWithOptions(eventCollector, eventOutput, decider, logger, Options{Enricher: eventEnricher})
}

// Options supplies optional event-pipeline integrations.
type Options struct {
	Enricher enrichment.Enricher
	Observer Observer
}

// NewWithOptions creates an Agent with explicit optional integrations.
func NewWithOptions(eventCollector collector.Collector, eventOutput output.Output, decider policy.Decider, logger Logger, options Options) *Agent {
	eventEnricher := options.Enricher
	if eventEnricher == nil {
		eventEnricher = enrichment.None{}
	}
	observer := options.Observer
	if observer == nil {
		observer = noopObserver{}
	}
	return &Agent{
		collector: eventCollector,
		output:    eventOutput,
		policy:    decider,
		enricher:  eventEnricher,
		observer:  observer,
		logger:    logger,
	}
}

// Run processes events until the context is cancelled or an unrecoverable
// output error occurs.
func (a *Agent) Run(ctx context.Context) error {
	defer a.collector.Close()
	var emitted, suppressed, wouldSuppress, protected uint64
	loggedExclusions := make(map[string]struct{})
	defer func() {
		a.logger.Printf(
			"policy summary emitted=%d suppressed=%d would_suppress=%d protected=%d",
			emitted, suppressed, wouldSuppress, protected,
		)
	}()

	runDone := make(chan struct{})
	defer close(runDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = a.collector.Close()
		case <-runDone:
		}
	}()

	a.logger.Println("Listening for events...")
	for {
		record, err := a.collector.Read()
		if err != nil {
			if errors.Is(err, collector.ErrClosed) && ctx.Err() != nil {
				a.logger.Println("Reader closed.")
				return nil
			}
			a.logger.Printf("reading event: %v", err)
			a.observer.ObserveReadError()
			continue
		}

		if record.LostSamples > 0 {
			a.logger.Printf("lost %d samples", record.LostSamples)
			a.observer.ObserveLostSamples(record.LostSamples)
			continue
		}
		// Enrichment and output perform file operations too. Dropping only this
		// daemon's TGID prevents feedback without trusting a process name.
		if record.Event.TGID == uint32(os.Getpid()) {
			continue
		}

		record.Event = a.enricher.Enrich(ctx, record.Event)
		decision := a.policy.Decide(record.Event)
		if decision.Protected {
			protected++
		}
		if decision.WouldSuppress {
			wouldSuppress++
		}
		logDecision := len(decision.MatchedRules) > 0
		if decision.WouldSuppress {
			if _, logged := loggedExclusions[decision.SuppressionRule]; logged {
				logDecision = false
			} else {
				loggedExclusions[decision.SuppressionRule] = struct{}{}
			}
		}
		if logDecision {
			a.logger.Printf(
				"policy decision action=%s class=%s protected=%t exception=%t would_suppress=%t suppressed=%t rules=%v reason=%q",
				decision.Action,
				decision.Class,
				decision.Protected,
				decision.ExceptionApplied,
				decision.WouldSuppress,
				decision.Suppressed,
				decision.MatchedRules,
				decision.Explanation,
			)
		}
		if decision.Suppressed {
			suppressed++
			a.observer.ObserveEvent(record.Event, decision, false)
			continue
		}

		if err := a.output.Write(record.Event); err != nil {
			a.observer.ObserveEvent(record.Event, decision, false)
			a.observer.ObserveOutputError()
			return fmt.Errorf("write event: %w", err)
		}
		emitted++
		a.observer.ObserveEvent(record.Event, decision, true)

		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}
