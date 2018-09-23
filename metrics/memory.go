package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"

	log "github.com/sirupsen/logrus"
)

// The Memory metric gather memory usage statistics from the host machine
type Memory struct{}

// Name for the Memory metric
func (m Memory) Name() string {
	return "memory"
}

// Gather memory usage statistics from the host machine and return the following data points
// - MemoryUtilization (percent)
// - MemoryUsed (bytes)
// - MemoryAvailable (bytes)
func (m Memory) Gather() (Data, error) {
	log.Debug("gathering memory info")
	memoryMetrics, err := mem.VirtualMemory()
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather memory data")
	}

	memoryUtilization := NewDataPoint("MemoryUtilization", memoryMetrics.UsedPercent, UnitPercent)
	memoryUsed := NewDataPoint("MemoryUsed", float64(memoryMetrics.Used), UnitBytes)
	memoryAvailable := NewDataPoint("MemoryAvailable", float64(memoryMetrics.Available), UnitBytes)
	return Data([]*Point{&memoryUtilization, &memoryUsed, &memoryAvailable}), nil
}
