package enrichment

import (
	"context"

	"kubefim/internal/event"
)

// Enricher adds userspace metadata without changing the kernel event contract.
type Enricher interface {
	Enrich(context.Context, event.Event) event.Event
}

// None leaves events unchanged.
type None struct{}

func (None) Enrich(_ context.Context, value event.Event) event.Event { return value }
