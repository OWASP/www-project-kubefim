package output

import (
	"fmt"
	"io"

	"kubefim/internal/event"
)

// Text writes the development-friendly KubeFIM log format.
type Text struct {
	writer io.Writer
}

func NewText(writer io.Writer) *Text {
	return &Text{writer: writer}
}

func (t *Text) Write(event event.Event) error {
	_, err := fmt.Fprintf(
		t.writer,
		"[%s] PID=%d UID=%d COMM=%s PATH=%s\n",
		event.Operation,
		event.PID,
		event.UID,
		event.Comm,
		event.Path,
	)
	return err
}
