package metrics

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Unit for a reported data point
type Unit string

const (
	UnitSeconds         Unit = "Seconds"
	UnitMicroseconds    Unit = "Microseconds"
	UnitMilliseconds    Unit = "Milliseconds"
	UnitBytes           Unit = "Bytes"
	UnitKilobytes       Unit = "Kilobytes"
	UnitMegabytes       Unit = "Megabytes"
	UnitGigabytes       Unit = "Gigabytes"
	UnitTerabytes       Unit = "Terabytes"
	UnitBits            Unit = "Bits"
	UnitKilobits        Unit = "Kilobits"
	UnitMegabits        Unit = "Megabits"
	UnitGigabits        Unit = "Gigabits"
	UnitTerabits        Unit = "Terabits"
	UnitPercent         Unit = "Percent"
	UnitCount           Unit = "Count"
	UnitBytesSecond     Unit = "Bytes/Second"
	UnitKilobytesSecond Unit = "Kilobytes/Second"
	UnitMegabytesSecond Unit = "Megabytes/Second"
	UnitGigabytesSecond Unit = "Gigabytes/Second"
	UnitTerabytesSecond Unit = "Terabytes/Second"
	UnitBitsSecond      Unit = "Bits/Second"
	UnitKilobitsSecond  Unit = "Kilobits/Second"
	UnitMegabitsSecond  Unit = "Megabits/Second"
	UnitGigabitsSecond  Unit = "Gigabits/Second"
	UnitTerabitsSecond  Unit = "Terabits/Second"
	UnitCountSecond     Unit = "Count/Second"
	UnitNone            Unit = "None"
)

// Dimension for a collected data point
type Dimension struct {
	Name  string
	Value string
}

// NewDimension creates a new Dimension with the given name and value or error if name is blank
func NewDimension(name, value string) (Dimension, error) {
	sanitisedName := strings.TrimSpace(name)
	if sanitisedName == "" {
		return Dimension{}, errors.New("dimension name cannot be blank")
	}

	return Dimension{Name: sanitisedName, Value: value}, nil
}

// MapToDimensions creates a list of dimensions from a map of key, value strings
func MapToDimensions(input map[string]string) ([]Dimension, error) {
	dimensions := make([]Dimension, 0, len(input))
	for k, v := range input {
		d, err := NewDimension(k, v)
		if err != nil {
			return []Dimension{}, err
		}
		dimensions = append(dimensions, d)
	}

	return dimensions, nil
}

// Point represent an instance of collected data
type Point struct {
	Name       string
	Dimensions []Dimension
	Timestamp  time.Time
	Value      float64
	Unit       Unit
}

// NewDataPoint creates a new data Point
func NewDataPoint(name string, value float64, unit Unit, dimensions ...Dimension) Point {
	p := Point{Name: name, Value: value, Unit: unit, Timestamp: time.Now().UTC()}
	p.AddDimensions(dimensions...)
	return p
}

// AddDimensions adds the provided dimensions to the current data Point
func (p *Point) AddDimensions(dimensions ...Dimension) {
	if p.Dimensions == nil {
		p.Dimensions = make([]Dimension, 0, len(dimensions))
	}

	p.Dimensions = append(p.Dimensions, dimensions...)
}

// Data is a collection of collected data points
type Data []*Point

// AddDimensions adds the provided dimensions to all the points in the Data collection
func (data *Data) AddDimensions(dimensions ...Dimension) {
	for _, p := range *data {
		p.AddDimensions(dimensions...)
	}
}

// Batch partitions the data collection into a list of smaller collection each
// of maximum size given by batchSize
func (data Data) Batch(batchSize int) []Data {
	if len(data) == 0 {
		return []Data{}
	}

	var batches []Data
	var d = data
	for batchSize < len(d) {
		d, batches = d[batchSize:], append(batches, d[0:batchSize:batchSize])
	}
	batches = append(batches, d)
	return batches
}

// Metric is an interface for any specific implementation that can gather
// statistics and return data points
type Metric interface {
	Name() string
	Gather() (Data, error)
}
