package version

import "runtime/debug"

// Resolve exposes the internal build-metadata resolver for tests
func Resolve(version, commit string, readBuildInfo func() (*debug.BuildInfo, bool)) Info {
	return resolve(version, commit, readBuildInfo)
}
