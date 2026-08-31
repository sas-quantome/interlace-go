/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"fmt"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/randommer"
	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestFormatter_WriteGobDataToFile(t *testing.T) {
	require.Contains(t, createTestGob(t), "sdk_emails_s.gob.gz")
}

func TestFormatter_OverWriteGobFile(t *testing.T) {
	require.NoError(t, OverWriteGobFile(createTestGob(t), testSortedEmails()))
}

func TestFormatter_LoadDataFromGobFile(t *testing.T) {
	var emails []string
	require.NoError(t, LoadDataFromGobFile(createTestGob(t), &emails))
	for i, e := range emails {
		fmt.Println(i, e)
	}
}

func TestFormatter_ReplaceAtIndex(t *testing.T) {
	s := OutPath
	require.Equal(t, "OUT_XATH", ReplaceAtIndex(s, 'X', 4))
}

func TestFormatter_StringToGzipBytes(t *testing.T) {
	s := "This is a test string."
	gz, err := StringToGzipBytes(s)
	require.NoError(t, err)
	z, e := GzipBytesToString(gz)
	require.NoError(t, e)
	require.Equal(t, s, z)
}

func TestFormatter_CopyFile(t *testing.T) {
	fp := tester.TestPath()
	src := fp + "/README.md"
	dst := fp + "/READNE_COPY.md"
	require.NoError(t, CopyFile(src, dst))
	require.NoError(t, os.Remove(dst))
}

func testSortedEmails() []string {
	emails := make([]string, 100)
	for i := range 100 {
		emails[i] = randommer.RandomEmail()
	}
	slices.Sort(emails)
	return emails
}

func createTestGob(t *testing.T) string {
	return testWriteGob(t, "sdk_emails_s.gob.gz", testSortedEmails())
}

func testWriteGob(t *testing.T, name string, data any) string {
	fp := TmpRunPath(t)
	fn := path.Join(fp, name)
	require.NoError(t, WriteGobDataToFile(fn, data))
	return fn
}

func TestExtractNumber(t *testing.T) {
	require.Equal(t, "123", ExtractNumber("GO:000123", "0", "GO:"))
}

func TestPopOutString(t *testing.T) {
	require.Equal(t, "abcdef", PopOutString("sdk_abcdef.gob.gz", "sdk_", ".gob.gz"))
}

func TestSplitAccessionVersion(t *testing.T) {
	a, v := SplitAccessionVersion("AC012345.1")
	require.Equal(t, "AC012345", a)
	require.Equal(t, "1", v)
}
