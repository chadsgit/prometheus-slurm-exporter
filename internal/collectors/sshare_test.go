package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFairShareMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/sshare.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	fsm := parseFairShareMetrics(data)

	// top-level accounts only
	assert.Contains(t, fsm, "physics")
	assert.Equal(t, 0.6, fsm["physics"].fairshare)
	assert.Contains(t, fsm, "chemistry")
	assert.Equal(t, 0.4, fsm["chemistry"].fairshare)
	assert.Contains(t, fsm, "biology")
	assert.Equal(t, 0.1, fsm["biology"].fairshare)

	// per-user sub-entries must not be in the map
	assert.NotContains(t, fsm, "  alice")
	assert.NotContains(t, fsm, "alice")
	assert.NotContains(t, fsm, "  carol")
}

func TestParseFairShareMetricsMalformedLine(t *testing.T) {
	assert.NotPanics(t, func() {
		parseFairShareMetrics([]byte("nopipes\n"))
	})
}
