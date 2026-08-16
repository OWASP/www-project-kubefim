package policy

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"kubefim/internal/event"
)

type Evaluator struct {
	mode            string
	accessDefault   Action
	mutationDefault Action
	protectedPaths  []pathPrefix
	rules           []rule
	exceptions      []exception
	now             func() time.Time
}

type rule struct {
	id     string
	match  matcher
	action Action
	reason string
}

type exception struct {
	id      string
	match   matcher
	reason  string
	expires time.Time
}

type matcher struct {
	operations map[event.Operation]struct{}
	paths      []pathPrefix
	comms      map[string]struct{}
	uids       map[uint32]struct{}
	success    *bool
}

type pathPrefix struct {
	value     string
	recursive bool
}

func (e *Evaluator) Decide(value event.Event) Decision {
	class := Classify(value)
	action := e.mutationDefault
	if class == ClassAccess {
		action = e.accessDefault
	}

	decision := Decision{Action: action, Class: class, Explanation: "default policy"}
	decision.Protected = e.matchesProtectedPath(value)
	if decision.Protected {
		decision.Action = ActionAlert
		decision.MatchedRules = append(decision.MatchedRules, "builtin/protected-path")
		decision.Explanation = "protected path activity"
	}

	for _, candidate := range e.rules {
		if candidate.match.matches(value) {
			decision.MatchedRules = append(decision.MatchedRules, candidate.id)
			if candidate.action > decision.Action {
				decision.Action = candidate.action
				decision.Explanation = candidate.reason
			}
		}
	}

	if decision.Protected {
		return decision
	}
	for _, candidate := range e.exceptions {
		if !candidate.expires.After(e.now()) {
			continue
		}
		if decision.Action == ActionAlert && candidate.match.matches(value) {
			decision.Action = ActionAudit
			decision.Suppressed = true
			decision.MatchedRules = append(decision.MatchedRules, candidate.id)
			decision.Explanation = candidate.reason
		}
	}

	return decision
}

func (e *Evaluator) Mode() string { return e.mode }

func (e *Evaluator) matchesProtectedPath(value event.Event) bool {
	for _, candidate := range e.protectedPaths {
		if candidate.matches(value.Path) || candidate.matches(value.DestinationPath) {
			return true
		}
	}
	return false
}

func compileMatch(value MatchConfig) (matcher, error) {
	result := matcher{
		operations: make(map[event.Operation]struct{}),
		comms:      make(map[string]struct{}),
		uids:       make(map[uint32]struct{}),
		success:    value.Success,
	}

	for _, name := range value.Operations {
		operation, err := parseOperation(name)
		if err != nil {
			return matcher{}, err
		}
		result.operations[operation] = struct{}{}
	}
	for _, value := range value.PathPrefixes {
		prefix, err := compilePathPrefix(value)
		if err != nil {
			return matcher{}, err
		}
		result.paths = append(result.paths, prefix)
	}
	for _, value := range value.Comms {
		if strings.TrimSpace(value) == "" {
			return matcher{}, fmt.Errorf("comm cannot be empty")
		}
		result.comms[value] = struct{}{}
	}
	for _, value := range value.UIDs {
		result.uids[value] = struct{}{}
	}

	if len(result.operations) == 0 && len(result.paths) == 0 && len(result.comms) == 0 &&
		len(result.uids) == 0 && result.success == nil {
		return matcher{}, fmt.Errorf("match must contain at least one predicate")
	}
	return result, nil
}

func (m matcher) matches(value event.Event) bool {
	if len(m.operations) > 0 {
		if _, ok := m.operations[value.Operation]; !ok {
			return false
		}
	}
	if len(m.paths) > 0 {
		matched := false
		for _, candidate := range m.paths {
			if candidate.matches(value.Path) || candidate.matches(value.DestinationPath) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(m.comms) > 0 {
		if _, ok := m.comms[value.Comm]; !ok {
			return false
		}
	}
	if len(m.uids) > 0 {
		if _, ok := m.uids[value.UID]; !ok {
			return false
		}
	}
	if m.success != nil && value.Successful() != *m.success {
		return false
	}
	return true
}

func compilePathPrefix(value string) (pathPrefix, error) {
	if value == "" || !strings.HasPrefix(value, "/") {
		return pathPrefix{}, fmt.Errorf("path prefix %q must be absolute", value)
	}
	if strings.ContainsAny(value, "?[") || (strings.Contains(value, "*") && !strings.HasSuffix(value, "/**")) {
		return pathPrefix{}, fmt.Errorf("path prefix %q supports only a trailing /** wildcard", value)
	}

	recursive := strings.HasSuffix(value, "/**")
	value = strings.TrimSuffix(value, "/**")
	value = filepath.Clean(value)
	return pathPrefix{value: value, recursive: recursive}, nil
}

func (p pathPrefix) matches(value string) bool {
	if value == "" || !strings.HasPrefix(value, "/") {
		return false
	}
	value = filepath.Clean(value)
	if value == p.value {
		return true
	}
	return p.recursive && strings.HasPrefix(value, p.value+string(filepath.Separator))
}

func parseOperation(value string) (event.Operation, error) {
	switch strings.ToLower(value) {
	case "open":
		return event.OperationOpen, nil
	case "create":
		return event.OperationCreate, nil
	case "delete":
		return event.OperationDelete, nil
	case "rename":
		return event.OperationRename, nil
	case "chmod":
		return event.OperationChmod, nil
	default:
		return 0, fmt.Errorf("unsupported operation %q", value)
	}
}
