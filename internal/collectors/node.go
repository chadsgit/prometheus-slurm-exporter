package collectors

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

type nodeMetrics struct {
	memAlloc   uint64
	memTotal   uint64
	cpuAlloc   uint64
	cpuIdle    uint64
	cpuOther   uint64
	cpuTotal   uint64
	nodeStatus string
}

func parseNodeMetrics(input []byte) map[string]*nodeMetrics {
	nodes := make(map[string]*nodeMetrics)
	lines := strings.Split(string(input), "\n")
	sort.Strings(lines)
	for _, line := range removeDuplicates(lines) {
		fields := strings.Fields(line)
		if len(fields) < 5 { // Bug #2 fix: bounds check before indexing
			continue
		}
		cpuParts := strings.Split(fields[3], "/")
		if len(cpuParts) < 4 { // Bug #2 fix: bounds check on CPU A/I/O/T string
			continue
		}
		name := fields[0]
		memAlloc, _ := strconv.ParseUint(fields[1], 10, 64)
		memTotal, _ := strconv.ParseUint(fields[2], 10, 64)
		cpuAlloc, _ := strconv.ParseUint(cpuParts[0], 10, 64)
		cpuIdle, _ := strconv.ParseUint(cpuParts[1], 10, 64)
		cpuOther, _ := strconv.ParseUint(cpuParts[2], 10, 64)
		cpuTotal, _ := strconv.ParseUint(cpuParts[3], 10, 64)
		nodes[name] = &nodeMetrics{
			memAlloc:   memAlloc,
			memTotal:   memTotal,
			cpuAlloc:   cpuAlloc,
			cpuIdle:    cpuIdle,
			cpuOther:   cpuOther,
			cpuTotal:   cpuTotal,
			nodeStatus: fields[4],
		}
	}
	return nodes
}

// NewNodeCollector creates a Prometheus collector for per-node CPU and memory metrics.
func NewNodeCollector() *NodeCollector {
	labels := []string{"node", "status"}
	return &NodeCollector{
		cpuAlloc: prometheus.NewDesc("slurm_node_cpu_alloc", "Allocated CPUs per node", labels, nil),
		cpuIdle:  prometheus.NewDesc("slurm_node_cpu_idle", "Idle CPUs per node", labels, nil),
		cpuOther: prometheus.NewDesc("slurm_node_cpu_other", "Other CPUs per node", labels, nil),
		cpuTotal: prometheus.NewDesc("slurm_node_cpu_total", "Total CPUs per node", labels, nil),
		memAlloc: prometheus.NewDesc("slurm_node_mem_alloc", "Allocated memory per node", labels, nil),
		memTotal: prometheus.NewDesc("slurm_node_mem_total", "Total memory per node", labels, nil),
	}
}

// NodeCollector implements prometheus.Collector for per-node CPU/memory metrics.
type NodeCollector struct {
	cpuAlloc *prometheus.Desc
	cpuIdle  *prometheus.Desc
	cpuOther *prometheus.Desc
	cpuTotal *prometheus.Desc
	memAlloc *prometheus.Desc
	memTotal *prometheus.Desc
}

func (nc *NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nc.cpuAlloc
	ch <- nc.cpuIdle
	ch <- nc.cpuOther
	ch <- nc.cpuTotal
	ch <- nc.memAlloc
	ch <- nc.memTotal
}

func (nc *NodeCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("sinfo", []string{"-h", "-N", "-O", "NodeList,AllocMem,Memory,CPUsState,StateLong"})
	if err != nil {
		log.Printf("warn: sinfo per-node failed: %v", err)
		return
	}
	for name, m := range parseNodeMetrics(data) {
		ch <- prometheus.MustNewConstMetric(nc.cpuAlloc, prometheus.GaugeValue, float64(m.cpuAlloc), name, m.nodeStatus)
		ch <- prometheus.MustNewConstMetric(nc.cpuIdle, prometheus.GaugeValue, float64(m.cpuIdle), name, m.nodeStatus)
		ch <- prometheus.MustNewConstMetric(nc.cpuOther, prometheus.GaugeValue, float64(m.cpuOther), name, m.nodeStatus)
		ch <- prometheus.MustNewConstMetric(nc.cpuTotal, prometheus.GaugeValue, float64(m.cpuTotal), name, m.nodeStatus)
		ch <- prometheus.MustNewConstMetric(nc.memAlloc, prometheus.GaugeValue, float64(m.memAlloc), name, m.nodeStatus)
		ch <- prometheus.MustNewConstMetric(nc.memTotal, prometheus.GaugeValue, float64(m.memTotal), name, m.nodeStatus)
	}
}
