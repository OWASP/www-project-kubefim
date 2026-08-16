package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kubefim/internal/agent"
	collectorebpf "kubefim/internal/collector/ebpf"
	"kubefim/internal/output"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := log.Default()
	logger.Println("Attaching BPF tracepoints...")

	eventCollector, err := collectorebpf.New(logger)
	if err != nil {
		logger.Fatalf("initialize eBPF collector: %v", err)
	}

	kubefim := agent.New(eventCollector, output.NewText(os.Stdout), logger)
	if err := kubefim.Run(ctx); err != nil {
		logger.Fatalf("run KubeFIM: %v", err)
	}
}
