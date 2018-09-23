package util

import (
	"fmt"
)

// AppMetadata records  information like version, build time
// and build number for the current application
type AppMetadata struct {
	Version     string
	BuildTime   string
	BuildNumber string
}

// Format the app metadata
func (m AppMetadata) String() string {
	return fmt.Sprintf("%s (version), %s (build time), %s (build number)", m.Version, m.BuildTime, m.BuildNumber)
}
