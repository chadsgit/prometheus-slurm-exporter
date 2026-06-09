package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueueMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/squeue.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	qm := parseQueueMetrics(data)
	assert.Equal(t, float64(28), qm.running)
	assert.Equal(t, float64(4), qm.pending)
	assert.Equal(t, float64(1), qm.configuring)
	assert.Equal(t, float64(1), qm.preempted)
	assert.Equal(t, float64(1), qm.nodeFail)
	assert.Equal(t, float64(1), qm.completed)
	assert.Equal(t, float64(1), qm.failed)
	assert.Equal(t, float64(1), qm.timeout)
	assert.Equal(t, float64(1), qm.suspended)
	assert.Equal(t, float64(1), qm.cancelled)
	assert.Equal(t, float64(2), qm.completing)
}

func TestParseQueueMetricsEmpty(t *testing.T) {
	qm := parseQueueMetrics([]byte(""))
	assert.Equal(t, float64(0), qm.running)
	assert.Equal(t, float64(0), qm.pending)
}

func TestParseQueueMetricsPendingDependency(t *testing.T) {
	data := []byte("123,PENDING,Dependency\n456,PENDING,Resources\n")
	qm := parseQueueMetrics(data)
	assert.Equal(t, float64(2), qm.pending)
	assert.Equal(t, float64(1), qm.pendingDep)
}
