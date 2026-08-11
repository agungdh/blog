package version

import "fmt"

var (
	Version   = "v0.5.5"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, BuildTime)
}
