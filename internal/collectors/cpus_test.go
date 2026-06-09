package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCPUsMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/sinfo_cpus.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	cm := parseCPUsMetrics(data)
	assert.Equal(t, float64(5725), cm.alloc)
	assert.Equal(t, float64(877), cm.idle)
	assert.Equal(t, float64(34), cm.other)
	assert.Equal(t, float64(6636), cm.total)
}

func TestParseCPUsMetricsEmpty(t *testing.T) {
	cm := parseCPUsMetrics([]byte(""))
	assert.Equal(t, float64(0), cm.alloc)
	assert.Equal(t, float64(0), cm.total)
}
