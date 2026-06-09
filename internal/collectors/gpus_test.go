package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTotalGPUs(t *testing.T) {
	data, err := os.ReadFile("testdata/sinfo_gpus.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	total := parseTotalGPUs(data)
	assert.Equal(t, float64(12), total, "expected 4+8 = 12 (null line skipped)")
}

func TestParseTotalGPUsBoundsCheck(t *testing.T) {
	// Bug #3 regression: lines with only one field must not panic
	assert.NotPanics(t, func() {
		parseTotalGPUs([]byte("node01\n"))
	})
	assert.NotPanics(t, func() {
		parseTotalGPUs([]byte("\n"))
	})
}

func TestParseAllocatedGPUsOldFormat(t *testing.T) {
	data, err := os.ReadFile("testdata/sacct_alloctres_old.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	alloc := parseAllocatedGPUs(data)
	assert.Equal(t, float64(6), alloc, "expected 2+4 = 6 (gpu:N format)")
}

func TestParseAllocatedGPUsNewFormat(t *testing.T) {
	// Bug #6 regression: Slurm 23+ uses gres/gpu=N instead of gpu:N
	data, err := os.ReadFile("testdata/sacct_alloctres_new.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	alloc := parseAllocatedGPUs(data)
	assert.Equal(t, float64(6), alloc, "expected 4+2 = 6 (gres/gpu=N format)")
}

func TestGPUUtilizationNaNGuard(t *testing.T) {
	// Bug #4 regression: utilization must not be NaN when total is 0
	utilization := gpuUtilization(0, 0)
	assert.Equal(t, float64(0), utilization, "utilization with zero total must be 0 not NaN")
	utilization = gpuUtilization(6, 12)
	assert.Equal(t, float64(0.5), utilization)
}
