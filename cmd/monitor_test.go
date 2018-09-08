package cmd

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mon-put-data/metrics"
)

func TestBuildMetrics(t *testing.T) {
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
			output := buildMetrics(tc.input)
			assert.ElementsMatch(t, tc.expected, output)
		})
	}
}

type mockMetric struct {
	mock.Mock
}

func (m mockMetric) Name() string {
	return "mock"
}

func (m mockMetric) Gather() (metrics.Data, error) {
	args := m.Called()
	return args.Get(0).(metrics.Data), args.Error(1)
}

func TestGather(t *testing.T) {
	t.Run("valid metrics", func(t *testing.T) {
		m := new(mockMetric)
		m.On("Gather").Return(metrics.Data{&metrics.Point{}}, nil)

		validMetrics := []metrics.Metric{m, m}
		data := Gather(validMetrics)
		assert.Len(t, data, 2)
	})

	t.Run("failing metric", func(t *testing.T) {
		m := new(mockMetric)
		m.On("Gather").Return(metrics.Data{&metrics.Point{}}, nil)

		f := new(mockMetric)
		f.On("Gather").Return(metrics.Data{&metrics.Point{}}, errors.New("an error"))

		validMetrics := []metrics.Metric{m, f}
		data := Gather(validMetrics)
		assert.Len(t, data, 1)
	})
}

type mockCloudWatchClient struct {
	mock.Mock

	cloudwatchiface.CloudWatchAPI
}

func (m mockCloudWatchClient) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatch.PutMetricDataOutput), args.Error(1)
}

func createDataAndExpectedCWInput(numDataPoints int, timestamp time.Time, namespace string) (metrics.Data, []*cloudwatch.MetricDatum) {
	data := make([]*metrics.Point, numDataPoints)
	expected := make([]*cloudwatch.MetricDatum, numDataPoints)
	for i := range data {
		p := metrics.Point{
			Name: strconv.Itoa(i),
			Unit: metrics.UnitNone,
			Value: float64(i),
			Timestamp: timestamp,
			Dimensions: []metrics.Dimension{{Name: "name", Value: strconv.Itoa(i)}},
		}
		data[i] = &p

		cw := cloudwatch.MetricDatum{
			MetricName: aws.String(strconv.Itoa(i)),
			Unit: aws.String("None"),
			Value: aws.Float64(float64(i)),
			Timestamp: aws.Time(timestamp),
			Dimensions: []*cloudwatch.Dimension{{Name: aws.String("name"), Value: aws.String(strconv.Itoa(i))}},
		}
		expected[i] = &cw
	}

	return data, expected
}

func TestPut(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		mockClient := new(mockCloudWatchClient)
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{}).Return(&cloudwatch.PutMetricDataOutput{}, nil)

		err := Put(metrics.Data{}, mockClient, "namespace")
		assert.NoError(t, err)
		mockClient.AssertNotCalled(t, "PutMetricData")
	})

	t.Run("single batch", func(t *testing.T) {
		numDataPoints := 5
		timestamp := time.Date(2018, 9, 1, 10, 0, 0, 0, time.UTC)
		namespace := "a namespace"

		data, expected := createDataAndExpectedCWInput(numDataPoints, timestamp, namespace)

		mockClient := new(mockCloudWatchClient)
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected,
		}).Return(&cloudwatch.PutMetricDataOutput{}, nil).Once()

		err := Put(data, mockClient, namespace)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("multiple batches", func(t *testing.T) {
		numDataPoints := 25
		timestamp := time.Date(2018, 9, 1, 10, 0, 0, 0, time.UTC)
		namespace := "a namespace"

		data, expected := createDataAndExpectedCWInput(numDataPoints, timestamp, namespace)

		mockClient := new(mockCloudWatchClient)
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[:20],
		}).Return(&cloudwatch.PutMetricDataOutput{}, nil).Once()
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[20:],
		}).Return(&cloudwatch.PutMetricDataOutput{}, nil).Once()

		err := Put(data, mockClient, namespace)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("multiple batches single error", func(t *testing.T) {
		numDataPoints := 25
		timestamp := time.Date(2018, 9, 1, 10, 0, 0, 0, time.UTC)
		namespace := "a namespace"

		data, expected := createDataAndExpectedCWInput(numDataPoints, timestamp, namespace)

		mockClient := new(mockCloudWatchClient)
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[:20],
		}).Return(&cloudwatch.PutMetricDataOutput{}, errors.New("an error")).Once()
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[20:],
		}).Return(&cloudwatch.PutMetricDataOutput{}, nil).Once()

		err := Put(data, mockClient, namespace)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "an error")
		mockClient.AssertExpectations(t)
	})

	t.Run("multiple batches multiple errors", func(t *testing.T) {
		numDataPoints := 25
		timestamp := time.Date(2018, 9, 1, 10, 0, 0, 0, time.UTC)
		namespace := "a namespace"

		data, expected := createDataAndExpectedCWInput(numDataPoints, timestamp, namespace)

		mockClient := new(mockCloudWatchClient)
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[:20],
		}).Return(&cloudwatch.PutMetricDataOutput{}, errors.New("first error")).Once()
		mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: expected[20:],
		}).Return(&cloudwatch.PutMetricDataOutput{}, errors.New("second error")).Once()

		err := Put(data, mockClient, namespace)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first error")
		assert.Contains(t, err.Error(), "second error")
		mockClient.AssertExpectations(t)
	})
}

func TestMonitor(t *testing.T) {
	numDataPoints := 2
	timestamp := time.Date(2018, 9, 1, 10, 0, 0, 0, time.UTC)
	namespace := "a namespace"
	extraDimension := metrics.Dimension{Name: "id", Value: "localhost"}

	data, expected := createDataAndExpectedCWInput(numDataPoints, timestamp, namespace)
	for _, e := range expected {
		d := cloudwatch.Dimension{Name: aws.String(extraDimension.Name), Value: aws.String(extraDimension.Value)}
		e.Dimensions = append(e.Dimensions, &d)
	}

	m := new(mockMetric)
	m.On("Gather").Return(metrics.Data(data), nil)

	mockClient := new(mockCloudWatchClient)
	mockClient.On("PutMetricData", &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: expected,
	}).Return(&cloudwatch.PutMetricDataOutput{}, nil).Once()

	Monitor([]metrics.Metric{m}, []metrics.Dimension{extraDimension}, mockClient, namespace)

	m.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

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
