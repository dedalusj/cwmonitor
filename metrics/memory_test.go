package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory_Name(t *testing.T) {
	m := Memory{}
	assert.Equal(t, "memory", m.Name())
}

func TestMemory_Gather(t *testing.T) {
	m := Memory{}
	data, err := m.Gather()
	assert.NoError(t, err)
	assert.Len(t, data, 3)

	assert.Equal(t, data[0].Name, "MemoryUtilization")
	assert.Equal(t, string(data[0].Unit), string(UnitPercent))

	assert.Equal(t, data[1].Name, "MemoryUsed")
	assert.Equal(t, string(data[1].Unit), string(UnitBytes))

	assert.Equal(t, data[2].Name, "MemoryAvailable")
	assert.Equal(t, string(data[2].Unit), string(UnitBytes))
}
