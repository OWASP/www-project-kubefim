package event

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// rawEvent must remain byte-for-byte compatible with struct event_t in
// bpf/src/common.h.
type rawEvent struct {
	TimestampNS     uint64
	CgroupID        uint64
	ReturnValue     int64
	SchemaVersion   uint32
	EventType       uint32
	PID             uint32
	TGID            uint32
	PPID            uint32
	UID             uint32
	GID             uint32
	MountNamespace  uint32
	PIDNamespace    uint32
	Comm            [16]byte
	Path            [256]byte
	DestinationPath [256]byte
	Reserved        uint32
}

const RawSize = 592

// Decode converts the fixed-size little-endian kernel payload into an Event.
func Decode(data []byte) (Event, error) {
	if len(data) < RawSize {
		return Event{}, fmt.Errorf("event payload is %d bytes, want at least %d", len(data), RawSize)
	}

	var raw rawEvent
	if err := binary.Read(bytes.NewReader(data[:RawSize]), binary.LittleEndian, &raw); err != nil {
		return Event{}, fmt.Errorf("decode event payload: %w", err)
	}

	if raw.SchemaVersion != SchemaVersion {
		return Event{}, fmt.Errorf("unsupported event schema version %d", raw.SchemaVersion)
	}

	return Event{
		SchemaVersion:   raw.SchemaVersion,
		TimestampNS:     raw.TimestampNS,
		CgroupID:        raw.CgroupID,
		ReturnValue:     raw.ReturnValue,
		PID:             raw.PID,
		TGID:            raw.TGID,
		PPID:            raw.PPID,
		UID:             raw.UID,
		GID:             raw.GID,
		MountNamespace:  raw.MountNamespace,
		PIDNamespace:    raw.PIDNamespace,
		Comm:            trimCString(raw.Comm[:]),
		Path:            trimCString(raw.Path[:]),
		DestinationPath: trimCString(raw.DestinationPath[:]),
		Operation:       Operation(raw.EventType),
	}, nil
}

func trimCString(value []byte) string {
	return strings.TrimRight(string(value), "\x00")
}
