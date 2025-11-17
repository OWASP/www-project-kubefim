//go:build tools

package tools

import (
	_ "github.com/cilium/ebpf/cmd/bpf2go"
	- "github.com/cilium/ebpf/link"
	- "github.com/cilium/ebpf/perf"
	- "github.com/cilium/ebpf/rlimit"
)