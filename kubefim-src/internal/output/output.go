package output

import "kubefim/internal/event"

// Output consumes an enriched filesystem event.
type Output interface {
	Write(event.Event) error
}
