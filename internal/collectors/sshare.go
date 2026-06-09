package collectors

import (
	"log"
	"strconv"
	"strings"

	"github.com/chadsgit/prometheus-slurm-exporter/internal/slurm"
	"github.com/prometheus/client_golang/prometheus"
)

type fairShareMetrics struct {
	fairshare float64
}

// parseFairShareMetrics parses output of: sshare -n -P -o account,fairshare
// Lines indented with spaces are per-user sub-entries and are skipped.
func parseFairShareMetrics(input []byte) map[string]*fairShareMetrics {
	accounts := make(map[string]*fairShareMetrics)
	for _, line := range strings.Split(string(input), "\n") {
		if strings.HasPrefix(line, " ") {
			continue
		}
		if !strings.Contains(line, "|") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		account := strings.TrimSpace(parts[0])
		if account == "" {
			continue
		}
		fairshare, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		accounts[account] = &fairShareMetrics{fairshare: fairshare}
	}
	return accounts
}

// NewFairShareCollector creates a Prometheus collector for per-account fair-share values.
func NewFairShareCollector() *FairShareCollector {
	return &FairShareCollector{
		fairshare: prometheus.NewDesc("slurm_account_fairshare", "Fair-share value for account", []string{"account"}, nil),
	}
}

// FairShareCollector implements prometheus.Collector for sshare fair-share metrics.
type FairShareCollector struct {
	fairshare *prometheus.Desc
}

func (fsc *FairShareCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fsc.fairshare
}

func (fsc *FairShareCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := slurm.Run("sshare", []string{"-n", "-P", "-o", "account,fairshare"})
	if err != nil {
		log.Printf("warn: sshare failed: %v", err)
		return
	}
	for acct, m := range parseFairShareMetrics(data) {
		ch <- prometheus.MustNewConstMetric(fsc.fairshare, prometheus.GaugeValue, m.fairshare, acct)
	}
}
