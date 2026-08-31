/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package cowriter

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/runner"
	"github.com/stretchr/testify/require"
)

func TestCowriter_Open(t *testing.T) {
	cw := NewCowriter()
	require.NotNil(t, cw)
	require.NoError(t, cw.Open(TmpRunPath(t), "cowriter_*.txt", true, false))
	require.True(t, cw.IsOpen())
	require.True(t, cw.IsTemporary())
	require.False(t, cw.IsGzip())
	require.False(t, cw.IsRemoved())
	require.NoError(t, cw.Close())
	require.True(t, runner.ValidatePathExists(cw.Path()))
}

func TestCowriter_Write(t *testing.T) {
	cw := NewCowriter()
	require.NoError(t, cw.Open(TmpRunPath(t), "cowriter_*.txt", true, false))
	require.Equal(t, ".txt", path.Ext(cw.Path()))
	require.Equal(t, 0, cw.BytesWritten())
	var buf bytes.Buffer
	buf.WriteString("A buffer string.\n")
	require.NoError(t, cw.Write(buf.Bytes()))
	require.Less(t, 0, cw.BytesWritten())
	require.NoError(t, cw.Close())
}

func TestCowriter_WriteString(t *testing.T) {
	cw := NewCowriter()
	require.NoError(t, cw.Open(TmpRunPath(t), "cowriter_*.txt", true, true))
	require.True(t, cw.IsGzip())
	require.Equal(t, ".gz", path.Ext(cw.Path()))
	require.Equal(t, 0, cw.BytesWritten())
	require.NoError(t, cw.WriteString("This ia plain text file.\n"))
	require.NoError(t, cw.WriteString("This ia plain text file.\n"))
	require.NoError(t, cw.WriteString("This ia plain text file.\n"))
	require.Less(t, 0, cw.BytesWritten())
	require.NoError(t, cw.Close())

}

func TestCowriter_Sort(t *testing.T) {
	cw := NewCowriter()
	require.NoError(t, cw.Open(TmpRunPath(t), "cowriter_*.txt", true, true))
	require.True(t, cw.IsGzip())
	require.Equal(t, ".gz", path.Ext(cw.Path()))
	require.NoError(t, cw.WriteString("zaz\n"))
	require.NoError(t, cw.WriteString("101\n"))
	require.NoError(t, cw.WriteString("ABC\n"))
	require.NoError(t, cw.Close())
	require.NoError(t, cw.Sort(""))
	require.NoError(t, cw.Remove())
	require.NoError(t, cw.Open(TmpRunPath(t), "cowriter_*.txt", true, false))
	require.False(t, cw.IsGzip())
	require.NoError(t, cw.WriteString("8474\n"))
	require.NoError(t, cw.WriteString("101\n"))
	require.NoError(t, cw.WriteString("9483\n"))
	require.NoError(t, cw.Close())
	require.NoError(t, cw.Sort("-n"))
}

func TestCowriter_Remove(t *testing.T) {
	cw := NewCowriter()
	fp := TmpRunPath(t)
	require.NoError(t, cw.Open(fp, "cowriter.txt", false, false))
	require.False(t, cw.IsTemporary())
	require.False(t, cw.IsRemoved())
	require.False(t, cw.IsGzip())
	require.True(t, runner.ValidatePathExists(path.Join(fp, "cowriter.txt")))
	require.NoError(t, cw.Remove())
	require.True(t, cw.IsRemoved())
	require.Nil(t, cw.File)
	require.False(t, runner.ValidatePathExists(path.Join(fp, "cowriter.txt")))
}

func TestCowriter_Clear(t *testing.T) {
	cw := NewCowriter()
	dir := TmpRunPath(t)
	require.NoError(t, cw.Open(dir, "cowriter_*.txt", true, true))
	require.NotNil(t, cw.File)
	for _, f := range []WtrBitFlag{WtrAliasA, WtrAliasB, WtrAliasC, WtrAliasD} {
		cw.SetBitFlag(f, true)
	}
	for _, f := range []WtrBitFlag{WtrAliasA, WtrAliasB, WtrAliasC, WtrAliasD} {
		require.True(t, cw.GetBitFlag(f))
	}
	require.True(t, cw.IsOpen())
	require.True(t, cw.IsTemporary())
	require.True(t, cw.IsGzip())
	require.NoError(t, cw.Close())
	// Save the filepath before clearing
	fp := cw.Path()
	require.True(t, runner.ValidatePathExists(fp))
	// Clear the Cowriter
	cw.Clear()
	require.Nil(t, cw.File)
	require.Empty(t, cw.Path())
	for _, f := range []WtrBitFlag{WtrAliasA, WtrAliasB, WtrAliasC, WtrAliasD} {
		require.False(t, cw.GetBitFlag(f))
	}
	require.False(t, cw.IsOpen())
	require.False(t, cw.IsTemporary())
	require.False(t, cw.IsGzip())
	require.True(t, runner.ValidatePathExists(fp))
	// Only after clearing, a safe reuse of the allocated Cowriter is possible
	require.NoError(t, cw.Open(dir, "cowriter.txt.gz", false, true))
	require.NotNil(t, cw.File)
	require.True(t, cw.IsGzip())
	require.NoError(t, cw.Close())
	require.True(t, runner.ValidatePathExists(cw.Path()))
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
