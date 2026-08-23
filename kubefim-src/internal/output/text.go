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
	workload := ""
	if event.Kubernetes.PodUID != "" {
		workload = fmt.Sprintf(" K8S=%s/%s CONTAINER=%s", event.Kubernetes.Namespace, event.Kubernetes.PodName, event.Kubernetes.ContainerName)
	} else if event.Container.Host {
		workload = " CONTAINER=host"
	} else if event.Container.ID != "" {
		workload = fmt.Sprintf(" CONTAINER=%s:%s", event.Container.Runtime, event.Container.ID)
	}
	_, err := fmt.Fprintf(
		t.writer,
		"[%s] PID=%d UID=%d COMM=%s PATH=%s%s\n",
		event.Operation,
		event.PID,
		event.UID,
		event.Comm,
		event.Path,
		workload,
	)
	return err
}
