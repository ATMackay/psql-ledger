package service

import (
	"encoding/hex"
	"runtime/debug"
	"time"
)

var (
	gitCommit string // overwritten by -ldflag "-X 'github.com/ATMackay/psql-ledger/service.gitCommit=$commit_hash'"
	buildDate string // overwritten by -ldflag "-X 'github.com/qredo/fusionchain/keyring/pkg/common.buildDate=$build_date'"
)

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

// Date returns a build date generator
var Date = func() string {
	if buildDate != "" {
		return buildDate
	}
	return time.Now().Format(time.RFC3339)
}()

func mustDecodeHex(input string) {
	_, err := hex.DecodeString(input)
	if err != nil {
		panic(err)
	}
}
