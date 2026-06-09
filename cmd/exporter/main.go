package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/collectors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddr := flag.String("listen-address", ":8080", "Address to listen on for HTTP requests")
	gpuAcct := flag.Bool("gpus-acct", false, "Enable GPU accounting metrics (requires sacct access)")
	fairshareAcct := flag.Bool("fairshare-acct", false, "Enable fair-share metrics (requires sshare)")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewAccountsCollector(),
		collectors.NewCPUsCollector(),
		collectors.NewNodesCollector(),
		collectors.NewNodeCollector(),
		collectors.NewPartitionsCollector(),
		collectors.NewQueueCollector(),
		collectors.NewSchedulerCollector(),
		collectors.NewUsersCollector(),
	)
	if *gpuAcct {
		reg.MustRegister(collectors.NewGPUCollector())
	}
	if *fairshareAcct {
		reg.MustRegister(collectors.NewFairShareCollector())
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("starting prometheus-slurm-exporter on %s (gpus-acct=%v fairshare-acct=%v)",
		*listenAddr, *gpuAcct, *fairshareAcct)
	if err := http.ListenAndServe(*listenAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
