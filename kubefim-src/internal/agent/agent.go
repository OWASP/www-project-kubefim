package agent

import (
	"context"
	"errors"
	"fmt"

	"kubefim/internal/collector"
	"kubefim/internal/output"
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
	logger    Logger
}

func New(eventCollector collector.Collector, eventOutput output.Output, logger Logger) *Agent {
	return &Agent{collector: eventCollector, output: eventOutput, logger: logger}
}

// Run processes events until the context is cancelled or an unrecoverable
// output error occurs.
func (a *Agent) Run(ctx context.Context) error {
	defer a.collector.Close()

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
			continue
		}

		if record.LostSamples > 0 {
			a.logger.Printf("lost %d samples", record.LostSamples)
			continue
		}

		if err := a.output.Write(record.Event); err != nil {
			return fmt.Errorf("write event: %w", err)
		}

		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}
