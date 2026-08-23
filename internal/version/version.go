package version

import "fmt"

var (
	// Version is the current semantic version of Tarak.
	Version = "1.0.8"

	// Commit is the git commit hash at build time.
	Commit = "0d7e080"

	// BuildDate is the timestamp of the build.
	BuildDate = "2026-08-23T08:55:27Z"

	// Author is the developer or organization who built this version.
	Author = "vikukumar"
)

// String returns a formatted version string.
func String() string {
	return fmt.Sprintf("v%s (commit: %s, built: %s, author: %s)", Version, Commit, BuildDate, Author)
}

// Info returns a map of the version details.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildDate": BuildDate,
		"author":    Author,
	}
}
