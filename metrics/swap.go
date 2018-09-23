package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"

	log "github.com/sirupsen/logrus"
)

// The Swap metric gather swap usage statistics from the host machine
type Swap struct{}

// Name of the swap metric
func (s Swap) Name() string {
	return "swap"
}

// Gather swap usage statistics from the host machine and return data points
// for SwapUtilization (percent), SwapUsed (bytes) and SwapFree (bytes)
func (s Swap) Gather() (Data, error) {
	log.Debug("gathering swap info")
	swapMetrics, err := mem.SwapMemory()
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather memory data")
	}

	swapUtilization := NewDataPoint("SwapUtilization", swapMetrics.UsedPercent, UnitPercent)
	swapUsed := NewDataPoint("SwapUsed", float64(swapMetrics.Used), UnitBytes)
	swapFree := NewDataPoint("SwapFree", float64(swapMetrics.Free), UnitBytes)
	return Data([]*Point{&swapUtilization, &swapUsed, &swapFree}), nil
}
