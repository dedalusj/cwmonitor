package util

import (
	"fmt"
	"strings"
)

type MultiError struct {
	Errors []error
}

func (me *MultiError) Add(err error) {
	if err == nil {
		return
	}
	if me.Errors == nil {
		me.Errors = make([]error, 0, 5)
	}
	me.Errors = append(me.Errors, err)
}

func (me MultiError) Error() string {
	if len(me.Errors) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", me.Errors[0])
	}

	points := make([]string, len(me.Errors))
	for i, err := range me.Errors {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(me.Errors), strings.Join(points, "\n\t"))
}

func (me *MultiError) ErrorOrNil() error {
	if me == nil {
		return nil
	}
	if len(me.Errors) == 0 {
		return nil
	}

	return me
}

