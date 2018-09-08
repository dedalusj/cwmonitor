package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwap_Name(t *testing.T) {
	s := Swap{}
	assert.Equal(t, "swap", s.Name())
}

func TestSwap_Gather(t *testing.T) {
	s := Swap{}
	data, err := s.Gather()

	assert.NoError(t, err)
	assert.Len(t, data, 3)

	assert.Equal(t, data[0].Name, "SwapUtilization")
	assert.Equal(t, string(data[0].Unit), string(UnitPercent))

	assert.Equal(t, data[1].Name, "SwapUsed")
	assert.Equal(t, string(data[1].Unit), string(UnitBytes))

	assert.Equal(t, data[2].Name, "SwapFree")
	assert.Equal(t, string(data[2].Unit), string(UnitBytes))
}
