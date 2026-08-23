package enrichment

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"kubefim/internal/event"
)

var containerIDPattern = regexp.MustCompile(`(?:^|[-:/])([a-f0-9]{64})(?:\.scope)?(?:$|/)`)

type procCgroupReader interface {
	Open(pid uint32) (io.ReadCloser, error)
}

type hostProc struct{ root string }

func (p hostProc) Open(pid uint32) (io.ReadCloser, error) {
	return os.Open(filepath.Join(p.root, fmt.Sprint(pid), "cgroup"))
}

type cgroupResolver struct{ proc procCgroupReader }

func newCgroupResolver(root string) cgroupResolver {
	return cgroupResolver{proc: hostProc{root: root}}
}

func (r cgroupResolver) Resolve(pid uint32) (event.Container, bool) {
	file, err := r.proc.Open(pid)
	if err != nil {
		return event.Container{}, false
	}
	defer file.Close()
	return parseCgroup(file)
}

func parseCgroup(reader io.Reader) (event.Container, bool) {
	scanner := bufio.NewScanner(reader)
	host := true
	parsed := false
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 3)
		if len(parts) != 3 {
			continue
		}
		parsed = true
		path := parts[2]
		if path != "/" && path != "" && !strings.HasPrefix(path, "/init.scope") &&
			!strings.HasPrefix(path, "/system.slice/") && !strings.HasPrefix(path, "/user.slice/") {
			host = false
		}
		match := containerIDPattern.FindStringSubmatch(path)
		if len(match) != 2 {
			continue
		}
		return event.Container{ID: match[1], Runtime: runtimeFromCgroup(path)}, true
	}
	if err := scanner.Err(); err != nil {
		return event.Container{}, false
	}
	if !parsed {
		return event.Container{}, false
	}
	return event.Container{Host: host}, true
}

func runtimeFromCgroup(path string) string {
	switch {
	case strings.Contains(path, "cri-containerd") || strings.Contains(path, "containerd"):
		return "containerd"
	case strings.Contains(path, "crio"):
		return "cri-o"
	case strings.Contains(path, "docker"):
		return "docker"
	default:
		return "unknown"
	}
}
