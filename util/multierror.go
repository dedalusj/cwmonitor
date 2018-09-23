package util

import (
	"fmt"
	"strings"
)

// MultiError allows to collect and format multiple errors at once
type MultiError struct {
	Errors []error
}

// Add the given error to the current MultiError if not nil
func (m *MultiError) Add(err error) {
	if err == nil {
		return
	}
	if m.Errors == nil {
		m.Errors = make([]error, 0, 5)
	}
	m.Errors = append(m.Errors, err)
}

func (m MultiError) Error() string {
	if len(m.Errors) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n", m.Errors[0])
	}

	points := make([]string, len(m.Errors))
	for i, err := range m.Errors {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n",
		len(m.Errors), strings.Join(points, "\n\t"))
}

// ErrorOrNil returns nil if the MultiError is nil or no errors were attached to it
// otherwise returns the current MultiError
func (m *MultiError) ErrorOrNil() error {
	if m == nil {
		return nil
	}
	if len(m.Errors) == 0 {
		return nil
	}

	return m
}
