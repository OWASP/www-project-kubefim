package event

import "fmt"

// Operation identifies the filesystem operation reported by the eBPF program.
type Operation uint32

const (
	OperationOpen   Operation = 1
	OperationCreate Operation = 2
	OperationDelete Operation = 3
	OperationRename Operation = 4
	OperationChmod  Operation = 5
)

func (o Operation) String() string {
	switch o {
	case OperationOpen:
		return "OPEN"
	case OperationCreate:
		return "CREATE"
	case OperationDelete:
		return "DELETE"
	case OperationRename:
		return "RENAME"
	case OperationChmod:
		return "CHMOD"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", uint32(o))
	}
}
