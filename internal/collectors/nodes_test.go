package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNodesMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/sinfo_nodes.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	nm := parseNodesMetrics(data)
	assert.Greater(t, nm.alloc, float64(0), "expected allocated nodes > 0")
	assert.Greater(t, nm.idle, float64(0), "expected idle nodes > 0")
	assert.Greater(t, nm.down, float64(0), "expected down nodes > 0")
	assert.Greater(t, nm.drain, float64(0), "expected drain nodes > 0")
	assert.Equal(t, float64(1), nm.fail, "expected exactly 1 fail node")
}

func TestParseNodesMetricsEmpty(t *testing.T) {
	nm := parseNodesMetrics([]byte(""))
	assert.Equal(t, float64(0), nm.alloc)
	assert.Equal(t, float64(0), nm.idle)
}
