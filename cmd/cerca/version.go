package main

import (
	"fmt"
	"runtime/debug"
	"time"
)

type CommitInfo struct {
	Date string
	Hash string
}

func (c CommitInfo) String() string {
	return fmt.Sprintf("latest commit %s - %s", c.Date, c.Hash)
}

func version() {
	now := time.Now()
	// we can only get the git commit info when built by invoking `go build` (as opposed to being built & immediately
	// executed via `go run`)
	isBuiltBinary := false
	var commit CommitInfo
	var golangVersion string
	date := now.UTC().Format(time.UnixDate)
	if info, ok := debug.ReadBuildInfo(); ok {
		golangVersion = info.GoVersion
		for _, setting := range info.Settings {
			if setting.Key == "vcs" {
				isBuiltBinary = true
			}
			if setting.Key == "vcs.revision" {
				commit.Hash = setting.Value
			}
			if setting.Key == "vcs.time" {
				t, err := time.Parse(time.RFC3339, setting.Value)
				if err == nil {
					commit.Date = t.Format(time.DateOnly)
				}
			}
		}
	}
	inform("built on %s with %s", date, golangVersion)
	if isBuiltBinary {
		inform(commit.String())
	} else {
		inform("running with go run, git hash not available")
	}

}
