/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package tester

import (
	"os"
	"path"
	"runtime"
)

// GitRepoPath locates the absolute path of a Git repository given its project and repository name.
func GitRepoPath() string {
	_, filename, _, _ := runtime.Caller(0)
	// Try finding it by traversing up from tester.go location (tester.go is in pkg/tester/tester.go)
	// which is 3 levels below repo root.
	repoRoot := path.Dir(path.Dir(path.Dir(filename)))
	if _, err := os.Stat(path.Join(repoRoot, "go.mod")); err == nil {
		return repoRoot
	}
	return ""
}

func TestBasePath() string {
	return path.Join(GitRepoPath(), "test", "base")
}

func TestDataPath() string {
	return path.Join(GitRepoPath(), "test", "data")
}

func TestPath() string {
	return path.Join(GitRepoPath(), "test")
}

func TestTimeTags() string {
	return path.Join(GitRepoPath(), "test", "base", "entity", "btc", "btc_usd")
}
