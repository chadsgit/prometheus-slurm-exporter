package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNodeMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/sinfo_mem.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	metrics := parseNodeMetrics(data)
	assert.Contains(t, metrics, "b001")
	assert.Equal(t, uint64(327680), metrics["b001"].memAlloc)
	assert.Equal(t, uint64(386000), metrics["b001"].memTotal)
	assert.Equal(t, uint64(32), metrics["b001"].cpuAlloc)
	assert.Equal(t, uint64(0), metrics["b001"].cpuIdle)
	assert.Equal(t, uint64(0), metrics["b001"].cpuOther)
	assert.Equal(t, uint64(32), metrics["b001"].cpuTotal)
}

func TestParseNodeMetricsShortLine(t *testing.T) {
	// Bug #2 regression: lines with fewer than 5 fields must not panic
	assert.NotPanics(t, func() {
		parseNodeMetrics([]byte("shortline\n"))
	})
	assert.NotPanics(t, func() {
		parseNodeMetrics([]byte("node01 100 200 1/0\n"))
	})
	assert.NotPanics(t, func() {
		parseNodeMetrics([]byte("node01 100 200 1/0/0\n"))
	})
}
