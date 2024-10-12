package service

import (
	"encoding/hex"
	"runtime/debug"
	"time"
)

var (
	gitCommit  string // overwritten by -ldflag "-X 'github.com/ATMackay/psql-ledger/service.gitCommit=$commit_hash'"
	commitDate string // overwritten by -ldflag "-X 'github.com/ATMackay/psql-ledger/service.commitDate=$commit_date'"
	buildDate  string // overwritten by -ldflag "-X 'github.com/ATMackay/psql-ledger/service.buildDate=$build_date'"
	version    string // overwritten by -ldflag "-X 'github.com/ATMackay/psql-ledger/service.version=$version'"
)

// Version is the full Git semver tag
var Version = func() string {
	if version != "" {
		return version
	}
	return "v0.0.0-dev"
}()

// GitCommitHash https://icinga.com/blog/2022/05/25/embedding-git-commit-information-in-go-binaries/
var GitCommitHash = func() string {

	// Try embedded value
	if len(gitCommit) > 7 {
		mustDecodeHex(gitCommit[0:8]) // will panic if build has been generated with a malicious $commit_hash value
		return gitCommit[0:8]
	}
	var commit string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			}
		}
	}
	if commit == "" {
		return "00000000" // default commit string
	}
	mustDecodeHex(commit)
	return commit
}()

// BuildDate returns a build date generator
var BuildDate = func() string {
	if buildDate != "" {
		return buildDate
	}
	return time.Now().Format(time.RFC3339)
}()

// CommitDate returns a commit date generator
var CommitDate = func() string {
	if commitDate != "" {
		return commitDate
	}
	return time.Now().Format(time.RFC3339)
}()

func mustDecodeHex(input string) {
	_, err := hex.DecodeString(input)
	if err != nil {
		panic(err)
	}
}
