package policy

import (
	"fmt"
	"strings"

	"kubefim/internal/event"
)

const (
	APIVersion = "policy.kubefim.org/v1alpha1"
	Kind       = "KubeFIMPolicy"
)

type Action uint8

const (
	ActionAggregate Action = 1
	ActionAudit     Action = 2
	ActionAlert     Action = 3
)

func (a Action) String() string {
	switch a {
	case ActionAggregate:
		return "aggregate"
	case ActionAudit:
		return "audit"
	case ActionAlert:
		return "alert"
	default:
		return fmt.Sprintf("unknown(%d)", a)
	}
}

func parseAction(value string) (Action, error) {
	switch strings.ToLower(value) {
	case "aggregate":
		return ActionAggregate, nil
	case "audit":
		return ActionAudit, nil
	case "alert":
		return ActionAlert, nil
	default:
		return 0, fmt.Errorf("unsupported action %q", value)
	}
}

type Class string

const (
	ClassAccess   Class = "access"
	ClassMutation Class = "mutation"
)

func Classify(value event.Event) Class {
	if value.Operation == event.OperationOpen {
		return ClassAccess
	}
	return ClassMutation
}

// Decision explains how a policy evaluated an event. Observe mode sets
// WouldSuppress without setting Suppressed, so the event still reaches output.
type Decision struct {
	Action           Action
	Class            Class
	Protected        bool
	ExceptionApplied bool
	WouldSuppress    bool
	Suppressed       bool
	SuppressionRule  string
	MatchedRules     []string
	Explanation      string
}

type Decider interface {
	Decide(event.Event) Decision
}
