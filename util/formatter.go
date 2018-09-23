package util

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Formatter is a custom logrus formatter to emit logs in the form "<timestamp> | <level> | <message>"
type Formatter struct{}

// Format creates a formatted message for the given entry
func (f *Formatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.UTC().Format(time.RFC3339)
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message
	return []byte(fmt.Sprintf("%s | %-7s | %s\n", timestamp, level, message)), nil
}
