package app

// Build-time variables set via ldflags.
var (
	BuildVersion = "dev"
	BuildCommit  = "unknown"
	BuildDate    = "unknown"
)

// BuildInfo returns a human-readable build string.
func BuildInfo() string {
	return BuildVersion + " (" + BuildCommit + ", " + BuildDate + ")"
}
