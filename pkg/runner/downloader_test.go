/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloader_Download(t *testing.T) {
	var files []string
	mgr := NewManager()
	mgr.StartLoggers(-1)
	files = append(files, "https://ftp.ensembl.org/pub/README")
	files = append(files, "https://raw.githubusercontent.com/oborel/obo-relations/master/ro.obo")
	files = append(files, "https://ftp.ebi.ac.uk/pub/databases/gwas/releases/latest/gwas-efo-trait-mappings.tsv")
	require.NoError(t, mgr.Download(TmpRunPath(t), files))
}

// TmpRunPath returns the tmp path, and it gets deleted right after the test ends.
func TmpRunPath(t *testing.T) string {
	testTmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(testTmpDir, 0755))
	if _, err := os.Stat(testTmpDir); os.IsNotExist(err) {
		t.Logf("test ephemeral path not found: %s", testTmpDir)
		require.NoError(t, err)
	} else {
		t.Logf("test ephemeral path: %s", testTmpDir)
	}
	return testTmpDir
}
