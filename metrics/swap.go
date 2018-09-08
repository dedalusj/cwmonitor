package metrics

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"
)

type Swap struct{}

func (s Swap) Name() string {
	return "swap"
}

func (s Swap) Gather() (Data, error) {
	swapMetrics, err := mem.SwapMemory()
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to gather memory data")
	}

	swapUtilization := NewDataPoint("SwapUtilization", swapMetrics.UsedPercent, UnitPercent)
	swapUsed := NewDataPoint("SwapUsed", float64(swapMetrics.Used), UnitBytes)
	swapFree := NewDataPoint("SwapFree", float64(swapMetrics.Free), UnitBytes)
	return Data([]*Point{&swapUtilization, &swapUsed, &swapFree}), nil
}
