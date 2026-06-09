package slurm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunSuccess(t *testing.T) {
	out, err := Run("echo", []string{"hello"})
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", string(out))
}

func TestRunError(t *testing.T) {
	_, err := Run("false", nil)
	assert.Error(t, err)
}
