package collectors

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Bug #6: match both old gpu:N and new gres/gpu=N (Slurm 23+)
	reGPUOld = regexp.MustCompile(`gpu:(\d+)`)
	reGPUNew = regexp.MustCompile(`gres/gpu=(\d+)`)
)

// parseTotalGPUs parses output of: sinfo -h -o "%n %G"
func parseTotalGPUs(input []byte) float64 {
	var total float64
	for _, line := range strings.Split(string(input), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 { // Bug #3 fix: bounds check before indexing [1]
			continue
		}
		gres := fields[1]
		if gres == "(null)" {
			continue
		}
		// gres field may be "gpu:N" or "gpu:N(S:0)" — extract leading number
		if m := reGPUOld.FindStringSubmatch(gres); len(m) == 2 {
			n, _ := strconv.ParseFloat(m[1], 64)
			total += n
		}
	}
	return total
}

// parseAllocatedGPUs parses sacct AllocTRES output (one TRES string per line).
// Handles both old gpu:N (Slurm <23) and new gres/gpu=N (Slurm 23+) formats.
func parseAllocatedGPUs(input []byte) float64 {
	var total float64
	for _, line := range strings.Split(string(input), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Try new format first (gres/gpu=N), then fall back to old (gpu:N)
		if m := reGPUNew.FindStringSubmatch(line); len(m) == 2 {
			n, _ := strconv.ParseFloat(m[1], 64)
			total += n
		} else if m := reGPUOld.FindStringSubmatch(line); len(m) == 2 {
			n, _ := strconv.ParseFloat(m[1], 64)
			total += n
		}
	}
	return total
}

// gpuUtilization returns allocated/total, guarding against NaN when total is 0.
func gpuUtilization(allocated, total float64) float64 {
	if total == 0 { // Bug #4 fix: avoid NaN from 0/0
		return 0
	}
	return allocated / total
}

// NewGPUCollector creates a Prometheus collector for cluster-wide GPU metrics.
func NewGPUCollector() *GPUCollector {
	return &GPUCollector{
		total:       prometheus.NewDesc("slurm_gpus_total", "Total GPUs in cluster", nil, nil),
		allocated:   prometheus.NewDesc("slurm_gpus_alloc", "Allocated GPUs", nil, nil),
		utilization: prometheus.NewDesc("slurm_gpus_utilization", "GPU utilization ratio (0-1)", nil, nil),
	}
}

// GPUCollector implements prometheus.Collector for GPU metrics.
type GPUCollector struct {
	total       *prometheus.Desc
	allocated   *prometheus.Desc
	utilization *prometheus.Desc
}

func (gc *GPUCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- gc.total
	ch <- gc.allocated
	ch <- gc.utilization
}

func (gc *GPUCollector) Collect(ch chan<- prometheus.Metric) {
	// Bug #5 fix: args as separate slice elements, not a single string with spaces
	sinfoData, err := slurm.Run("sinfo", []string{"-h", "-o", "%n %G"})
	if err != nil {
		log.Printf("warn: sinfo -o %%n %%G failed: %v", err)
		return
	}
	total := parseTotalGPUs(sinfoData)

	sacctData, err := slurm.Run("sacct", []string{"-a", "-X", "--noheader", "-o", "AllocTRES", "--state=R"})
	if err != nil {
		log.Printf("warn: sacct AllocTRES failed: %v", err)
		return
	}
	allocated := parseAllocatedGPUs(sacctData)

	ch <- prometheus.MustNewConstMetric(gc.total, prometheus.GaugeValue, total)
	ch <- prometheus.MustNewConstMetric(gc.allocated, prometheus.GaugeValue, allocated)
	ch <- prometheus.MustNewConstMetric(gc.utilization, prometheus.GaugeValue, gpuUtilization(allocated, total))
}
