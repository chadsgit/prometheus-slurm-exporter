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
	reUserPending   = regexp.MustCompile(`^pending`)
	reUserRunning   = regexp.MustCompile(`^running`)
	reUserSuspended = regexp.MustCompile(`^suspended`)
)

type userJobMetrics struct {
	pending     float64
	running     float64
	runningCPUs float64
	suspended   float64
}

// parseUsersMetrics parses output of: squeue -a -r -h -o %A|%u|%T|%C
func parseUsersMetrics(input []byte) map[string]*userJobMetrics {
	users := make(map[string]*userJobMetrics)
	for _, line := range strings.Split(string(input), "\n") {
		if !strings.Contains(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		user := parts[1]
		if _, ok := users[user]; !ok {
			users[user] = &userJobMetrics{}
		}
		state := strings.ToLower(parts[2])
		cpus, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		switch {
		case reUserPending.MatchString(state):
			users[user].pending++
		case reUserRunning.MatchString(state):
			users[user].running++
			users[user].runningCPUs += cpus
		case reUserSuspended.MatchString(state):
			users[user].suspended++
		}
	}
	return users
}

// NewUsersCollector creates a Prometheus collector for per-user job metrics.
func NewUsersCollector() *UsersCollector {
	labels := []string{"user"}
	return &UsersCollector{
		pending:     prometheus.NewDesc("slurm_user_jobs_pending", "Pending jobs for user", labels, nil),
		running:     prometheus.NewDesc("slurm_user_jobs_running", "Running jobs for user", labels, nil),
		runningCPUs: prometheus.NewDesc("slurm_user_cpus_running", "Running CPUs for user", labels, nil),
		suspended:   prometheus.NewDesc("slurm_user_jobs_suspended", "Suspended jobs for user", labels, nil),
	}
}

// UsersCollector implements prometheus.Collector for per-user job counts.
type UsersCollector struct {
	pending     *prometheus.Desc
	running     *prometheus.Desc
	runningCPUs *prometheus.Desc
	suspended   *prometheus.Desc
}

func (uc *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- uc.pending
	ch <- uc.running
	ch <- uc.runningCPUs
	ch <- uc.suspended
}

func (uc *UsersCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("squeue", []string{"-a", "-r", "-h", "-o", "%A|%u|%T|%C"})
	if err != nil {
		log.Printf("warn: squeue users failed: %v", err)
		return
	}
	for user, m := range parseUsersMetrics(data) {
		if m.pending > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending, prometheus.GaugeValue, m.pending, user)
		}
		if m.running > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running, prometheus.GaugeValue, m.running, user)
		}
		if m.runningCPUs > 0 {
			ch <- prometheus.MustNewConstMetric(uc.runningCPUs, prometheus.GaugeValue, m.runningCPUs, user)
		}
		if m.suspended > 0 {
			ch <- prometheus.MustNewConstMetric(uc.suspended, prometheus.GaugeValue, m.suspended, user)
		}
	}
}
