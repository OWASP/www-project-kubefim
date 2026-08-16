package output

import (
	"bytes"
	"testing"

	"kubefim/internal/event"
)

func TestTextWrite(t *testing.T) {
	var destination bytes.Buffer
	writer := NewText(&destination)

	err := writer.Write(event.Event{
		PID:       42,
		UID:       1000,
		Comm:      "touch",
		Path:      "/tmp/example",
		Operation: event.OperationCreate,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "[CREATE] PID=42 UID=1000 COMM=touch PATH=/tmp/example\n"
	if got := destination.String(); got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}
