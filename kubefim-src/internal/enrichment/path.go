package enrichment

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"kubefim/internal/event"
)

type PathResolver struct {
	procRoot string
}

func NewPathResolver(procRoot string) PathResolver {
	return PathResolver{procRoot: procRoot}
}

func (r PathResolver) Enrich(_ context.Context, value event.Event) event.Event {
	pid := value.TGID
	if pid == 0 {
		pid = value.PID
	}
	value.ResolvedPath, value.PathResolution = r.resolve(pid, value.DirFD, value.Path, value.BaseInode, value.BaseDevice)
	if value.DestinationPath != "" {
		value.ResolvedDestinationPath, value.DestinationPathResolution = r.resolve(
			pid, value.DestinationDirFD, value.DestinationPath,
			value.DestinationBaseInode, value.DestinationBaseDevice,
		)
	}
	return value
}

func (r PathResolver) resolve(pid uint32, dirfd int32, requested string, expectedInode, expectedDevice uint64) (string, event.PathResolution) {
	if requested == "" {
		return "", event.PathResolutionUnresolved
	}
	if filepath.IsAbs(requested) {
		return filepath.Clean(requested), event.PathResolutionAbsolute
	}

	processRoot := filepath.Join(r.procRoot, strconv.FormatUint(uint64(pid), 10))
	if _, err := os.Stat(processRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", event.PathResolutionProcessExited
		}
		return "", event.PathResolutionUnresolved
	}

	link := filepath.Join(processRoot, "cwd")
	missing := event.PathResolutionProcessExited
	if dirfd != event.ATFDCWD {
		link = filepath.Join(processRoot, "fd", strconv.FormatInt(int64(dirfd), 10))
		missing = event.PathResolutionFDUnavailable
	}
	base, err := os.Readlink(link)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", missing
		}
		return "", event.PathResolutionUnresolved
	}
	if !filepath.IsAbs(base) || strings.HasSuffix(base, " (deleted)") {
		return "", event.PathResolutionUnresolved
	}
	if expectedInode == 0 && expectedDevice == 0 {
		return "", event.PathResolutionUnresolved
	}
	info, err := os.Stat(link)
	if err != nil {
		return "", event.PathResolutionUnresolved
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Ino != expectedInode || uint64(stat.Dev) != expectedDevice {
		return "", event.PathResolutionBaseChanged
	}
	confirmed, err := os.Readlink(link)
	if err != nil || confirmed != base {
		return "", event.PathResolutionBaseChanged
	}
	return filepath.Clean(filepath.Join(base, requested)), event.PathResolutionResolved
}
