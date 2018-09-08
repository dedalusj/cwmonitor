package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisk_Name(t *testing.T) {
	d := Disk{}
	assert.Equal(t, "disk", d.Name())
}

func TestDisk_Gather(t *testing.T) {
	d := Disk{}
	data, err := d.Gather()
	assert.NoError(t, err)
	assert.Len(t, data, 3)

	assert.Equal(t, data[0].Name, "DiskUtilization")
	assert.Equal(t, string(data[0].Unit), string(UnitPercent))

	assert.Equal(t, data[1].Name, "DiskUsed")
	assert.Equal(t, string(data[1].Unit), string(UnitBytes))

	assert.Equal(t, data[2].Name, "DiskFree")
	assert.Equal(t, string(data[2].Unit), string(UnitBytes))
}
