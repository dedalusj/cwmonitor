package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPU_Name(t *testing.T) {
	cpu := CPU{}
	assert.Equal(t, "cpu", cpu.Name())
}

func TestCPU_Gather(t *testing.T) {
	cpu := CPU{}
	data, err := cpu.Gather()

	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, data[0].Name, "CPUUtilization")
	assert.Equal(t, string(data[0].Unit), string(UnitPercent))
}
