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
	reSchedThreads  = regexp.MustCompile(`^Server thread`)
	reSchedQueue    = regexp.MustCompile(`^Agent queue`)
	reSchedDBD      = regexp.MustCompile(`^DBD Agent`)
	reSchedLastCyc  = regexp.MustCompile(`^\s+Last cycle$`)
	reSchedMeanCyc  = regexp.MustCompile(`^\s+Mean cycle$`)
	reSchedCPM      = regexp.MustCompile(`^\s+Cycles per`)
	reSchedDepthMean = regexp.MustCompile(`^\s+Depth Mean$`)
	reSchedTBS      = regexp.MustCompile(`^\s+Total backfilled jobs \(since last slurm start\)`)
	reSchedTBC      = regexp.MustCompile(`^\s+Total backfilled jobs \(since last stats cycle start\)`)
	reSchedTBH      = regexp.MustCompile(`^\s+Total backfilled heterogeneous job components`)
)

type schedulerMetrics struct {
	threads                       float64
	queueSize                     float64
	dbdQueueSize                  float64
	lastCycle                     float64
	meanCycle                     float64
	cyclePerMinute                float64
	backfillLastCycle             float64
	backfillMeanCycle             float64
	backfillDepthMean             float64
	totalBackfilledJobsSinceStart float64
	totalBackfilledJobsSinceCycle float64
	totalBackfilledHeterogeneous  float64
}

func parseSchedulerMetrics(input []byte) *schedulerMetrics {
	var sm schedulerMetrics
	lcCount := 0
	mcCount := 0
	for _, line := range strings.Split(string(input), "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := parts[0]
		val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		switch {
		case reSchedThreads.MatchString(key):
			sm.threads = val
		case reSchedQueue.MatchString(key):
			sm.queueSize = val
		case reSchedDBD.MatchString(key):
			sm.dbdQueueSize = val
		case reSchedLastCyc.MatchString(key):
			// Bug #7 fix: if/else if so the counter guard is mutually exclusive per iteration.
			// Without else, setting lcCount=1 in the first block immediately triggers the
			// second block on the same line, assigning backfillLastCycle the main value.
			if lcCount == 0 {
				sm.lastCycle = val
				lcCount = 1
			} else if lcCount == 1 {
				sm.backfillLastCycle = val
			}
		case reSchedMeanCyc.MatchString(key):
			if mcCount == 0 {
				sm.meanCycle = val
				mcCount = 1
			} else if mcCount == 1 {
				sm.backfillMeanCycle = val
			}
		case reSchedCPM.MatchString(key):
			sm.cyclePerMinute = val
		case reSchedDepthMean.MatchString(key):
			sm.backfillDepthMean = val
		case reSchedTBS.MatchString(key):
			sm.totalBackfilledJobsSinceStart = val
		case reSchedTBC.MatchString(key):
			sm.totalBackfilledJobsSinceCycle = val
		case reSchedTBH.MatchString(key):
			sm.totalBackfilledHeterogeneous = val
		}
	}
	return &sm
}

// NewSchedulerCollector creates a Prometheus collector for Slurm scheduler stats.
func NewSchedulerCollector() *SchedulerCollector {
	return &SchedulerCollector{
		threads:                       prometheus.NewDesc("slurm_scheduler_threads", "Scheduler thread count", nil, nil),
		queueSize:                     prometheus.NewDesc("slurm_scheduler_queue_size", "Scheduler queue length", nil, nil),
		dbdQueueSize:                  prometheus.NewDesc("slurm_scheduler_dbd_queue_size", "DBD agent queue length", nil, nil),
		lastCycle:                     prometheus.NewDesc("slurm_scheduler_last_cycle", "Scheduler last cycle time (µs)", nil, nil),
		meanCycle:                     prometheus.NewDesc("slurm_scheduler_mean_cycle", "Scheduler mean cycle time (µs)", nil, nil),
		cyclePerMinute:                prometheus.NewDesc("slurm_scheduler_cycle_per_minute", "Scheduler cycles per minute", nil, nil),
		backfillLastCycle:             prometheus.NewDesc("slurm_scheduler_backfill_last_cycle", "Backfill scheduler last cycle time (µs)", nil, nil),
		backfillMeanCycle:             prometheus.NewDesc("slurm_scheduler_backfill_mean_cycle", "Backfill scheduler mean cycle time (µs)", nil, nil),
		backfillDepthMean:             prometheus.NewDesc("slurm_scheduler_backfill_depth_mean", "Backfill scheduler mean depth", nil, nil),
		totalBackfilledJobsSinceStart: prometheus.NewDesc("slurm_scheduler_backfilled_jobs_since_start_total", "Backfilled jobs since last Slurm start", nil, nil),
		totalBackfilledJobsSinceCycle: prometheus.NewDesc("slurm_scheduler_backfilled_jobs_since_cycle_total", "Backfilled jobs since last stats cycle", nil, nil),
		totalBackfilledHeterogeneous:  prometheus.NewDesc("slurm_scheduler_backfilled_heterogeneous_total", "Backfilled heterogeneous job components", nil, nil),
	}
}

// SchedulerCollector implements prometheus.Collector for sdiag scheduler metrics.
type SchedulerCollector struct {
	threads                       *prometheus.Desc
	queueSize                     *prometheus.Desc
	dbdQueueSize                  *prometheus.Desc
	lastCycle                     *prometheus.Desc
	meanCycle                     *prometheus.Desc
	cyclePerMinute                *prometheus.Desc
	backfillLastCycle             *prometheus.Desc
	backfillMeanCycle             *prometheus.Desc
	backfillDepthMean             *prometheus.Desc
	totalBackfilledJobsSinceStart *prometheus.Desc
	totalBackfilledJobsSinceCycle *prometheus.Desc
	totalBackfilledHeterogeneous  *prometheus.Desc
}

func (sc *SchedulerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.threads
	ch <- sc.queueSize
	ch <- sc.dbdQueueSize
	ch <- sc.lastCycle
	ch <- sc.meanCycle
	ch <- sc.cyclePerMinute
	ch <- sc.backfillLastCycle
	ch <- sc.backfillMeanCycle
	ch <- sc.backfillDepthMean
	ch <- sc.totalBackfilledJobsSinceStart
	ch <- sc.totalBackfilledJobsSinceCycle
	ch <- sc.totalBackfilledHeterogeneous
}

func (sc *SchedulerCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("sdiag", nil)
	if err != nil {
		log.Printf("warn: sdiag failed: %v", err)
		return
	}
	sm := parseSchedulerMetrics(data)
	ch <- prometheus.MustNewConstMetric(sc.threads, prometheus.GaugeValue, sm.threads)
	ch <- prometheus.MustNewConstMetric(sc.queueSize, prometheus.GaugeValue, sm.queueSize)
	ch <- prometheus.MustNewConstMetric(sc.dbdQueueSize, prometheus.GaugeValue, sm.dbdQueueSize)
	ch <- prometheus.MustNewConstMetric(sc.lastCycle, prometheus.GaugeValue, sm.lastCycle)
	ch <- prometheus.MustNewConstMetric(sc.meanCycle, prometheus.GaugeValue, sm.meanCycle)
	ch <- prometheus.MustNewConstMetric(sc.cyclePerMinute, prometheus.GaugeValue, sm.cyclePerMinute)
	ch <- prometheus.MustNewConstMetric(sc.backfillLastCycle, prometheus.GaugeValue, sm.backfillLastCycle)
	ch <- prometheus.MustNewConstMetric(sc.backfillMeanCycle, prometheus.GaugeValue, sm.backfillMeanCycle)
	ch <- prometheus.MustNewConstMetric(sc.backfillDepthMean, prometheus.GaugeValue, sm.backfillDepthMean)
	ch <- prometheus.MustNewConstMetric(sc.totalBackfilledJobsSinceStart, prometheus.GaugeValue, sm.totalBackfilledJobsSinceStart)
	ch <- prometheus.MustNewConstMetric(sc.totalBackfilledJobsSinceCycle, prometheus.GaugeValue, sm.totalBackfilledJobsSinceCycle)
	ch <- prometheus.MustNewConstMetric(sc.totalBackfilledHeterogeneous, prometheus.GaugeValue, sm.totalBackfilledHeterogeneous)
}
