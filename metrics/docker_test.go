package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type DockerMockClient struct {
	mock.Mock
	client.ContainerAPIClient
}

func (m DockerMockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	args := m.Called(options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m DockerMockClient) ContainerStats(ctx context.Context, container string, stream bool) (types.ContainerStats, error) {
	args := m.Called(container, stream)
	return args.Get(0).(types.ContainerStats), args.Error(1)
}

func (m DockerMockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	args := m.Called(containerID)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func makeContainer(containerId string) types.Container {
	return types.Container{ID: containerId, Names: []string{"name-"+containerId}}
}

func makeContainerDimensions(id string) []Dimension {
	return []Dimension{
		{Name: "Container", Value: "name-"+id},
	}
}

func makeContainerStats(numCPUs int, totalUsage, systemUsage, memoryUsage uint64) types.ContainerStats {
	stats := types.StatsJSON{}
	stats.CPUStats.CPUUsage.PercpuUsage = make([]uint64, numCPUs)
	stats.CPUStats.CPUUsage.TotalUsage = totalUsage
	stats.CPUStats.SystemUsage = systemUsage
	stats.MemoryStats.Usage = memoryUsage
	b, _ := json.Marshal(stats)
	return types.ContainerStats{
		Body: ioutil.NopCloser(bytes.NewReader(b)),
	}
}

func makeContainerDetails(healthState string) types.ContainerJSON {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Health: &types.Health{Status: healthState},
			},
		},
		Mounts:            []types.MountPoint{},
		Config:            &container.Config{},
		NetworkSettings:   &types.NetworkSettings{},
	}
}

func TestGetDimensionsFromContainer(t *testing.T) {
	t.Run("uses name if available", func(t *testing.T) {
		c := types.Container{ID: "id", Names: []string{"name"}}
		dims := GetDimensionsFromContainer(c)
		assert.Equal(t, []Dimension{{Name: "Container", Value: "name"}}, dims)
	})

	t.Run("strip slashes from name", func(t *testing.T) {
		c := types.Container{ID: "id", Names: []string{"/name"}}
		dims := GetDimensionsFromContainer(c)
		assert.Equal(t, []Dimension{{Name: "Container", Value: "name"}}, dims)
	})

	t.Run("use id if name not available", func(t *testing.T) {
		c := types.Container{ID: "id"}
		dims := GetDimensionsFromContainer(c)
		assert.Equal(t, []Dimension{{Name: "Container", Value: "id"}}, dims)
	})
}

func TestDockerMetric_InitClient(t *testing.T) {
	t.Run("docker-stat init client correctly initialise the docker client", func(t *testing.T) {
		d := DockerStat{}
		d.InitClient()
		assert.NotNil(t, d.client)
	})

	t.Run("docker-health init client correctly initialise the docker client", func(t *testing.T) {
		d := DockerHealth{}
		d.InitClient()
		assert.NotNil(t, d.client)
	})
}

func TestDockerStat_Name(t *testing.T) {
	d := DockerStat{}
	assert.Equal(t, "docker-stat", d.Name())
}

func TestDockerStat_Gather(t *testing.T) {
	t.Run("empty container list", func(t *testing.T) {
		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return([]types.Container{}, nil)

		d := DockerStat{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 0)
		mockClient.AssertExpectations(t)
	})

	t.Run("error for container list", func(t *testing.T) {
		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return([]types.Container{}, errors.New("an error"))

		d := DockerStat{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.Error(t, err)
		assert.Len(t, data, 0)
		mockClient.AssertExpectations(t)
	})

	t.Run("stats from multiple container", func(t *testing.T) {
		containerId1, containerId2 := "c1", "c2"
		containers := []types.Container{makeContainer(containerId1), makeContainer(containerId2)}
		stats1 := makeContainerStats(2, 100, 200, 200)
		stats2 := makeContainerStats(2, 50, 200, 400)

		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return(containers, nil)
		mockClient.On("ContainerStats", containerId1, false).Return(stats1, nil)
		mockClient.On("ContainerStats", containerId2, false).Return(stats2, nil)

		expectedDimensions1 := makeContainerDimensions(containerId1)
		expectedDimensions2 := makeContainerDimensions(containerId2)

		d := DockerStat{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 4)

		assert.Equal(t, data[0].Name, "CPUUtilization")
		assert.Equal(t, string(data[0].Unit), string(UnitPercent))
		assert.Equal(t, 1.0, data[0].Value)
		assert.Equal(t, expectedDimensions1, data[0].Dimensions)

		assert.Equal(t, data[1].Name, "MemoryUtilization")
		assert.Equal(t, string(data[1].Unit), string(UnitBytes))
		assert.Equal(t, 200.0, data[1].Value)
		assert.Equal(t, expectedDimensions1, data[1].Dimensions)

		assert.Equal(t, data[2].Name, "CPUUtilization")
		assert.Equal(t, string(data[2].Unit), string(UnitPercent))
		assert.Equal(t, 0.5, data[2].Value)
		assert.Equal(t, expectedDimensions2, data[2].Dimensions)

		assert.Equal(t, data[3].Name, "MemoryUtilization")
		assert.Equal(t, string(data[3].Unit), string(UnitBytes))
		assert.Equal(t, 400.0, data[3].Value)
		assert.Equal(t, expectedDimensions2, data[3].Dimensions)

		mockClient.AssertExpectations(t)
	})

	t.Run("stats returns error", func(t *testing.T) {
		containerId := "c"
		containers := []types.Container{makeContainer(containerId)}

		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return(containers, nil)
		mockClient.On("ContainerStats", containerId, false).Return(types.ContainerStats{}, errors.New("an error"))

		d := DockerStat{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 0)

		mockClient.AssertExpectations(t)
	})

	t.Run("stats return invalid payload", func(t *testing.T) {
		containerId := "c"
		containers := []types.Container{makeContainer(containerId)}
		stats := types.ContainerStats{
			Body: ioutil.NopCloser(bytes.NewReader([]byte("invalid"))),
		}

		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return(containers, nil)
		mockClient.On("ContainerStats", containerId, false).Return(stats, nil)

		d := DockerStat{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 0)

		mockClient.AssertExpectations(t)
	})
}

func TestDockerHealth_Name(t *testing.T) {
	d := DockerHealth{}
	assert.Equal(t, "docker-health", d.Name())
}

func TestDockerHealth_Gather(t *testing.T) {
	t.Run("empty container list", func(t *testing.T) {
		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return([]types.Container{}, nil)

		d := DockerHealth{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 0)
		mockClient.AssertExpectations(t)
	})

	t.Run("error for container list", func(t *testing.T) {
		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return([]types.Container{}, errors.New("an error"))

		d := DockerHealth{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.Error(t, err)
		assert.Len(t, data, 0)
		mockClient.AssertExpectations(t)
	})

	t.Run("health from multiple container", func(t *testing.T) {
		containerId1, containerId2 := "c1", "c2"
		containers := []types.Container{makeContainer(containerId1), makeContainer(containerId2)}
		healthyContainer := makeContainerDetails("Healthy")
		unhealthyContainer := makeContainerDetails("Unhealthy")

		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return(containers, nil)
		mockClient.On("ContainerInspect", containerId1).Return(healthyContainer, nil)
		mockClient.On("ContainerInspect", containerId2).Return(unhealthyContainer, nil)

		expectedDimensions1 := makeContainerDimensions(containerId1)
		expectedDimensions2 := makeContainerDimensions(containerId2)

		d := DockerHealth{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 2)

		assert.Equal(t, data[0].Name, "Health")
		assert.Equal(t, string(data[0].Unit), string(UnitCount))
		assert.Equal(t, 1.0, data[0].Value)
		assert.Equal(t, expectedDimensions1, data[0].Dimensions)

		assert.Equal(t, data[1].Name, "Health")
		assert.Equal(t, string(data[1].Unit), string(UnitCount))
		assert.Equal(t, 0.0, data[1].Value)
		assert.Equal(t, expectedDimensions2, data[1].Dimensions)

		mockClient.AssertExpectations(t)
	})

	t.Run("error from container inspect", func(t *testing.T) {
		containerId1 := "c1"
		containers := []types.Container{makeContainer(containerId1)}

		mockClient := new(DockerMockClient)
		mockClient.On("ContainerList", types.ContainerListOptions{All: false}).Return(containers, nil)
		mockClient.On("ContainerInspect", containerId1).Return(types.ContainerJSON{}, errors.New("an error"))

		d := DockerHealth{dockerMetric{client: mockClient}}
		data, err := d.Gather()

		assert.NoError(t, err)
		assert.Len(t, data, 0)

		mockClient.AssertExpectations(t)
	})
}
