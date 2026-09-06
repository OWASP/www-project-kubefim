package event

const SchemaVersion uint32 = 2

const ATFDCWD int32 = -100

// Event is the userspace representation of one kernel event.
type Event struct {
	SchemaVersion             uint32
	TimestampNS               uint64
	CgroupID                  uint64
	ReturnValue               int64
	PID                       uint32
	TGID                      uint32
	PPID                      uint32
	UID                       uint32
	GID                       uint32
	MountNamespace            uint32
	PIDNamespace              uint32
	Comm                      string
	Path                      string
	DestinationPath           string
	DirFD                     int32
	DestinationDirFD          int32
	BaseInode                 uint64
	BaseDevice                uint64
	DestinationBaseInode      uint64
	DestinationBaseDevice     uint64
	ResolvedPath              string
	ResolvedDestinationPath   string
	PathResolution            PathResolution
	DestinationPathResolution PathResolution
	Operation                 Operation
	Detection                 Detection
	Container                 Container
	Kubernetes                Kubernetes
}

type Detection struct {
	Action           string
	Class            string
	Protected        bool
	ExceptionApplied bool
	MatchedRules     []string
	Reason           string
}

type PathResolution string

const (
	PathResolutionAbsolute      PathResolution = "absolute"
	PathResolutionResolved      PathResolution = "resolved"
	PathResolutionUnresolved    PathResolution = "unresolved"
	PathResolutionProcessExited PathResolution = "process_exited"
	PathResolutionFDUnavailable PathResolution = "fd_unavailable"
	PathResolutionBaseChanged   PathResolution = "base_changed"
)

func (e Event) PolicyPath() string {
	if e.ResolvedPath != "" {
		return e.ResolvedPath
	}
	if e.PathResolution == PathResolutionAbsolute || (len(e.Path) > 0 && e.Path[0] == '/') {
		return e.Path
	}
	return ""
}

func (e Event) PolicyDestinationPath() string {
	if e.ResolvedDestinationPath != "" {
		return e.ResolvedDestinationPath
	}
	if e.DestinationPathResolution == PathResolutionAbsolute || (len(e.DestinationPath) > 0 && e.DestinationPath[0] == '/') {
		return e.DestinationPath
	}
	return ""
}

func (e Event) HasUnresolvedPath() bool {
	if e.Path != "" && e.PathResolution != "" && e.PolicyPath() == "" {
		return true
	}
	return e.DestinationPath != "" && e.DestinationPathResolution != "" && e.PolicyDestinationPath() == ""
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
