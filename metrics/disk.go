package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"

	log "github.com/sirupsen/logrus"
)

// The Disk metric gather disk usage statistics from the host machine
type Disk struct{}

// Name of the disk metric
func (d Disk) Name() string {
	return "disk"
}

// Gather disk usage statistics and return the following data points
// - DiskUtilization (percent)
// - DiskUsed (bytes)
// - DiskFree (bytes)
func (d Disk) Gather() (Data, error) {
	log.Debug("gathering disk info")
	diskMetrics, err := disk.Usage("/")
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather disk data")
	}

	diskUtilization := NewDataPoint("DiskUtilization", diskMetrics.UsedPercent, UnitPercent)
	diskUsed := NewDataPoint("DiskUsed", float64(diskMetrics.Used), UnitBytes)
	diskFree := NewDataPoint("DiskFree", float64(diskMetrics.Free), UnitBytes)
	return Data([]*Point{&diskUtilization, &diskUsed, &diskFree}), nil
}
