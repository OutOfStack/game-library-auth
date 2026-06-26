package version

import (
	"runtime/debug"
	"sync"
)

const devVersion = "dev"

var (
	appVersion = devVersion
	appCommit  = ""

	cachedInfo Info
	cacheOnce  sync.Once
)

// Info contains build and source metadata for the running binary
type Info struct {
	Version   string `json:"version,omitempty"`
	Commit    string `json:"commit,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
	BuildTime string `json:"buildTime,omitempty"`
	Modified  string `json:"modified,omitempty"`
}

// Get returns build metadata embedded by the Go toolchain and release build
func Get() Info {
	cacheOnce.Do(func() {
		cachedInfo = resolve(appVersion, appCommit, debug.ReadBuildInfo)
	})
	return cachedInfo
}

func resolve(version, commit string, readBuildInfo func() (*debug.BuildInfo, bool)) Info {
	info := Info{
		Version: version,
		Commit:  commit,
	}

	bi, ok := readBuildInfo()
	if ok && bi != nil {
		info.GoVersion = bi.GoVersion

		if isDevVersion(info.Version) && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			info.Version = bi.Main.Version
		}

		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				if !hasKnownCommit(info.Commit) {
					info.Commit = setting.Value
				}
			case "vcs.time":
				info.BuildTime = setting.Value
			case "vcs.modified":
				info.Modified = setting.Value
			}
		}
	}

	if isDevVersion(info.Version) && hasKnownCommit(info.Commit) {
		info.Version = info.Commit
	}

	if info.Version == "" {
		info.Version = devVersion
	}

	if info.Modified == "true" && info.Version != devVersion {
		info.Version += "+dirty"
	}

	return info
}

func isDevVersion(v string) bool {
	return v == "" || v == devVersion
}

func hasKnownCommit(commit string) bool {
	return commit != "" && commit != "unknown"
}
