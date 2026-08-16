package policy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Spec       SpecConfig `yaml:"spec"`
}

type SpecConfig struct {
	Mode           string            `yaml:"mode"`
	Defaults       DefaultsConfig    `yaml:"defaults"`
	ProtectedPaths []string          `yaml:"protectedPaths"`
	Rules          []RuleConfig      `yaml:"rules"`
	Exceptions     []ExceptionConfig `yaml:"exceptions"`
}

type DefaultsConfig struct {
	Access   string `yaml:"access"`
	Mutation string `yaml:"mutation"`
}

type RuleConfig struct {
	ID     string      `yaml:"id"`
	Match  MatchConfig `yaml:"match"`
	Action string      `yaml:"action"`
	Reason string      `yaml:"reason"`
	Owner  string      `yaml:"owner"`
}

type ExceptionConfig struct {
	ID      string      `yaml:"id"`
	Match   MatchConfig `yaml:"match"`
	Reason  string      `yaml:"reason"`
	Owner   string      `yaml:"owner"`
	Expires string      `yaml:"expires"`
}

type MatchConfig struct {
	Operations   []string `yaml:"operations"`
	PathPrefixes []string `yaml:"pathPrefixes"`
	Comms        []string `yaml:"comms"`
	UIDs         []uint32 `yaml:"uids"`
	Success      *bool    `yaml:"success"`
}

func LoadFile(filename string, now time.Time) (*Evaluator, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open policy file: %w", err)
	}
	defer file.Close()

	return Decode(file, now)
}

func Decode(reader io.Reader, now time.Time) (*Evaluator, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode policy: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("policy file must contain exactly one YAML document")
		}
		return nil, fmt.Errorf("decode trailing policy document: %w", err)
	}

	return compile(config, now)
}

func Default() *Evaluator {
	return &Evaluator{
		mode:            "observe",
		accessDefault:   ActionAggregate,
		mutationDefault: ActionAudit,
		now:             time.Now,
	}
}

func compile(config Config, now time.Time) (*Evaluator, error) {
	if config.APIVersion != APIVersion {
		return nil, fmt.Errorf("apiVersion must be %q", APIVersion)
	}
	if config.Kind != Kind {
		return nil, fmt.Errorf("kind must be %q", Kind)
	}

	mode := strings.ToLower(config.Spec.Mode)
	if mode == "" {
		mode = "observe"
	}
	if mode != "observe" {
		return nil, fmt.Errorf("mode %q is unsupported; only observe is safe in v1alpha1", mode)
	}

	accessDefault, err := defaultAction(config.Spec.Defaults.Access, ActionAggregate)
	if err != nil {
		return nil, fmt.Errorf("defaults.access: %w", err)
	}
	mutationDefault, err := defaultAction(config.Spec.Defaults.Mutation, ActionAudit)
	if err != nil {
		return nil, fmt.Errorf("defaults.mutation: %w", err)
	}

	evaluator := &Evaluator{
		mode:            mode,
		accessDefault:   accessDefault,
		mutationDefault: mutationDefault,
		now:             time.Now,
	}

	for _, value := range config.Spec.ProtectedPaths {
		matcher, err := compilePathPrefix(value)
		if err != nil {
			return nil, fmt.Errorf("protectedPaths: %w", err)
		}
		evaluator.protectedPaths = append(evaluator.protectedPaths, matcher)
	}

	ids := make(map[string]struct{})
	for _, value := range config.Spec.Rules {
		if err := validateMetadata(value.ID, value.Reason, value.Owner, ids); err != nil {
			return nil, fmt.Errorf("rule: %w", err)
		}
		matcher, err := compileMatch(value.Match)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", value.ID, err)
		}
		action, err := parseAction(value.Action)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", value.ID, err)
		}
		evaluator.rules = append(evaluator.rules, rule{
			id: value.ID, match: matcher, action: action, reason: value.Reason,
		})
	}

	for _, value := range config.Spec.Exceptions {
		if err := validateMetadata(value.ID, value.Reason, value.Owner, ids); err != nil {
			return nil, fmt.Errorf("exception: %w", err)
		}
		matcher, err := compileMatch(value.Match)
		if err != nil {
			return nil, fmt.Errorf("exception %q: %w", value.ID, err)
		}
		if len(value.Match.Operations) == 0 || len(value.Match.PathPrefixes) == 0 ||
			len(value.Match.Comms) == 0 || len(value.Match.UIDs) == 0 {
			return nil, fmt.Errorf("exception %q must match operation, path, comm, and UID", value.ID)
		}
		expires, err := time.Parse(time.RFC3339, value.Expires)
		if err != nil {
			return nil, fmt.Errorf("exception %q has invalid RFC3339 expiry: %w", value.ID, err)
		}
		if !expires.After(now) {
			return nil, fmt.Errorf("exception %q expired at %s", value.ID, expires.Format(time.RFC3339))
		}
		evaluator.exceptions = append(evaluator.exceptions, exception{
			id: value.ID, match: matcher, reason: value.Reason, expires: expires,
		})
	}

	return evaluator, nil
}

func defaultAction(value string, fallback Action) (Action, error) {
	if value == "" {
		return fallback, nil
	}
	return parseAction(value)
}

func validateMetadata(id, reason, owner string, ids map[string]struct{}) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(reason) == "" || strings.TrimSpace(owner) == "" {
		return errors.New("id, reason, and owner are required")
	}
	if _, exists := ids[id]; exists {
		return fmt.Errorf("duplicate id %q", id)
	}
	ids[id] = struct{}{}
	return nil
}
