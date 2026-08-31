/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package extractor

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestExtractor_NewExtractor(t *testing.T) {
	ex := NewExtractor(1024)
	require.NotNil(t, ex)
	require.Equal(t, 1024, len(ex.Buffer))
}

func TestExtractor_Reset(t *testing.T) {
	ex := NewExtractor(1024)
	ex.index = 10
	ex.Reset()
	require.Equal(t, 0, ex.index)
	require.False(t, ex.frOpen)
}

func TestExtractor_Read(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024 * 8)
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			ex.TestShow()
			ex.IncrementIndex()
		}
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
	}
	require.NoError(t, ex.End())
}

func TestExtractor_BeginningOfLastColumn(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024 * 8)
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			b := ex.Buffer[ex.Index()]
			if b == '\t' || b == '\n' {
				colID := ex.Column() // CRITICAL: Capture column index *before* Extractor increments it
				if b == '\t' {
					ex.IndexAtEndOfColumn(colID)
				} else {
					ex.IndexAtEndOfLine(colID)
				}
				if ex.Row() == 10 && colID == 7 {
					fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), colID, ex.ColumnData())
					require.Equal(t, byte('3'), ex.Buffer[ex.BeginningOfLastColumn()])
					fmt.Println("The char at the beginning of the last column is 3.")
				}
				ex.IndexAtColumnData()
			} else {
				ex.IndexAtColumnData()
			}
			ex.IncrementIndex()
		}
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
	}
	require.NoError(t, ex.End())
}

func TestExtractor_ColumnData(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024 * 8)
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			b := ex.Buffer[ex.Index()]
			if b == '\t' || b == '\n' {
				colID := ex.Column() // CRITICAL: Capture column index *before* Extractor increments it
				if b == '\t' {
					ex.IndexAtEndOfColumn(colID)
				} else {
					ex.IndexAtEndOfLine(colID)
				}
				if ex.Row() == 10 && colID == 7 {
					fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), colID, ex.ColumnData())
					require.Equal(t, "30971692", ex.ColumnData())
				}
				ex.IndexAtColumnData()
			} else {
				ex.IndexAtColumnData()
			}
			ex.IncrementIndex()
		}
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
	}
	require.NoError(t, ex.End())
}

func TestExtractor_EndOfLastColumn(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024 * 8)
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			b := ex.Buffer[ex.Index()]
			if b == '\t' || b == '\n' {
				colID := ex.Column() // CRITICAL: Capture column index *before* Extractor increments it
				if b == '\t' {
					ex.IndexAtEndOfColumn(colID)
				} else {
					ex.IndexAtEndOfLine(colID)
				}
				if ex.Row() == 10 && colID == 7 {
					fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), colID, ex.ColumnData())
					require.Equal(t, byte('\r'), ex.Buffer[ex.EndOfLastColumn()-1])
					require.Equal(t, byte('\n'), ex.Buffer[ex.EndOfLastColumn()])
				}
				ex.IndexAtColumnData()
			} else {
				ex.IndexAtColumnData()
			}
			ex.IncrementIndex()
		}
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
	}
	require.NoError(t, ex.End())
}

func TestExtractor_CatchTruncatedColumnData(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024) // smaller than input file size 7743
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	var c uint8
	var line, buf string
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			b := ex.Buffer[ex.Index()]
			if b == '\t' || b == '\n' {
				colID := ex.Column() // CRITICAL: Capture column index *before* Extractor increments it
				if b == '\t' {
					ex.IndexAtEndOfColumn(colID)
				} else {
					ex.IndexAtEndOfLine(colID)
					line = ex.LineData()
					fmt.Println("line:", ex.Row()-1, "data:", line)
					if len(buf) > 0 {
						fmt.Printf("buf: [%s]\n", buf)
						require.True(t, strings.Contains(line, strings.TrimSpace(buf)))
					}
					buf = ""
				}
				fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), colID, ex.ColumnData())
				ex.IndexAtColumnData()
			} else {
				ex.IndexAtColumnData()
			}
			c = ex.Buffer[ex.Index()]
			ex.IncrementIndex()
		}
		ex.CatchTruncatedLineData()
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
		if ex.lastBuffPartColumnData != "" {
			buf = ex.lastBuffPartColumnData + string(ex.Buffer[:1])
			require.Equal(t, c, ex.lastBuffPartColumnData[len(ex.lastBuffPartColumnData)-1])
			fmt.Printf("[%s]+[%s]\n", ex.lastBuffPartColumnData, string(ex.Buffer[:32]))
		}
	}
	require.NoError(t, ex.End())
}

func TestExtractor_CatchTruncatedLineData(t *testing.T) {
	tsvInpFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	ex := NewExtractor(1024) // smaller than input file size 7743
	require.NoError(t, ex.Start(tsvInpFile))
	bytesRead := ex.Read()
	var c uint8
	for bytesRead > 0 {
		for ex.Index() < bytesRead {
			ex.TestShow()
			c = ex.Buffer[ex.Index()]
			ex.IncrementIndex()
		}
		ex.CatchTruncatedLineData()
		bytesRead = ex.ReadNext()
		if ex.lastBuffPartLineData != "" {
			require.Equal(t, c, ex.lastBuffPartLineData[len(ex.lastBuffPartLineData)-1])
			fmt.Printf("[%s]+[%s]\n", ex.lastBuffPartLineData, string(ex.Buffer[:32]))
		}
	}
	require.NoError(t, ex.End())
}
