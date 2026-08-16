package event

const SchemaVersion uint32 = 1

// Event is the userspace representation of one filesystem event.
type Event struct {
	SchemaVersion   uint32
	TimestampNS     uint64
	CgroupID        uint64
	ReturnValue     int64
	PID             uint32
	TGID            uint32
	PPID            uint32
	UID             uint32
	GID             uint32
	MountNamespace  uint32
	PIDNamespace    uint32
	Comm            string
	Path            string
	DestinationPath string
	Operation       Operation
}

func (e Event) Successful() bool {
	return e.ReturnValue >= 0
}
