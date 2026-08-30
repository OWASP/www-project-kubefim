package agent

import (
	"kubefim/internal/event"
	"kubefim/internal/policy"
)

// Observer receives bounded operational measurements from the event pipeline.
// Implementations must be safe for use by the agent and HTTP handlers at the
// same time.
type Observer interface {
	ObserveEvent(event.Event, policy.Decision, bool)
	ObserveLostSamples(uint64)
	ObserveReadError()
	ObserveOutputError()
}

type noopObserver struct{}

func (noopObserver) ObserveEvent(event.Event, policy.Decision, bool) {}
func (noopObserver) ObserveLostSamples(uint64)                       {}
func (noopObserver) ObserveReadError()                               {}
func (noopObserver) ObserveOutputError()                             {}
