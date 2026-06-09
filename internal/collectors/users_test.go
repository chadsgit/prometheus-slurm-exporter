package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUsersMetrics(t *testing.T) {
	data, err := os.ReadFile("testdata/squeue_users.txt")
	if err != nil {
		t.Fatalf("cannot open test data: %v", err)
	}
	um := parseUsersMetrics(data)

	assert.Contains(t, um, "alice")
	assert.Equal(t, float64(2), um["alice"].running)
	assert.Equal(t, float64(48), um["alice"].runningCPUs)
	assert.Equal(t, float64(0), um["alice"].pending)
	assert.Equal(t, float64(1), um["alice"].suspended)

	assert.Contains(t, um, "bob")
	assert.Equal(t, float64(1), um["bob"].running)
	assert.Equal(t, float64(1), um["bob"].pending)

	assert.Contains(t, um, "carol")
	assert.Equal(t, float64(1), um["carol"].running)
	assert.Equal(t, float64(1), um["carol"].pending)
}

func TestParseUsersMetricsMalformedLine(t *testing.T) {
	assert.NotPanics(t, func() {
		parseUsersMetrics([]byte("nopipes\n"))
	})
	assert.NotPanics(t, func() {
		parseUsersMetrics([]byte("a|b\n"))
	})
}
