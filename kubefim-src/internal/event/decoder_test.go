package event

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDecode(t *testing.T) {
	raw := rawEvent{
		TimestampNS:    123456,
		CgroupID:       789,
		ReturnValue:    4,
		SchemaVersion:  SchemaVersion,
		EventType:      uint32(OperationRename),
		PID:            42,
		TGID:           40,
		PPID:           10,
		UID:            1000,
		GID:            1001,
		MountNamespace: 4026532000,
		PIDNamespace:   4026532001,
	}
	copy(raw.Comm[:], "touch")
	copy(raw.Path[:], "/tmp/source")
	copy(raw.DestinationPath[:], "/tmp/destination")

	var payload bytes.Buffer
	if err := binary.Write(&payload, binary.LittleEndian, raw); err != nil {
		t.Fatal(err)
	}

	got, err := Decode(payload.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion || got.TimestampNS != 123456 || got.CgroupID != 789 ||
		got.ReturnValue != 4 || got.PID != 42 || got.TGID != 40 || got.PPID != 10 ||
		got.UID != 1000 || got.GID != 1001 || got.MountNamespace != 4026532000 ||
		got.PIDNamespace != 4026532001 || got.Comm != "touch" || got.Path != "/tmp/source" ||
		got.DestinationPath != "/tmp/destination" || got.Operation != OperationRename || !got.Successful() {
		t.Fatalf("unexpected event: %+v", got)
	}
}

func TestDecodeRejectsShortPayload(t *testing.T) {
	if _, err := Decode(make([]byte, RawSize-1)); err == nil {
		t.Fatal("Decode accepted a short payload")
	}
}

func TestRawEventSizeMatchesKernelContract(t *testing.T) {
	if got := binary.Size(rawEvent{}); got != RawSize {
		t.Fatalf("raw event size is %d, want %d", got, RawSize)
	}
}

func TestDecodeRejectsUnknownSchema(t *testing.T) {
	raw := rawEvent{SchemaVersion: SchemaVersion + 1}
	var payload bytes.Buffer
	if err := binary.Write(&payload, binary.LittleEndian, raw); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(payload.Bytes()); err == nil {
		t.Fatal("Decode accepted an unsupported schema")
	}
}

func TestEventSuccessful(t *testing.T) {
	if (Event{ReturnValue: -1}).Successful() {
		t.Fatal("negative syscall return value reported as successful")
	}
	if !(Event{ReturnValue: 0}).Successful() {
		t.Fatal("zero syscall return value reported as failed")
	}
}
