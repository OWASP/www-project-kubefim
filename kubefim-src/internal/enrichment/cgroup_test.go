package enrichment

import (
	"strings"
	"testing"
)

func TestParseCgroup(t *testing.T) {
	id := strings.Repeat("a", 64)
	tests := []struct {
		name, data, runtime string
		host, found         bool
	}{
		{name: "cgroup v2 containerd", data: "0::/kubepods.slice/kubepods-burstable.slice/cri-containerd-" + id + ".scope\n", runtime: "containerd", found: true},
		{name: "cgroup v1 crio", data: "11:memory:/kubepods/besteffort/crio-" + id + "\n", runtime: "cri-o", found: true},
		{name: "docker", data: "0::/docker/" + id + "\n", runtime: "docker", found: true},
		{name: "host root", data: "0::/\n", host: true, found: true},
		{name: "host system service", data: "0::/system.slice/sshd.service\n", host: true, found: true},
		{name: "unrecognized non-host cgroup", data: "0::/kubepods.slice/unknown.scope\n", found: true},
		{name: "malformed", data: "invalid\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseCgroup(strings.NewReader(test.data))
			if ok != test.found || got.Host != test.host || got.Runtime != test.runtime {
				t.Fatalf("parseCgroup() = %+v, %t", got, ok)
			}
			if test.runtime != "" && got.ID != id {
				t.Fatalf("container ID = %q", got.ID)
			}
		})
	}
}

func TestParseCgroupRejectsShortContainerID(t *testing.T) {
	got, ok := parseCgroup(strings.NewReader("0::/docker/abc123\n"))
	if !ok || got.ID != "" || got.Host {
		t.Fatalf("short ID was attributed: %+v, %t", got, ok)
	}
}
