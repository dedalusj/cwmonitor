package monitor

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dedalusj/cwmonitor/metrics"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestConfig_validate(t *testing.T) {
	t.Run("validates name", func(t *testing.T) {
		c := Config{}
		err := c.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("validates interval", func(t *testing.T) {
		c := Config{}
		err := c.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "interval")
	})

	t.Run("validates hostid", func(t *testing.T) {
		c := Config{}
		err := c.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hostid")
	})

	t.Run("validates metrics", func(t *testing.T) {
		c := Config{}
		err := c.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metrics")
	})

	t.Run("valid", func(t *testing.T) {
		c := Config{
			Namespace: "namespace",
			Interval:  time.Minute,
			HostId:    "id",
			Metrics:   "cpu,memory",
		}
		err := c.validate()
		assert.NoError(t, err)
	})
}

func TestConfig_getRequestedMetrics(t *testing.T) {
	testCases := []struct {
		input    string
		expected []metrics.Metric
	}{
		{input: "", expected: []metrics.Metric{}},
		{input: "memory", expected: []metrics.Metric{metrics.Memory{}}},
		{input: "swap", expected: []metrics.Metric{metrics.Swap{}}},
		{input: "cpu", expected: []metrics.Metric{metrics.CPU{}}},
		{input: "disk", expected: []metrics.Metric{metrics.Disk{}}},
		{input: "docker-stats", expected: []metrics.Metric{metrics.DockerStat{}}},
		{input: "docker-health", expected: []metrics.Metric{metrics.DockerHealth{}}},
		{input: "cpu,memory", expected: []metrics.Metric{metrics.CPU{}, metrics.Memory{}}},
		{input: "cpu,foo", expected: []metrics.Metric{metrics.CPU{}}},
		{input: ",", expected: []metrics.Metric{}},
		{input: "cpu,", expected: []metrics.Metric{metrics.CPU{}}},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := Config{Metrics: tc.input}
			output := c.getRequestedMetrics()
			assert.ElementsMatch(t, tc.expected, output)
		})
	}
}

func TestConfig_getExtraDimensions(t *testing.T) {
	c := Config{HostId: "id"}
	dim := c.getExtraDimensions()

	assert.Len(t, dim, 1)
	assert.Equal(t, dim[0].Name, "Host")
	assert.Equal(t, dim[0].Value, "id")
}

func TestConfig_logConfig(t *testing.T) {
	hook := test.NewGlobal()
	Config{
		Namespace:   "namespace",
		Interval:    time.Second * 30,
		HostId:      "id",
		Metrics:     "cpu,memory",
		DockerLabel: "a_label",
	}.logConfig()

	messages := make([]string, len(hook.AllEntries()))
	for i, e := range hook.AllEntries() {
		messages[i] = e.Message
	}
	logOutput := strings.Join(messages, "\n")

	assert.Contains(t, logOutput, "namespace")
	assert.Contains(t, logOutput, "30s")
	assert.Contains(t, logOutput, "id")
	assert.Contains(t, logOutput, "cpu,memory")
	assert.Contains(t, logOutput, "a_label")
}
