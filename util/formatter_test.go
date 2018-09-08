package util

import (
	"fmt"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_Format(t *testing.T) {
	message := "a message"
	now := time.Now()
	expected := []byte(fmt.Sprintf("%s | %-7s | %s\n", time.Now().UTC().Format(time.RFC3339), "INFO", message))

	formatter := Formatter{}
	result, err := formatter.Format(&log.Entry{Time: now, Level: log.InfoLevel, Message: message})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
