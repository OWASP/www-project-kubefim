package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
	"kubefim/bpf"
)

type Event struct {
    PID       uint32
    UID       uint32
    Comm      [16]byte
    Path      [256]byte
    EventType uint32
}

func (e *Event) EventName() string {
	switch e.EventType {
	case 1:
		return "OPEN"
	case 2:
		return "CREATE"
	case 3:
		return "DELETE"
	case 4:
		return "RENAME"
	case 5:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

func main() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("Removing memlock: %v", err)
	}

	objs := bpf.BpfObjects{}
	if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	// Open a Perf Event reader from the BPF map.
	rd, err := perf.NewReader(objs.Events, os.Getpagesize())
	if err != nil {
		log.Fatalf("creating perf event reader: %s", err)
	}
	defer rd.Close()

	links := []link.Link{}

	attach := func(tp link.Link, err error) {
		if err != nil {
			log.Fatalf("Attach error: %v", err)
		}
		links = append(links, tp)
	}

	log.Println("Attaching BPF tracepoints...")

	attach(link.Tracepoint("syscalls", "sys_enter_openat", objs.TpOpenat, nil))
	attach(link.Tracepoint("syscalls", "sys_enter_unlinkat", objs.TpUnlink, nil))
	attach(link.Tracepoint("syscalls", "sys_enter_renameat2", objs.TpRename, nil))
	attach(link.Tracepoint("syscalls", "sys_enter_fchmod", objs.TpChmodFchmod, nil))
    attach(link.Tracepoint("syscalls", "sys_enter_fchmodat", objs.TpChmodFchmodat, nil))
    attach(link.Tracepoint("syscalls", "sys_enter_fchmodat2", objs.TpChmodFchmodat2, nil))

	defer func() {
		for _, l := range links {
			l.Close()
		}
	}()

	log.Println("Listening for events...")

	// Handle shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Exiting...")
		rd.Close()
	}()

	var event Event
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				log.Println("Reader closed.")
				return
			}
			log.Printf("reading from perf reader: %s", err)
			continue
		}

		// The perf event array can lose samples if the userspace program
		// does not keep up.
		if record.LostSamples > 0 {
			log.Printf("lost %d samples", record.LostSamples)
			continue
		}

		// Decode the raw data into our Event struct
		if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parsing perf event: %s", err)
			continue
		}

		// Print the raw, unfiltered event data
		comm := strings.Trim(string(event.Comm[:]), "\x00")
		path := strings.Trim(string(event.Path[:]), "\x00")
		fmt.Printf("[%s] PID=%d UID=%d COMM=%s PATH=%s\n", event.EventName(), event.PID, event.UID, comm, path,)
	}
}