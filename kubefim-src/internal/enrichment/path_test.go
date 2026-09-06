package enrichment

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"kubefim/internal/event"
)

func TestPathResolver(t *testing.T) {
	procRoot := t.TempDir()
	processRoot := filepath.Join(procRoot, "42")
	if err := os.MkdirAll(filepath.Join(processRoot, "fd"), 0o755); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(procRoot, "workspace", "app")
	fd := filepath.Join(procRoot, "etc", "service")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(fd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cwd, filepath.Join(processRoot, "cwd")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(fd, filepath.Join(processRoot, "fd", "7")); err != nil {
		t.Fatal(err)
	}
	cwdInode, cwdDevice := fileIdentity(t, cwd)
	fdInode, fdDevice := fileIdentity(t, fd)

	resolver := NewPathResolver(procRoot)
	tests := []struct {
		name       string
		path       string
		dirfd      int32
		inode      uint64
		device     uint64
		want       string
		resolution event.PathResolution
	}{
		{name: "absolute", path: "/etc/../etc/passwd", dirfd: 9, want: "/etc/passwd", resolution: event.PathResolutionAbsolute},
		{name: "cwd", path: "config/app.yaml", dirfd: event.ATFDCWD, inode: cwdInode, device: cwdDevice, want: filepath.Join(cwd, "config/app.yaml"), resolution: event.PathResolutionResolved},
		{name: "directory fd", path: "../token", dirfd: 7, inode: fdInode, device: fdDevice, want: filepath.Join(procRoot, "etc", "token"), resolution: event.PathResolutionResolved},
		{name: "missing fd", path: "token", dirfd: 8, resolution: event.PathResolutionFDUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := resolver.Enrich(context.Background(), event.Event{PID: 42, Path: test.path, DirFD: test.dirfd, BaseInode: test.inode, BaseDevice: test.device})
			if got.ResolvedPath != test.want || got.PathResolution != test.resolution {
				t.Fatalf("resolved path = %q (%s), want %q (%s)", got.ResolvedPath, got.PathResolution, test.want, test.resolution)
			}
			if got.Path != test.path {
				t.Fatalf("requested path changed to %q", got.Path)
			}
		})
	}
}

func TestPathResolverRename(t *testing.T) {
	procRoot := t.TempDir()
	processRoot := filepath.Join(procRoot, "71")
	if err := os.MkdirAll(filepath.Join(processRoot, "fd"), 0o755); err != nil {
		t.Fatal(err)
	}
	for fd, target := range map[string]string{"3": "/old", "4": "/new"} {
		target = filepath.Join(procRoot, target)
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(processRoot, "fd", fd)); err != nil {
			t.Fatal(err)
		}
	}
	oldInode, oldDevice := fileIdentity(t, filepath.Join(procRoot, "old"))
	newInode, newDevice := fileIdentity(t, filepath.Join(procRoot, "new"))

	got := NewPathResolver(procRoot).Enrich(context.Background(), event.Event{
		PID: 71, Path: "source", DestinationPath: "target", DirFD: 3, DestinationDirFD: 4,
		BaseInode: oldInode, BaseDevice: oldDevice, DestinationBaseInode: newInode, DestinationBaseDevice: newDevice,
	})
	if got.ResolvedPath != filepath.Join(procRoot, "old", "source") || got.ResolvedDestinationPath != filepath.Join(procRoot, "new", "target") {
		t.Fatalf("rename paths are %+v", got)
	}
	if got.PathResolution != event.PathResolutionResolved || got.DestinationPathResolution != event.PathResolutionResolved {
		t.Fatalf("rename resolution is %s/%s", got.PathResolution, got.DestinationPathResolution)
	}
}

func TestPathResolverRejectsChangedBase(t *testing.T) {
	procRoot := t.TempDir()
	processRoot := filepath.Join(procRoot, "8")
	base := filepath.Join(procRoot, "base")
	if err := os.MkdirAll(processRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(base, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(base, filepath.Join(processRoot, "cwd")); err != nil {
		t.Fatal(err)
	}

	got := NewPathResolver(procRoot).Enrich(context.Background(), event.Event{
		PID: 8, Path: "file", DirFD: event.ATFDCWD, BaseInode: 1, BaseDevice: 1,
	})
	if got.ResolvedPath != "" || got.PathResolution != event.PathResolutionBaseChanged {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestPathResolverProcessExited(t *testing.T) {
	got := NewPathResolver(t.TempDir()).Enrich(context.Background(), event.Event{
		PID: 999, Path: "relative", DirFD: event.ATFDCWD,
	})
	if got.ResolvedPath != "" || got.PathResolution != event.PathResolutionProcessExited {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestPathResolverRejectsDeletedBase(t *testing.T) {
	procRoot := t.TempDir()
	processRoot := filepath.Join(procRoot, "5")
	if err := os.MkdirAll(processRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/tmp/work (deleted)", filepath.Join(processRoot, "cwd")); err != nil {
		t.Fatal(err)
	}

	got := NewPathResolver(procRoot).Enrich(context.Background(), event.Event{
		PID: 5, Path: "file", DirFD: event.ATFDCWD,
	})
	if got.PathResolution != event.PathResolutionUnresolved {
		t.Fatalf("deleted base reported as %s", got.PathResolution)
	}
}

func TestEnricherChain(t *testing.T) {
	first := enricherFunc(func(value event.Event) event.Event { value.Path = "first"; return value })
	second := enricherFunc(func(value event.Event) event.Event { value.Path += "+second"; return value })
	got := (Chain{first, second}).Enrich(context.Background(), event.Event{})
	if got.Path != "first+second" {
		t.Fatalf("path = %q", got.Path)
	}
}

type enricherFunc func(event.Event) event.Event

func (f enricherFunc) Enrich(_ context.Context, value event.Event) event.Event { return f(value) }

func fileIdentity(t *testing.T, path string) (uint64, uint64) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat := info.Sys().(*syscall.Stat_t)
	return stat.Ino, uint64(stat.Dev)
}
