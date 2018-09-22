package util

import (
	"fmt"
)

type AppMetadata struct {
	Version     string
	BuildTime   string
	BuildNumber string
}

func (m AppMetadata) String() string {
	return fmt.Sprintf("%s (version), %s (build time), %s (build number)", m.Version, m.BuildTime, m.BuildNumber)
}
