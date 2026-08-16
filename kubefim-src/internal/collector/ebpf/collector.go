package ebpf

import (
	"errors"
	"fmt"
	"os"
	"sync"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"

	generated "kubefim/bpf"
	"kubefim/internal/collector"
	"kubefim/internal/event"
)

// Logger is the logging surface needed by the collector.
type Logger interface {
	Printf(format string, args ...any)
}

// Collector loads KubeFIM's eBPF objects and reads their perf events.
type Collector struct {
	objects generated.BpfObjects
	reader  *perf.Reader
	links   []link.Link
	logger  Logger

	closeOnce sync.Once
	closeErr  error
}

// perfBufferPages is deliberately larger than the one-page prototype buffer.
// Filesystem tracepoints, especially openat, can burst heavily on a busy node.
const perfBufferPages = 64

// New loads the eBPF collection and attaches every tracepoint supported by the
// current kernel.
func New(logger Logger) (*Collector, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("remove memlock limit: %w", err)
	}

	c := &Collector{logger: logger}
	if err := generated.LoadBpfObjects(&c.objects, nil); err != nil {
		return nil, fmt.Errorf("load eBPF objects: %w", err)
	}

	reader, err := perf.NewReader(c.objects.Events, perfBufferPages*os.Getpagesize())
	if err != nil {
		_ = c.objects.Close()
		return nil, fmt.Errorf("create perf event reader: %w", err)
	}
	c.reader = reader

	programs := []struct {
		entryName    string
		entryProgram *cebpf.Program
		exitName     string
		exitProgram  *cebpf.Program
	}{
		{"sys_enter_openat", c.objects.TpEnterOpenat, "sys_exit_openat", c.objects.TpExitOpenat},
		{"sys_enter_unlinkat", c.objects.TpEnterUnlinkat, "sys_exit_unlinkat", c.objects.TpExitUnlinkat},
		{"sys_enter_renameat2", c.objects.TpEnterRenameat2, "sys_exit_renameat2", c.objects.TpExitRenameat2},
		{"sys_enter_chmod", c.objects.TpEnterChmod, "sys_exit_chmod", c.objects.TpExitChmod},
		{"sys_enter_fchmodat", c.objects.TpEnterFchmodat, "sys_exit_fchmodat", c.objects.TpExitFchmodat},
		{"sys_enter_fchmodat2", c.objects.TpEnterFchmodat2, "sys_exit_fchmodat2", c.objects.TpExitFchmodat2},
	}

	for _, program := range programs {
		if err := c.attachPair(program.entryName, program.entryProgram, program.exitName, program.exitProgram); err != nil {
			_ = c.Close()
			return nil, err
		}
	}
	if len(c.links) == 0 {
		_ = c.Close()
		return nil, errors.New("no supported filesystem tracepoints are available on this kernel")
	}

	return c, nil
}

func (c *Collector) attachPair(entryName string, entryProgram *cebpf.Program, exitName string, exitProgram *cebpf.Program) error {
	entry, unavailable, err := c.attach(entryName, entryProgram)
	if err != nil {
		return err
	}
	if unavailable {
		return nil
	}

	exit, unavailable, err := c.attach(exitName, exitProgram)
	if err != nil {
		_ = entry.Close()
		return err
	}
	if unavailable {
		_ = entry.Close()
		c.logger.Printf("Tracepoint pair %s/%s is incomplete; skipping", entryName, exitName)
		return nil
	}

	c.links = append(c.links, entry, exit)
	return nil
}

func (c *Collector) attach(name string, program *cebpf.Program) (link.Link, bool, error) {
	tracepoint, err := link.Tracepoint("syscalls", name, program, nil)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Printf("Tracepoint %s is unavailable on this kernel; skipping", name)
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("attach tracepoint %s: %w", name, err)
	}
	return tracepoint, false, nil
}

// Read waits for and decodes the next perf-buffer record.
func (c *Collector) Read() (collector.Record, error) {
	record, err := c.reader.Read()
	if err != nil {
		if errors.Is(err, perf.ErrClosed) {
			return collector.Record{}, collector.ErrClosed
		}
		return collector.Record{}, err
	}
	if record.LostSamples > 0 {
		return collector.Record{LostSamples: record.LostSamples}, nil
	}

	decoded, err := event.Decode(record.RawSample)
	if err != nil {
		return collector.Record{}, err
	}
	return collector.Record{Event: decoded}, nil
}

// Close releases the reader, tracepoint links, maps, and programs.
func (c *Collector) Close() error {
	c.closeOnce.Do(func() {
		var closeErrors []error
		if c.reader != nil {
			closeErrors = append(closeErrors, c.reader.Close())
		}
		for _, tracepoint := range c.links {
			closeErrors = append(closeErrors, tracepoint.Close())
		}
		closeErrors = append(closeErrors, c.objects.Close())
		c.closeErr = errors.Join(closeErrors...)
	})
	return c.closeErr
}
