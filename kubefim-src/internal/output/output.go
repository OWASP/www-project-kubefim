package output

import (
	"fmt"
	"io"
	"strings"

	"kubefim/internal/event"
)

// Output consumes a kernel event.
type Output interface {
	Write(event.Event) error
}

func New(format string, writer io.Writer) (Output, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text":
		return NewText(writer), nil
	case "json":
		return NewJSON(writer), nil
	default:
		return nil, fmt.Errorf("unsupported output format %q; supported formats are text and json", format)
	}
}
