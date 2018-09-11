package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiError_Add(t *testing.T) {
	t.Run("adding nil error", func(t *testing.T) {
		me := MultiError{}
		me.Add(nil)
		assert.Len(t, me.Errors, 0)
	})

	t.Run("adding multiple errors", func(t *testing.T) {
		me := MultiError{}
		assert.Len(t, me.Errors, 0)

		err1 := errors.New("error 1")
		me.Add(err1)
		assert.Len(t, me.Errors, 1)
		assert.EqualError(t, me.Errors[0], err1.Error())

		err2 := errors.New("error 2")
		me.Add(err2)
		assert.Len(t, me.Errors, 2)
		assert.EqualError(t, me.Errors[1], err2.Error())
	})
}

func TestMultiError_Error(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		me := MultiError{}
		me.Add(errors.New("douglas"))

		assert.Contains(t, me.Error(), "1 error")
		assert.Contains(t, me.Error(), "douglas")
	})

	t.Run("multiple error", func(t *testing.T) {
		me := MultiError{}
		me.Add(errors.New("douglas"))
		me.Add(errors.New("adams"))

		assert.Contains(t, me.Error(), "2 errors")
		assert.Contains(t, me.Error(), "douglas")
		assert.Contains(t, me.Error(), "adams")
	})
}

func TestMultiError_ErrorOrNil(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		me := MultiError{}
		assert.Nil(t, me.ErrorOrNil())
	})

	t.Run("single error", func(t *testing.T) {
		me := MultiError{}
		me.Add(errors.New("error 1"))
		assert.Error(t, me.ErrorOrNil())
	})

	t.Run("nil error", func(t *testing.T) {
		me := (*MultiError)(nil)
		assert.NoError(t, me.ErrorOrNil())
	})
}
