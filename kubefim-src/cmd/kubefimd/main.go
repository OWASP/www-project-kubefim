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
	"kubefim/internal/metrics"
	"kubefim/internal/output"
	"kubefim/internal/policy"
)

func main() {
	configPath := flag.String("config", "", "path to a KubeFIM host policy YAML file")
	outputFormat := flag.String("output", "text", "event output format: text or json")
	metricsAddress := flag.String("metrics-address", "127.0.0.1:2112", "Prometheus and health listen address; empty disables HTTP")
	flag.Parse()

	signalContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(signalContext)
	defer cancel()

	logger := log.Default()
	metricRegistry := metrics.NewRegistry()
	if *metricsAddress != "" {
		metricsServer, err := metrics.Listen(*metricsAddress, metricRegistry.Handler())
		if err != nil {
			logger.Fatalf("listen for metrics: %v", err)
		}
		logger.Printf("Prometheus metrics listening on %s", metricsServer.Address())
		go func() {
			if err := metricsServer.Serve(); err != nil {
				logger.Printf("metrics server: %v", err)
				cancel()
			}
		}()
		defer func() {
			shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := metricsServer.Shutdown(shutdownContext); err != nil {
				logger.Printf("shut down metrics server: %v", err)
			}
		}()
	}
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

	kubefim := agent.NewWithOptions(eventCollector, eventOutput, policyEvaluator, logger, agent.Options{
		Enricher: eventEnricher,
		Observer: metricRegistry,
	})
	if err := kubefim.Run(ctx); err != nil {
		logger.Fatalf("run KubeFIM: %v", err)
	}
}
