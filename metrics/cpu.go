package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"

	log "github.com/sirupsen/logrus"
)

type CPU struct{}

func (c CPU) Name() string {
	return "cpu"
}

func (c CPU) Gather() (Data, error) {
	log.Debug("gathering CPU info")
	cpuMetrics, err := cpu.Percent(0, false)
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather cpu data")
	}

	cpuUtilization := NewDataPoint("CPUUtilization", cpuMetrics[0], UnitPercent)
	return Data([]*Point{&cpuUtilization}), nil
}

