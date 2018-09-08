package metrics

import (
	"time"

	"github.com/pkg/errors"
)

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

type Dimension struct {
	Name  string
	Value string
}

func NewDimension(name, value string) (Dimension, error) {
	if name == "" {
		return Dimension{}, errors.New("dimension name cannot be blank")
	}

	return Dimension{Name: name, Value: value}, nil
}

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

type Point struct {
	Name       string
	Dimensions []Dimension
	Timestamp  time.Time
	Value      float64
	Unit       Unit
}

func NewDataPoint(name string, value float64, unit Unit, dimensions ...Dimension) Point {
	p := Point{Name: name, Value: value, Unit: unit, Timestamp: time.Now().UTC()}
	p.AddDimensions(dimensions...)
	return p
}

func (p *Point) AddDimensions(dimensions ...Dimension) {
	if p.Dimensions == nil {
		p.Dimensions = make([]Dimension, 0, len(dimensions))
	}

	p.Dimensions = append(p.Dimensions, dimensions...)
}

type Data []*Point

func (data *Data) AddDimensions(dimensions ...Dimension) {
	for _, p := range *data {
		p.AddDimensions(dimensions...)
	}
}

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

type Metric interface {
	Name() string
	Gather() (Data, error)
}
