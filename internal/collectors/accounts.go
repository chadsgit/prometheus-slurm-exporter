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
	reAcctPending   = regexp.MustCompile(`^pending`)
	reAcctRunning   = regexp.MustCompile(`^running`)
	reAcctSuspended = regexp.MustCompile(`^suspended`)
)

type jobMetrics struct {
	pending     float64
	running     float64
	runningCPUs float64
	suspended   float64
}

// parseAccountsMetrics parses output of: squeue -a -r -h -o %A|%a|%T|%C
func parseAccountsMetrics(input []byte) map[string]*jobMetrics {
	accounts := make(map[string]*jobMetrics)
	for _, line := range strings.Split(string(input), "\n") {
		if !strings.Contains(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		account := parts[1]
		if _, ok := accounts[account]; !ok {
			accounts[account] = &jobMetrics{}
		}
		state := strings.ToLower(parts[2])
		cpus, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		switch {
		case reAcctPending.MatchString(state):
			accounts[account].pending++
		case reAcctRunning.MatchString(state):
			accounts[account].running++
			accounts[account].runningCPUs += cpus
		case reAcctSuspended.MatchString(state):
			accounts[account].suspended++
		}
	}
	return accounts
}

// NewAccountsCollector creates a Prometheus collector for per-account job metrics.
func NewAccountsCollector() *AccountsCollector {
	labels := []string{"account"}
	return &AccountsCollector{
		pending:     prometheus.NewDesc("slurm_account_jobs_pending", "Pending jobs for account", labels, nil),
		running:     prometheus.NewDesc("slurm_account_jobs_running", "Running jobs for account", labels, nil),
		runningCPUs: prometheus.NewDesc("slurm_account_cpus_running", "Running CPUs for account", labels, nil),
		suspended:   prometheus.NewDesc("slurm_account_jobs_suspended", "Suspended jobs for account", labels, nil),
	}
}

// AccountsCollector implements prometheus.Collector for per-account job counts.
type AccountsCollector struct {
	pending     *prometheus.Desc
	running     *prometheus.Desc
	runningCPUs *prometheus.Desc
	suspended   *prometheus.Desc
}

func (ac *AccountsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- ac.pending
	ch <- ac.running
	ch <- ac.runningCPUs
	ch <- ac.suspended
}

func (ac *AccountsCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("squeue", []string{"-a", "-r", "-h", "-o", "%A|%a|%T|%C"})
	if err != nil {
		log.Printf("warn: squeue accounts failed: %v", err)
		return
	}
	for acct, m := range parseAccountsMetrics(data) {
		if m.pending > 0 {
			ch <- prometheus.MustNewConstMetric(ac.pending, prometheus.GaugeValue, m.pending, acct)
		}
		if m.running > 0 {
			ch <- prometheus.MustNewConstMetric(ac.running, prometheus.GaugeValue, m.running, acct)
		}
		if m.runningCPUs > 0 {
			ch <- prometheus.MustNewConstMetric(ac.runningCPUs, prometheus.GaugeValue, m.runningCPUs, acct)
		}
		if m.suspended > 0 {
			ch <- prometheus.MustNewConstMetric(ac.suspended, prometheus.GaugeValue, m.suspended, acct)
		}
	}
}
