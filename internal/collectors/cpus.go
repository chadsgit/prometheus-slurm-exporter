package collectors

import (
	"log"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

type cpuMetrics struct {
	alloc float64
	idle  float64
	other float64
	total float64
}

func parseCPUsMetrics(input []byte) *cpuMetrics {
	var cm cpuMetrics
	s := strings.TrimSpace(string(input))
	if !strings.Contains(s, "/") {
		return &cm
	}
	parts := strings.Split(s, "/")
	if len(parts) < 4 {
		return &cm
	}
	cm.alloc, _ = strconv.ParseFloat(parts[0], 64)
	cm.idle, _ = strconv.ParseFloat(parts[1], 64)
	cm.other, _ = strconv.ParseFloat(parts[2], 64)
	cm.total, _ = strconv.ParseFloat(parts[3], 64)
	return &cm
}

// NewCPUsCollector creates a Prometheus collector for cluster-wide CPU state.
func NewCPUsCollector() *CPUsCollector {
	return &CPUsCollector{
		alloc: prometheus.NewDesc("slurm_cpus_alloc", "Allocated CPUs", nil, nil),
		idle:  prometheus.NewDesc("slurm_cpus_idle", "Idle CPUs", nil, nil),
		other: prometheus.NewDesc("slurm_cpus_other", "Other CPUs", nil, nil),
		total: prometheus.NewDesc("slurm_cpus_total", "Total CPUs", nil, nil),
	}
}

// CPUsCollector implements prometheus.Collector for sinfo CPU state.
type CPUsCollector struct {
	alloc *prometheus.Desc
	idle  *prometheus.Desc
	other *prometheus.Desc
	total *prometheus.Desc
}

func (cc *CPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.other
	ch <- cc.total
}

func (cc *CPUsCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("sinfo", []string{"-h", "-o", "%C"})
	if err != nil {
		log.Printf("warn: sinfo -o %%C failed: %v", err)
		return
	}
	cm := parseCPUsMetrics(data)
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, cm.alloc)
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, cm.idle)
	ch <- prometheus.MustNewConstMetric(cc.other, prometheus.GaugeValue, cm.other)
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, cm.total)
}
