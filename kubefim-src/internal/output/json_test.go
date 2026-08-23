package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"kubefim/internal/event"
)

func TestJSONWriteRename(t *testing.T) {
	var destination bytes.Buffer
	writer := NewJSON(&destination)

	err := writer.Write(event.Event{
		SchemaVersion:   1,
		TimestampNS:     123456789,
		CgroupID:        88,
		ReturnValue:     0,
		PID:             42,
		TGID:            40,
		PPID:            10,
		UID:             1000,
		GID:             1001,
		MountNamespace:  4026532000,
		PIDNamespace:    4026532001,
		Comm:            "mv",
		Path:            "/tmp/source",
		DestinationPath: "/tmp/destination",
		Operation:       event.OperationRename,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `{"api_version":"events.kubefim.org/v1alpha1","schema_version":1,"timestamp_ns":123456789,"operation":"rename","success":true,"return_value":0,"process":{"pid":42,"tgid":40,"ppid":10,"uid":1000,"gid":1001,"comm":"mv"},"target":{"path":"/tmp/source","destination_path":"/tmp/destination"},"linux":{"cgroup_id":88,"mount_namespace_id":4026532000,"pid_namespace_id":4026532001}}` + "\n"
	if got := destination.String(); got != want {
		t.Fatalf("JSON output is %q, want %q", got, want)
	}
}

func TestJSONWriteFailedExec(t *testing.T) {
	var destination bytes.Buffer
	writer := NewJSON(&destination)

	err := writer.Write(event.Event{
		SchemaVersion: 1,
		ReturnValue:   -13,
		PID:           50,
		TGID:          50,
		PPID:          49,
		UID:           1000,
		GID:           1000,
		Comm:          "node\nworker",
		Path:          "/tmp/setup",
		Operation:     event.OperationExec,
	})
	if err != nil {
		t.Fatal(err)
	}

	line := destination.String()
	if strings.Count(line, "\n") != 1 {
		t.Fatalf("JSON Lines record contains an unescaped newline: %q", line)
	}

	var got jsonEvent
	if err := json.Unmarshal(destination.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if got.Operation != "exec" || got.Success || got.ReturnValue != -13 {
		t.Fatalf("unexpected failed execution record: %+v", got)
	}
	if got.Process.Comm != "node\nworker" || got.Target.Path != "/tmp/setup" {
		t.Fatalf("unexpected process or target: %+v", got)
	}
	if got.Target.DestinationPath != "" {
		t.Fatalf("unexpected destination path %q", got.Target.DestinationPath)
	}
	if strings.Contains(line, "destination_path") {
		t.Fatalf("empty destination path was not omitted: %s", line)
	}
}

func TestJSONWriteKubernetesAttribution(t *testing.T) {
	var destination bytes.Buffer
	writer := NewJSON(&destination)
	value := event.Event{
		SchemaVersion: 1, Operation: event.OperationExec, ReturnValue: 0,
		Container:  event.Container{ID: strings.Repeat("a", 64), Runtime: "containerd"},
		Kubernetes: event.Kubernetes{Node: "worker-1", Namespace: "production", PodName: "api-123", PodUID: "pod-uid", ContainerName: "api", Image: "example/api:v1", ImageID: "sha256:image"},
	}
	if err := writer.Write(value); err != nil {
		t.Fatal(err)
	}
	var got jsonEvent
	if err := json.Unmarshal(destination.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Container == nil || got.Container.Runtime != "containerd" || got.Container.Host {
		t.Fatalf("container attribution = %+v", got.Container)
	}
	if got.Kubernetes == nil || got.Kubernetes.Namespace != "production" || got.Kubernetes.ContainerName != "api" {
		t.Fatalf("Kubernetes attribution = %+v", got.Kubernetes)
	}
}

func TestNewSelectsFormat(t *testing.T) {
	tests := []struct {
		format string
		want   any
	}{
		{format: "text", want: &Text{}},
		{format: " JSON ", want: &JSON{}},
	}

	for _, test := range tests {
		got, err := New(test.format, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("New(%q): %v", test.format, err)
		}
		switch test.want.(type) {
		case *Text:
			if _, ok := got.(*Text); !ok {
				t.Fatalf("New(%q) returned %T, want *Text", test.format, got)
			}
		case *JSON:
			if _, ok := got.(*JSON); !ok {
				t.Fatalf("New(%q) returned %T, want *JSON", test.format, got)
			}
		}
	}
}

func TestNewRejectsUnsupportedFormat(t *testing.T) {
	if _, err := New("xml", &bytes.Buffer{}); err == nil {
		t.Fatal("New accepted an unsupported output format")
	}
}
