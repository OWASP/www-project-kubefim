package collector

import (
	"errors"

	"kubefim/internal/event"
)

var ErrClosed = errors.New("collector closed")

// Record represents either a decoded event or a kernel lost-sample report.
type Record struct {
	Event       event.Event
	LostSamples uint64
}

// Collector is the event source used by the agent.
type Collector interface {
	Read() (Record, error)
	Close() error
}
