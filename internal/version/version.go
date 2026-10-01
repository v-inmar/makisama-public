package version

import (
	"fmt"
	"runtime"
)

// These variables will be populated at build time via -ldflags
var (
	Version   = "unknown"
	CommitHex = "unknown"
	BuildTime = "unknown"
)

// Info returns a formatted string of the build information.
func Info() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nBuilt: %s\nGo: %s", Version, CommitHex, BuildTime, runtime.Version())
}
