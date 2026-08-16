package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kubefim/internal/agent"
	collectorebpf "kubefim/internal/collector/ebpf"
	"kubefim/internal/output"
	"kubefim/internal/policy"
)

func main() {
	configPath := flag.String("config", "", "path to a KubeFIM host policy YAML file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := log.Default()
	policyEvaluator := policy.Default()
	if *configPath != "" {
		loadedPolicy, err := policy.LoadFile(*configPath, time.Now())
		if err != nil {
			logger.Fatalf("load policy: %v", err)
		}
		policyEvaluator = loadedPolicy
		logger.Printf("Loaded policy %s in %s mode", *configPath, policyEvaluator.Mode())
	}

	logger.Println("Attaching BPF tracepoints...")

	eventCollector, err := collectorebpf.New(logger)
	if err != nil {
		logger.Fatalf("initialize eBPF collector: %v", err)
	}

	kubefim := agent.New(eventCollector, output.NewText(os.Stdout), policyEvaluator, logger)
	if err := kubefim.Run(ctx); err != nil {
		logger.Fatalf("run KubeFIM: %v", err)
	}
}
