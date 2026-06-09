package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePartitionCPUs(t *testing.T) {
	data, err := os.ReadFile("testdata/sinfo_partitions.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	pm := parsePartitionCPUs(data)

	assert.Contains(t, pm, "gpu")
	assert.Equal(t, float64(100), pm["gpu"].allocated)
	assert.Equal(t, float64(50), pm["gpu"].idle)
	assert.Equal(t, float64(10), pm["gpu"].other)
	assert.Equal(t, float64(160), pm["gpu"].total)

	assert.Contains(t, pm, "cpu")
	assert.Equal(t, float64(500), pm["cpu"].allocated)
	assert.Equal(t, float64(720), pm["cpu"].total)

	assert.Contains(t, pm, "debug")
	assert.Equal(t, float64(0), pm["debug"].other)
}

func TestParsePartitionPending(t *testing.T) {
	data, err := os.ReadFile("testdata/squeue_pending_partitions.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	pm := map[string]*partitionMetrics{
		"gpu":   {},
		"cpu":   {},
		"debug": {},
	}
	parsePartitionPending(data, pm)
	assert.Equal(t, float64(2), pm["gpu"].pending)
	assert.Equal(t, float64(2), pm["cpu"].pending)
	assert.Equal(t, float64(1), pm["debug"].pending)
}

func TestParsePartitionCPUsMalformedLine(t *testing.T) {
	assert.NotPanics(t, func() {
		parsePartitionCPUs([]byte("gpu,100/50\n"))
	})
	assert.NotPanics(t, func() {
		parsePartitionCPUs([]byte("gpu\n"))
	})
}
