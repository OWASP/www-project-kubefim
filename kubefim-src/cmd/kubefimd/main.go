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

// The Go struct must exactly match the C struct
type Event struct {
	PID  uint32
	UID  uint32
	Comm [16]byte
	Path [256]byte
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

	// Attach the tracepoint
	tp, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TracepointOpenat, nil)
	if err != nil {
		log.Fatalf("attaching tracepoint: %s", err)
	}
	defer tp.Close()

	log.Println("Waiting for events (minimal baseline mode)...")

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
		fmt.Printf("PID: %-7d UID: %-5d Comm: %-15s Path: %s\n", event.PID, event.UID, comm, path)
	}
}