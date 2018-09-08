package monitor

import (
	"strconv"
	"testing"

	"cwmonitor/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("validates name", func(t *testing.T) {
		c := Config{}
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("validates interval", func(t *testing.T) {
		c := Config{}
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "interval")
	})

	t.Run("validates id", func(t *testing.T) {
		c := Config{}
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id")
	})

	t.Run("validates metrics", func(t *testing.T) {
		c := Config{}
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metrics")
	})
}

func TestConfig_GetRequestedMetrics(t *testing.T) {
	testCases := []struct{
		input    string
		expected []metrics.Metric
	}{
		{input: "", expected:[]metrics.Metric{}},
		{input: "memory", expected:[]metrics.Metric{metrics.Memory{}}},
		{input: "swap", expected:[]metrics.Metric{metrics.Swap{}}},
		{input: "cpu", expected:[]metrics.Metric{metrics.CPU{}}},
		{input: "disk", expected:[]metrics.Metric{metrics.Disk{}}},
		{input: "docker-stats", expected:[]metrics.Metric{metrics.DockerStat{}}},
		{input: "docker-health", expected:[]metrics.Metric{metrics.DockerHealth{}}},
		{input: "cpu,memory", expected:[]metrics.Metric{metrics.CPU{}, metrics.Memory{}}},
		{input: "cpu,foo", expected:[]metrics.Metric{metrics.CPU{}}},
		{input: ",", expected:[]metrics.Metric{}},
		{input: "cpu,", expected:[]metrics.Metric{metrics.CPU{}}},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := Config{Metrics: tc.input}
			output := c.GetRequestedMetrics()
			assert.ElementsMatch(t, tc.expected, output)
		})
	}
}

func TestConfig_GetExtraDimensions(t *testing.T) {
	c := Config{Id: "id"}
	dim := c.GetExtraDimensions()

	assert.Len(t, dim, 1)
	assert.Equal(t, dim[0].Name, "machine")
	assert.Equal(t, dim[0].Value, "id")
}
