package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSchedulerMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/sdiag.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	sm := parseSchedulerMetrics(data)
	assert.Equal(t, float64(3), sm.threads)
	assert.Equal(t, float64(0), sm.dbdQueueSize)
	assert.Equal(t, float64(97209), sm.lastCycle)
	assert.Equal(t, float64(74593), sm.meanCycle)
	assert.Equal(t, float64(63), sm.cyclePerMinute)
	assert.Equal(t, float64(1942890), sm.backfillLastCycle)
	assert.Equal(t, float64(1960820), sm.backfillMeanCycle)
	assert.Equal(t, float64(29324), sm.backfillDepthMean)
	assert.Equal(t, float64(111544), sm.totalBackfilledJobsSinceStart)
	assert.Equal(t, float64(793), sm.totalBackfilledJobsSinceCycle)
	assert.Equal(t, float64(10), sm.totalBackfilledHeterogeneous)
}

func TestParseSchedulerMetricsBug7(t *testing.T) {
	// Bug #7 regression: when backfilling stats are absent, backfillLastCycle
	// and backfillMeanCycle must stay 0, not mirror the main scheduler values.
	input := []byte(`Server thread count:  2
Agent queue size:     0
DBD Agent queue size: 0
Main schedule statistics (microseconds):
        Last cycle:   50000
        Mean cycle:   40000
        Cycles per minute: 30
`)
	sm := parseSchedulerMetrics(input)
	assert.Equal(t, float64(50000), sm.lastCycle)
	assert.Equal(t, float64(0), sm.backfillLastCycle, "backfillLastCycle must be 0 when backfill section is absent")
	assert.Equal(t, float64(40000), sm.meanCycle)
	assert.Equal(t, float64(0), sm.backfillMeanCycle, "backfillMeanCycle must be 0 when backfill section is absent")
}
