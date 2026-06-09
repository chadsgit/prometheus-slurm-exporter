package collectors

import (
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reNodeAlloc = regexp.MustCompile(`^alloc`)
	reNodeComp  = regexp.MustCompile(`^comp`)
	reNodeDown  = regexp.MustCompile(`^down`)
	reNodeDrain = regexp.MustCompile(`^drain`)
	reNodeFail  = regexp.MustCompile(`^fail`)
	reNodeErr   = regexp.MustCompile(`^err`)
	reNodeIdle  = regexp.MustCompile(`^idle`)
	reNodeMaint = regexp.MustCompile(`^maint`)
	reNodeMix   = regexp.MustCompile(`^mix`)
	reNodeResv  = regexp.MustCompile(`^res`)
)

type nodesMetrics struct {
	alloc float64
	comp  float64
	down  float64
	drain float64
	err   float64
	fail  float64
	idle  float64
	maint float64
	mix   float64
	resv  float64
}

func removeDuplicates(s []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range s {
		if len(v) > 0 && !seen[v] {
			out = append(out, v)
			seen[v] = true
		}
	}
	return out
}

func parseNodesMetrics(input []byte) *nodesMetrics {
	var nm nodesMetrics
	lines := strings.Split(string(input), "\n")
	sort.Strings(lines)
	for _, line := range removeDuplicates(lines) {
		if !strings.Contains(line, ",") {
			continue
		}
		parts := strings.Split(line, ",")
		count, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		state := parts[1]
		switch {
		case reNodeAlloc.MatchString(state):
			nm.alloc += count
		case reNodeComp.MatchString(state):
			nm.comp += count
		case reNodeDown.MatchString(state):
			nm.down += count
		case reNodeDrain.MatchString(state):
			nm.drain += count
		case reNodeFail.MatchString(state):
			nm.fail += count
		case reNodeErr.MatchString(state):
			nm.err += count
		case reNodeIdle.MatchString(state):
			nm.idle += count
		case reNodeMaint.MatchString(state):
			nm.maint += count
		case reNodeMix.MatchString(state):
			nm.mix += count
		case reNodeResv.MatchString(state):
			nm.resv += count
		}
	}
	return &nm
}

// NewNodesCollector creates a Prometheus collector for node state counts.
func NewNodesCollector() *NodesCollector {
	return &NodesCollector{
		alloc: prometheus.NewDesc("slurm_nodes_alloc", "Allocated nodes", nil, nil),
		comp:  prometheus.NewDesc("slurm_nodes_comp", "Completing nodes", nil, nil),
		down:  prometheus.NewDesc("slurm_nodes_down", "Down nodes", nil, nil),
		drain: prometheus.NewDesc("slurm_nodes_drain", "Drain nodes", nil, nil),
		err:   prometheus.NewDesc("slurm_nodes_err", "Error nodes", nil, nil),
		fail:  prometheus.NewDesc("slurm_nodes_fail", "Fail nodes", nil, nil),
		idle:  prometheus.NewDesc("slurm_nodes_idle", "Idle nodes", nil, nil),
		maint: prometheus.NewDesc("slurm_nodes_maint", "Maint nodes", nil, nil),
		mix:   prometheus.NewDesc("slurm_nodes_mix", "Mixed nodes", nil, nil),
		resv:  prometheus.NewDesc("slurm_nodes_resv", "Reserved nodes", nil, nil),
	}
}

// NodesCollector implements prometheus.Collector for sinfo node state counts.
type NodesCollector struct {
	alloc *prometheus.Desc
	comp  *prometheus.Desc
	down  *prometheus.Desc
	drain *prometheus.Desc
	err   *prometheus.Desc
	fail  *prometheus.Desc
	idle  *prometheus.Desc
	maint *prometheus.Desc
	mix   *prometheus.Desc
	resv  *prometheus.Desc
}

func (nc *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nc.alloc
	ch <- nc.comp
	ch <- nc.down
	ch <- nc.drain
	ch <- nc.err
	ch <- nc.fail
	ch <- nc.idle
	ch <- nc.maint
	ch <- nc.mix
	ch <- nc.resv
}

func (nc *NodesCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("sinfo", []string{"-h", "-o", "%D,%T"})
	if err != nil {
		log.Printf("warn: sinfo -o %%D,%%T failed: %v", err)
		return
	}
	nm := parseNodesMetrics(data)
	ch <- prometheus.MustNewConstMetric(nc.alloc, prometheus.GaugeValue, nm.alloc)
	ch <- prometheus.MustNewConstMetric(nc.comp, prometheus.GaugeValue, nm.comp)
	ch <- prometheus.MustNewConstMetric(nc.down, prometheus.GaugeValue, nm.down)
	ch <- prometheus.MustNewConstMetric(nc.drain, prometheus.GaugeValue, nm.drain)
	ch <- prometheus.MustNewConstMetric(nc.err, prometheus.GaugeValue, nm.err)
	ch <- prometheus.MustNewConstMetric(nc.fail, prometheus.GaugeValue, nm.fail)
	ch <- prometheus.MustNewConstMetric(nc.idle, prometheus.GaugeValue, nm.idle)
	ch <- prometheus.MustNewConstMetric(nc.maint, prometheus.GaugeValue, nm.maint)
	ch <- prometheus.MustNewConstMetric(nc.mix, prometheus.GaugeValue, nm.mix)
	ch <- prometheus.MustNewConstMetric(nc.resv, prometheus.GaugeValue, nm.resv)
}
