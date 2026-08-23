package event

const SchemaVersion uint32 = 1

// Event is the userspace representation of one kernel event.
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
	Container       Container
	Kubernetes      Kubernetes
}

// Container describes the runtime identity resolved from the process cgroup.
// Host is true only when the process was positively identified as host-owned.
type Container struct {
	ID      string
	Runtime string
	Host    bool
}

// Kubernetes contains pod metadata resolved from a node-local cache.
type Kubernetes struct {
	Node          string
	Namespace     string
	PodName       string
	PodUID        string
	ContainerName string
	Image         string
	ImageID       string
}

func (e Event) Successful() bool {
	return e.ReturnValue >= 0
}
