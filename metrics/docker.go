package metrics

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

func computeCpu(stats types.StatsJSON) float64 {
	//compute the cpu usage percentage
	//via https://github.com/docker/docker/blob/e884a515e96201d4027a6c9c1b4fa884fc2d21a3/api/client/container/stats_helpers.go#L199-L212
	newCpuUsage := stats.CPUStats.CPUUsage.TotalUsage
	newSystemUsage := stats.CPUStats.SystemUsage
	cpuDiff := float64(newCpuUsage)
	systemDiff := float64(newSystemUsage)
	return cpuDiff / systemDiff * float64(len(stats.CPUStats.CPUUsage.PercpuUsage))

}

type DockerStat struct {
	client client.ContainerAPIClient
}

func (d DockerStat) Name() string {
	return "docker-stat"
}

func (d DockerStat) getStats(containerID string) (types.StatsJSON, error) {
	response, err := d.client.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return types.StatsJSON{}, errors.Wrapf(err, "failed to fetch statistics for container ID [%s]", containerID)
	}
	defer response.Body.Close()

	dec := json.NewDecoder(response.Body)
	var v *types.StatsJSON
	err = dec.Decode(&v)
	if err != nil {
		return types.StatsJSON{}, errors.Wrapf(err, "failed to fetch statistics for container ID [%s]", containerID)
	}

	return *v, nil
}

func (d DockerStat) Gather() (Data, error) {
	log.Debug("gathering docker stats")

	if d.client == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			return Data{}, errors.Wrap(err, "failed to create docker client from environment variables")
		}

		d.client = cli
	}

	containers, err := d.client.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to list containers")
	}

	data := Data{}
	for _, container := range containers {
		dimensions := make([]Dimension, 0, 1)
		var containerName Dimension
		if len(container.Names) > 0 {
			containerName, _ = NewDimension("ContainerName", strings.Trim(container.Names[0], "/"))
		} else {
			containerName, _ = NewDimension("ContainerName", container.ID)
		}
		dimensions = append(dimensions, containerName)

		stats, err := d.getStats(container.ID)
		if err != nil {
			log.Warnf("failed to fetch statistics for container ID [%s]: %s", container.ID, err)
			continue
		}

		cpuUtilization := NewDataPoint("CPUUtilization", computeCpu(stats), UnitPercent, dimensions...)
		data = append(data, &cpuUtilization)
		
		memoryUtilization := NewDataPoint("MemoryUtilization", float64(stats.MemoryStats.Usage), UnitBytes, dimensions...)
		data = append(data, &memoryUtilization)
	}

	return data, nil
}

type DockerHealth struct {
	client client.ContainerAPIClient
}

func (d DockerHealth) Name() string {
	return "docker-health"
}

func (d DockerHealth) Gather() (Data, error) {
	log.Debug("gathering docker health")

	if d.client == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			return Data{}, errors.Wrap(err, "failed to create docker client from environment variables")
		}

		d.client = cli
	}

	containers, err := d.client.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to list containers")
	}

	data := Data{}
	for _, container := range containers {
		dimensions := make([]Dimension, 0, 1)
		var containerName Dimension
		if len(container.Names) > 0 {
			containerName, _ = NewDimension("ContainerName", strings.Trim(container.Names[0], "/"))
		} else {
			containerName, _ = NewDimension("ContainerName", container.ID)
		}
		dimensions = append(dimensions, containerName)

		c, err := d.client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Warnf("failed to inspect container ID [%s]: %s", container.ID, err)
			continue
		}

		var value = 0.0
		if c.State != nil && c.State.Health != nil && strings.ToLower(c.State.Health.Status) == "healthy" {
			value = 1.0
		}
		healthDataPoint := NewDataPoint("Health", value, UnitCount, dimensions...)
		data = append(data, &healthDataPoint)
	}

	return data, nil
}
