/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package tester

import (
	"os"
	"path"
	"runtime"
	"strings"
)

// Adapt this for your tests; here we assume the module path is github.com/sas-quantome/go-interlace.
const (

	// GitProjectName defines the name of the Git project associated with the repository.
	GitProjectName = "sas-quantome"

	// GitRepoName specifies the name of the Git repository associated with the project.
	GitRepoName = "go-interlace"
)

// GitRepoPath locates the absolute path of a Git repository given its project and repository name.
func GitRepoPath(gitProjectName, gitRepoName string) string {
	_, filename, _, _ := runtime.Caller(0)
	// Try finding it by traversing up from tester.go location (tester.go is in pkg/tester/tester.go)
	// which is 3 levels below repo root.
	repoRoot := path.Dir(path.Dir(path.Dir(filename)))
	if _, err := os.Stat(path.Join(repoRoot, "go.mod")); err == nil {
		return repoRoot
	}
	fp := path.Join(strings.Split(filename, gitProjectName)[0], gitProjectName, gitRepoName)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return ""
	}
	return fp
}

func TestBasePath() string {
	repoPath := GitRepoPath(GitProjectName, GitRepoName)
	return path.Join(repoPath, "test", "base")
}

func TestDataPath() string {
	repoPath := GitRepoPath(GitProjectName, GitRepoName)
	return path.Join(repoPath, "test", "data")
}

func TestPath() string {
	repoPath := GitRepoPath(GitProjectName, GitRepoName)
	return path.Join(repoPath, "test")
}

func TestTimeTags() string {
	repoPath := GitRepoPath(GitProjectName, GitRepoName)
	return path.Join(repoPath, "test", "base", "entity", "btc", "btc_usd")
}
