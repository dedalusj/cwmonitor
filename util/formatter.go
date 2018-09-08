package util

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Formatter struct{}

func (f *Formatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.UTC().Format(time.RFC3339)
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message
	return []byte(fmt.Sprintf("%s | %-7s | %s\n", timestamp, level, message)), nil
}
