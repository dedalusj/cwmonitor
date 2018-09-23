package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"

	log "github.com/sirupsen/logrus"
)

// The CPU metric gather CPU usage statistics from the host machine
type CPU struct{}

// Name for the CPU metric
func (c CPU) Name() string {
	return "cpu"
}

// Gather CPU usage statistics and return the CPUUtilization data point as percentage
func (c CPU) Gather() (Data, error) {
	log.Debug("gathering CPU info")
	cpuMetrics, err := cpu.Percent(0, false)
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather cpu data")
	}

	cpuUtilization := NewDataPoint("CPUUtilization", cpuMetrics[0], UnitPercent)
	return Data([]*Point{&cpuUtilization}), nil
}
