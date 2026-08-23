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
	"kubefim/internal/enrichment"
	"kubefim/internal/output"
	"kubefim/internal/policy"
)

func main() {
	configPath := flag.String("config", "", "path to a KubeFIM host policy YAML file")
	outputFormat := flag.String("output", "text", "event output format: text or json")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := log.Default()
	eventOutput, err := output.New(*outputFormat, os.Stdout)
	if err != nil {
		logger.Fatal(err)
	}
	policyEvaluator := policy.Default()
	var eventEnricher enrichment.Enricher = enrichment.None{}
	if node := os.Getenv("NODE_NAME"); node != "" {
		kubernetesEnricher, err := enrichment.NewInCluster(node, "/host/proc", logger)
		if err != nil {
			logger.Fatalf("initialize Kubernetes enrichment: %v", err)
		}
		eventEnricher = kubernetesEnricher
		go kubernetesEnricher.Run(ctx)
		logger.Printf("Kubernetes enrichment enabled for node %s", node)
	}
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

	kubefim := agent.New(eventCollector, eventOutput, policyEvaluator, logger, eventEnricher)
	if err := kubefim.Run(ctx); err != nil {
		logger.Fatalf("run KubeFIM: %v", err)
	}
}
