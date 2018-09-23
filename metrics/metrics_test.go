package metrics

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDimension(t *testing.T) {
	t.Run("non empty name and value", func(t *testing.T) {
		d, err := NewDimension("a", "1")
		assert.NoError(t, err)
		assert.Equal(t, Dimension{"a", "1"}, d)
	})

	t.Run("non empty name and empty value", func(t *testing.T) {
		d, err := NewDimension("a", "")
		assert.NoError(t, err)
		assert.Equal(t, Dimension{"a", ""}, d)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewDimension("", "1")
		assert.Error(t, err)
	})

	t.Run("blank name", func(t *testing.T) {
		_, err := NewDimension("   ", "1")
		assert.Error(t, err)
	})
}

func TestMapToDimension(t *testing.T) {
	t.Run("valid map", func(t *testing.T) {
		input := map[string]string{"a": "1", "b": "2", "c": ""}
		dimensions, err := MapToDimensions(input)

		assert.NoError(t, err)
		assert.Len(t, dimensions, 3)
		assert.Contains(t, dimensions, Dimension{"a", "1"})
		assert.Contains(t, dimensions, Dimension{"b", "2"})
		assert.Contains(t, dimensions, Dimension{"c", ""})
	})

	t.Run("empty name", func(t *testing.T) {
		input := map[string]string{"a": "1", "": "2"}
		_, err := MapToDimensions(input)

		assert.Error(t, err)
	})
}

func TestNewDataPoint(t *testing.T) {
	t.Run("no extra dimensions", func(t *testing.T) {
		p := NewDataPoint("a", 5.0, UnitCount)
		assert.Equal(t, "a", p.Name)
		assert.Equal(t, 5.0, p.Value)
		assert.Equal(t, UnitCount, p.Unit)
		assert.NotZero(t, p.Timestamp)
	})

	t.Run("single dimensions", func(t *testing.T) {
		dim := Dimension{"a", "1"}
		p := NewDataPoint("a", 5.0, UnitCount, dim)
		assert.Equal(t, "a", p.Name)
		assert.Equal(t, 5.0, p.Value)
		assert.Equal(t, UnitCount, p.Unit)
		assert.NotZero(t, p.Timestamp)
		assert.Len(t, p.Dimensions, 1)
		assert.Equal(t, dim, p.Dimensions[0])
	})

	t.Run("multiple dimensions", func(t *testing.T) {
		dim1, dim2 := Dimension{"a", "1"}, Dimension{"b", "2"}
		p := NewDataPoint("a", 5.0, UnitCount, dim1, dim2)
		assert.Equal(t, "a", p.Name)
		assert.Equal(t, 5.0, p.Value)
		assert.Equal(t, UnitCount, p.Unit)
		assert.NotZero(t, p.Timestamp)
		assert.Len(t, p.Dimensions, 2)
		assert.Equal(t, dim1, p.Dimensions[0])
		assert.Equal(t, dim2, p.Dimensions[1])
	})
}

func TestPoint_AddDimension(t *testing.T) {
	p := NewDataPoint("a", 5.0, UnitCount)
	assert.Len(t, p.Dimensions, 0)

	dim1, dim2 := Dimension{"a", "1"}, Dimension{"b", "2"}
	p.AddDimensions(dim1, dim2)
	assert.Len(t, p.Dimensions, 2)
	assert.Equal(t, dim1, p.Dimensions[0])
	assert.Equal(t, dim2, p.Dimensions[1])
}

func TestData_AddDimensions(t *testing.T) {
	p1 := NewDataPoint("a", 5.0, UnitCount)
	p2 := NewDataPoint("b", 3.0, UnitCount)

	data := Data([]*Point{&p1, &p2})
	assert.Len(t, data[0].Dimensions, 0)
	assert.Len(t, data[1].Dimensions, 0)

	dim1, dim2 := Dimension{"a", "1"}, Dimension{"b", "2"}
	data.AddDimensions(dim1, dim2)

	assert.Len(t, data[0].Dimensions, 2)
	assert.Equal(t, dim1, data[0].Dimensions[0])
	assert.Equal(t, dim2, data[0].Dimensions[1])

	assert.Len(t, data[1].Dimensions, 2)
	assert.Equal(t, dim1, data[1].Dimensions[0])
	assert.Equal(t, dim2, data[1].Dimensions[1])
}

func TestData_Batch(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		data := Data{}
		batches := data.Batch(3)
		assert.Len(t, batches, 0)
	})

	t.Run("batched", func(t *testing.T) {
		data := Data{}
		for i := 0; i <= 6; i++ {
			data = append(data, &Point{Name: strconv.Itoa(i)})
		}

		batches := data.Batch(3)

		assert.Len(t, batches, 3)
		assert.Len(t, batches[0], 3)
		assert.Len(t, batches[1], 3)
		assert.Len(t, batches[2], 1)
	})
}
