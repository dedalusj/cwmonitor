package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppMetadata_String(t *testing.T) {
	m := AppMetadata{Version: "1.0.0", BuildTime: "20180920T101010Z", BuildNumber: "1234"}
	s := m.String()
	assert.Contains(t, s, m.Version)
	assert.Contains(t, s, m.BuildTime)
	assert.Contains(t, s, m.BuildNumber)
}
