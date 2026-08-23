package event

import "testing"

func TestOperationString(t *testing.T) {
	tests := []struct {
		operation Operation
		want      string
	}{
		{OperationOpen, "OPEN"},
		{OperationCreate, "CREATE"},
		{OperationDelete, "DELETE"},
		{OperationRename, "RENAME"},
		{OperationChmod, "CHMOD"},
		{OperationExec, "EXEC"},
		{Operation(99), "UNKNOWN(99)"},
	}

	for _, test := range tests {
		if got := test.operation.String(); got != test.want {
			t.Errorf("Operation(%d).String() is %q, want %q", test.operation, got, test.want)
		}
	}
}
