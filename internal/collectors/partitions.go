package collectors

import (
	"log"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

type partitionMetrics struct {
	allocated float64
	idle      float64
	other     float64
	pending   float64
	total     float64
}

// parsePartitionCPUs parses output of: sinfo -h -o %R,%C
func parsePartitionCPUs(input []byte) map[string]*partitionMetrics {
	partitions := make(map[string]*partitionMetrics)
	for _, line := range strings.Split(string(input), "\n") {
		if !strings.Contains(line, ",") {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		name := parts[0]
		cpuParts := strings.Split(parts[1], "/")
		if len(cpuParts) < 4 {
			continue
		}
		allocated, _ := strconv.ParseFloat(cpuParts[0], 64)
		idle, _ := strconv.ParseFloat(cpuParts[1], 64)
		other, _ := strconv.ParseFloat(cpuParts[2], 64)
		total, _ := strconv.ParseFloat(cpuParts[3], 64)
		partitions[name] = &partitionMetrics{
			allocated: allocated,
			idle:      idle,
			other:     other,
			total:     total,
		}
	}
	return partitions
}

// parsePartitionPending adds pending job counts from: squeue -a -r -h -o %P --states=PENDING
func parsePartitionPending(input []byte, partitions map[string]*partitionMetrics) {
	for _, line := range strings.Split(string(input), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if _, ok := partitions[name]; ok {
			partitions[name].pending++
		}
	}
}

// NewPartitionsCollector creates a Prometheus collector for per-partition CPU and job metrics.
func NewPartitionsCollector() *PartitionsCollector {
	labels := []string{"partition"}
	return &PartitionsCollector{
		allocated: prometheus.NewDesc("slurm_partition_cpus_allocated", "Allocated CPUs for partition", labels, nil),
		idle:      prometheus.NewDesc("slurm_partition_cpus_idle", "Idle CPUs for partition", labels, nil),
		other:     prometheus.NewDesc("slurm_partition_cpus_other", "Other CPUs for partition", labels, nil),
		pending:   prometheus.NewDesc("slurm_partition_jobs_pending", "Pending jobs for partition", labels, nil),
		total:     prometheus.NewDesc("slurm_partition_cpus_total", "Total CPUs for partition", labels, nil),
	}
}

// PartitionsCollector implements prometheus.Collector for per-partition metrics.
type PartitionsCollector struct {
	allocated *prometheus.Desc
	idle      *prometheus.Desc
	other     *prometheus.Desc
	pending   *prometheus.Desc
	total     *prometheus.Desc
}

func (pc *PartitionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pc.allocated
	ch <- pc.idle
	ch <- pc.other
	ch <- pc.pending
	ch <- pc.total
}

func (pc *PartitionsCollector) Collect(ch chan<- prometheus.Metric) {
	sinfoData, err := slurm.Run("sinfo", []string{"-h", "-o", "%R,%C"})
	if err != nil {
		log.Printf("warn: sinfo partitions failed: %v", err)
		return
	}
	pm := parsePartitionCPUs(sinfoData)

	squeueData, err := slurm.Run("squeue", []string{"-a", "-r", "-h", "-o", "%P", "--states=PENDING"})
	if err != nil {
		log.Printf("warn: squeue pending partitions failed: %v", err)
		// continue with CPU data even if pending query fails
	} else {
		parsePartitionPending(squeueData, pm)
	}

	for name, m := range pm {
		if m.allocated > 0 {
			ch <- prometheus.MustNewConstMetric(pc.allocated, prometheus.GaugeValue, m.allocated, name)
		}
		if m.idle > 0 {
			ch <- prometheus.MustNewConstMetric(pc.idle, prometheus.GaugeValue, m.idle, name)
		}
		if m.other > 0 {
			ch <- prometheus.MustNewConstMetric(pc.other, prometheus.GaugeValue, m.other, name)
		}
		if m.pending > 0 {
			ch <- prometheus.MustNewConstMetric(pc.pending, prometheus.GaugeValue, m.pending, name)
		}
		if m.total > 0 {
			ch <- prometheus.MustNewConstMetric(pc.total, prometheus.GaugeValue, m.total, name)
		}
	}
}
