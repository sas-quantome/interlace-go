/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"os"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/randommer"
	"github.com/stretchr/testify/require"
)

func TestCataloguer_NewCataloguer(t *testing.T) {
	goCat := NewCataloguer()
	require.NotNil(t, goCat)
	require.Equal(t, 0, len(goCat.items))
}

func TestCataloguer_HashText(t *testing.T) {
	goCat := NewCataloguer()
	hash1 := goCat.HashText("test")
	hash2 := goCat.HashText("test")
	require.Equal(t, hash1, hash2)
	require.NotZero(t, hash1)
}

func TestCataloguer_SetBinPathAndFileNames(t *testing.T) {
	goCat := NewCataloguer()
	catFp := TmpRunPath(t)
	goCat.SetBinPathAndFileNames(catFp, "nameCat")
	require.Equal(t, goCat.binPath, catFp)
}

func TestCataloguer_CreateTmpFile(t *testing.T) {
	goCat := NewCataloguer()
	catFp := TmpRunPath(t)
	goCat.SetBinPathAndFileNames(catFp, "fileCat")
	require.NoError(t, goCat.CreateTmpFile(catFp))
}

func TestCataloguer_WriteTuple(t *testing.T) {
	goCat := NewCataloguer()
	catFp := TmpRunPath(t)
	goCat.SetBinPathAndFileNames(catFp, "saveCat")
	require.NoError(t, goCat.CreateTmpFile(catFp))
	N := 5
	hid := make([]uint64, N)
	txt := make([]string, N)
	for i := range N {
		txt[i] = randommer.RandomEmail()
		hid[i] = goCat.HashText(txt[i])
		require.NoError(t, goCat.WriteTuple(hid[i], txt[i]))
	}
	require.NoError(t, goCat.Save())
	require.NoError(t, goCat.RemoveTmpFile())
	require.True(t, goCat.newRecords == N)
}

// TmpRunPath returns the tmp path, and it gets deleted right after the test ends.
func TmpRunPath(t *testing.T) string {
	testTmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(testTmpDir, 0755))
	if _, err := os.Stat(testTmpDir); os.IsNotExist(err) {
		require.NoError(t, err)
	} else {
		t.Logf("test ephemeral path: %s", testTmpDir)
	}
	return testTmpDir
}
