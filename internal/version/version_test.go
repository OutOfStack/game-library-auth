package version_test

import (
	"runtime/debug"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestResolveDefaultsToDevWithoutBuildMetadata(t *testing.T) {
	info := version.Resolve("dev", "", stubBuildInfo(nil, false))

	assert.Equal(t, "dev", info.Version)
	assert.Empty(t, info.Commit)
	assert.Empty(t, info.GoVersion)
}

func TestResolveUsesLinkerValuesBeforeBuildInfo(t *testing.T) {
	info := version.Resolve("v1.2.3", "abc123", stubBuildInfo(&debug.BuildInfo{
		GoVersion: "go1.26.2",
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "def456"},
			{Key: "vcs.time", Value: "2026-05-03T10:00:00Z"},
			{Key: "vcs.modified", Value: "false"},
		},
	}, true))

	assert.Equal(t, "v1.2.3", info.Version)
	assert.Equal(t, "abc123", info.Commit)
	assert.Equal(t, "go1.26.2", info.GoVersion)
	assert.Equal(t, "2026-05-03T10:00:00Z", info.BuildTime)
	assert.Equal(t, "false", info.Modified)
}

func TestResolveFallsBackToVCSRevision(t *testing.T) {
	info := version.Resolve("dev", "", stubBuildInfo(&debug.BuildInfo{
		Main:      debug.Module{Version: "(devel)"},
		GoVersion: "go1.26.2",
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "def456"},
		},
	}, true))

	assert.Equal(t, "def456", info.Version)
	assert.Equal(t, "def456", info.Commit)
	assert.Equal(t, "go1.26.2", info.GoVersion)
}

func TestResolveTreatsUnknownCommitAsMissing(t *testing.T) {
	info := version.Resolve("dev", "unknown", stubBuildInfo(&debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "def456"},
		},
	}, true))

	assert.Equal(t, "def456", info.Version)
	assert.Equal(t, "def456", info.Commit)
}

func TestResolveFallsBackToModuleVersion(t *testing.T) {
	info := version.Resolve("dev", "", stubBuildInfo(&debug.BuildInfo{
		Main:      debug.Module{Version: "v1.4.0"},
		GoVersion: "go1.26.2",
	}, true))

	assert.Equal(t, "v1.4.0", info.Version)
	assert.Empty(t, info.Commit)
	assert.Equal(t, "go1.26.2", info.GoVersion)
}

func stubBuildInfo(bi *debug.BuildInfo, ok bool) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return bi, ok
	}
}
