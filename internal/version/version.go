package version

import "fmt"

var (
	Version   = "v0.6.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, BuildTime)
}
