package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAccountsMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/squeue_accounts.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	am := parseAccountsMetrics(data)

	assert.Contains(t, am, "physics")
	assert.Equal(t, float64(2), am["physics"].running)
	assert.Equal(t, float64(48), am["physics"].runningCPUs)
	assert.Equal(t, float64(0), am["physics"].pending)
	assert.Equal(t, float64(1), am["physics"].suspended)

	assert.Contains(t, am, "chemistry")
	assert.Equal(t, float64(1), am["chemistry"].running)
	assert.Equal(t, float64(1), am["chemistry"].pending)

	assert.Contains(t, am, "biology")
	assert.Equal(t, float64(1), am["biology"].running)
	assert.Equal(t, float64(1), am["biology"].pending)
}

func TestParseAccountsMetricsMalformedLine(t *testing.T) {
	assert.NotPanics(t, func() {
		parseAccountsMetrics([]byte("nopipes\n"))
	})
	assert.NotPanics(t, func() {
		parseAccountsMetrics([]byte("a|b\n"))
	})
}
