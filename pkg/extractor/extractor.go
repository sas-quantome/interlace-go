/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package extractor

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extractor is a struct that provides methods and attributes for reading, buffering, and processing file data efficiently.
type Extractor struct {
	Buffer                                                                              []byte
	fr                                                                                  *os.File
	br                                                                                  *bufio.Reader
	gr                                                                                  *gzip.Reader
	index, column, beginningOfLastLine, beginningOfLastColumn, endOfLastColumn, bufSize int
	lastLineStarts, lastLineEnds, bytesBetweenEndDelimiter, row, bytesRead, bytesTotal  int
	IsComment, hasPartialColumnData, hasPartialLineData, startCheckPoint, endCheckPoint bool
	frOpen, grOpen                                                                      bool
	lastBuffPartColumnData, lastBuffPartLineData, columnData                            string
}

// NewExtractor initializes and returns a new Extractor with the given buffer size.
func NewExtractor(bufSize int) *Extractor {
	extractor := new(Extractor)
	extractor.Reset()
	extractor.bufSize = bufSize
	extractor.Buffer = make([]byte, bufSize)
	return extractor
}

// Reset resets all internal state variables of the Extractor to their initial values, clearing any stored data or flags.
func (ex *Extractor) Reset() {
	ex.lastBuffPartColumnData = ""
	ex.lastBuffPartLineData = ""
	ex.columnData = ""
	ex.index, ex.row, ex.column, ex.lastLineStarts, ex.lastLineEnds = 0, 0, 0, 0, 0
	ex.beginningOfLastLine, ex.beginningOfLastColumn, ex.endOfLastColumn = 0, 0, 0
	ex.bytesBetweenEndDelimiter, ex.bytesRead, ex.bytesTotal = 0, 0, 0
	ex.frOpen, ex.grOpen, ex.IsComment, ex.hasPartialColumnData = false, false, false, false
	ex.hasPartialLineData, ex.startCheckPoint, ex.endCheckPoint = false, false, false
}

// ResetOffsets resets all extract-related offsets and flags to their initial state.
func (ex *Extractor) ResetOffsets() {
	ex.index, ex.beginningOfLastLine, ex.beginningOfLastColumn = 0, 0, 0
	ex.endOfLastColumn, ex.bytesBetweenEndDelimiter = 0, 0
	ex.startCheckPoint, ex.endCheckPoint = false, false
}

// Read reads data into the buffer from the underlying reader and returns the number of bytes read.
func (ex *Extractor) Read() int {
	//ex.bytesRead, _ = io.ReadAtLeast(ex.br, ex.Buffer, 1)
	ex.bytesRead, _ = io.ReadFull(ex.br, ex.Buffer)
	ex.bytesTotal += ex.bytesRead
	return ex.bytesRead
}

// ReadNext resets internal offsets and reads the next integer value from the source.
func (ex *Extractor) ReadNext() int {
	ex.ResetOffsets()
	return ex.Read()
}

// Start opens the specified file or stdin, initializes the appropriate reader, and prepares the extractor for processing.
func (ex *Extractor) Start(fn string) error {
	var err error
	if fn == "-" {
		ex.fr = os.Stdin
	} else {
		ex.fr, err = os.Open(fn)
	}
	if err == nil {
		ex.frOpen = true
		if filepath.Ext(fn) == ".gz" || filepath.Ext(fn) == ".bgz" {
			ex.gr, err = gzip.NewReader(ex.fr)
			if err == nil {
				ex.grOpen = true
				ex.br = bufio.NewReader(ex.gr)
			}
		} else if filepath.Ext(fn) == ".bz2" {
			ex.br = bufio.NewReader(bzip2.NewReader(bufio.NewReader(ex.fr)))
		} else {
			ex.br = bufio.NewReader(ex.fr)
		}
	}
	return err
}

// End closes any open resources associated with the Extractor and returns an error if any close operation fails.
func (ex *Extractor) End() error {
	str := ""
	if ex.grOpen {
		if err := ex.gr.Close(); err != nil {
			str += fmt.Sprintf("gzip reader close: %s: ", err.Error())
		}
	}
	if ex.frOpen && ex.fr != os.Stdin {
		if err := ex.fr.Close(); err != nil {
			str += fmt.Sprintf("file close: %s: ", err.Error())
		}
	}
	if str != "" {
		return errors.New(strings.TrimRight(str, ": "))
	}
	return nil
}

// IndexAtEndOfLine updates internal indexes and state to reflect the end position of the specified column in a line.
func (ex *Extractor) IndexAtEndOfLine(column int) {
	ex.IndexAtEndOfColumn(column)
	if ex.row == 0 {
		ex.lastLineStarts = 0 // test OK
	} else {
		ex.lastLineStarts = ex.beginningOfLastLine - 1 // test OK
	}
	ex.lastLineEnds = ex.index
	ex.beginningOfLastLine = 0
	ex.endOfLastColumn = ex.index
	ex.column = 0
	ex.endCheckPoint = true
	ex.row++
	ex.bytesBetweenEndDelimiter = 0
}

// IndexAtEndOfColumn processes the end of a column in the data buffer and updates relevant state variables.
func (ex *Extractor) IndexAtEndOfColumn(column int) {
	ex.columnData = ""
	ex.endOfLastColumn = ex.index
	if ex.column == 0 {
		ex.startCheckPoint = true
		if ex.Buffer[ex.beginningOfLastColumn] == '#' {
			ex.IsComment = true
		} else {
			ex.IsComment = false
		}
	}
	if ex.beginningOfLastLine == 0 {
		ex.beginningOfLastLine = ex.index // This is the beginning of the last line read (is the current).
	}
	// No exception
	if ex.column == column {
		if ex.hasPartialColumnData {
			ex.columnData = ex.lastBuffPartColumnData +
				string(ex.Buffer[ex.beginningOfLastColumn:ex.endOfLastColumn])
		} else {
			ex.columnData = string(ex.Buffer[ex.beginningOfLastColumn:ex.endOfLastColumn])
		}
		ex.columnData = strings.TrimSpace(ex.columnData)
	}
	if ex.hasPartialColumnData {
		ex.hasPartialColumnData = false
		ex.lastBuffPartColumnData = ""
	}
	ex.column++
}

// IndexAtEndOfColumnData processes the end of a column's data by updating tracked indices and checking for special conditions.
func (ex *Extractor) IndexAtEndOfColumnData(column int) {
	ex.columnData = ""
	ex.endOfLastColumn = ex.index
	if ex.column == 0 {
		ex.startCheckPoint = true
		if ex.Buffer[ex.beginningOfLastColumn] == '#' {
			ex.IsComment = true
		} else {
			ex.IsComment = false
		}
	}
	if ex.beginningOfLastLine == 0 {
		ex.beginningOfLastLine = ex.index // This is the beginning of the last line read (is the current).
	}
	// Exception
	if ex.column == column && ex.bytesBetweenEndDelimiter > 0 {
		if ex.hasPartialColumnData {
			ex.columnData = ex.lastBuffPartColumnData +
				string(ex.Buffer[ex.beginningOfLastColumn:ex.endOfLastColumn])
		} else {
			ex.columnData = string(ex.Buffer[ex.beginningOfLastColumn:ex.endOfLastColumn])
		}
		ex.columnData = strings.TrimSpace(ex.columnData)
	}
	if ex.hasPartialColumnData {
		ex.hasPartialColumnData = false
		ex.lastBuffPartColumnData = ""
	}
	ex.column++
}

// IndexAtColumnData processes and tracks column-related data within the Extractor by adjusting internal state variables.
func (ex *Extractor) IndexAtColumnData() {
	if ex.beginningOfLastLine == 0 {
		ex.beginningOfLastLine = ex.index // This is the beginning of the last line read (is the current).
	}
	if ex.endOfLastColumn >= 0 {
		ex.columnData = ""
		ex.endOfLastColumn = -1
		ex.beginningOfLastColumn = ex.index // This is the beginning of the column space.
	}
	ex.bytesBetweenEndDelimiter++
}

// CatchTruncatedColumnData identifies and captures any truncated column data within the buffer for further processing.
func (ex *Extractor) CatchTruncatedColumnData() {
	ex.lastBuffPartColumnData = ""
	ex.hasPartialColumnData = false
	if ex.endOfLastColumn == -1 && !ex.IsComment {
		ex.lastBuffPartColumnData = string(ex.Buffer[ex.beginningOfLastColumn:ex.index])
		ex.hasPartialColumnData = true
	}
}

// CatchTruncatedLineData extracts and stores partial line data from the buffer if the line is truncated.
func (ex *Extractor) CatchTruncatedLineData() {
	ex.lastBuffPartLineData = ""
	ex.hasPartialLineData = false
	if ex.beginningOfLastLine != 0 && !ex.IsComment {
		ex.lastBuffPartLineData = string(ex.Buffer[ex.beginningOfLastLine:ex.index])
		ex.hasPartialLineData = true
	}
}

// LineData retrieves and returns the fully assembled line data from the buffer as a trimmed string.
func (ex *Extractor) LineData() string {
	if ex.hasPartialLineData {
		s := strings.TrimSpace(ex.lastBuffPartLineData + string(ex.Buffer[0:ex.lastLineEnds]))
		ex.lastBuffPartLineData = ""
		ex.hasPartialLineData = false
		return s
	}
	return strings.TrimSpace(string(ex.Buffer[ex.lastLineStarts:ex.lastLineEnds]))
}

// CheckPoint validates the buffer checkpoints and returns an error if conditions are unmet to parse the file.
func (ex *Extractor) CheckPoint() error {
	// fmt.Printf("%v\t%v\t%d\n", ex.startCheckPoint, ex.endCheckPoint, len(ex.columnData))
	if ex.startCheckPoint == false || ex.endCheckPoint == false {
		if len(ex.columnData) > 0 {
			return errors.New("buffer checkpoint failed: increase input buffer size to parse this file")
		}
	}
	return nil
}

// DeleteTruncatedLineData clears any stored partial line data and resets the corresponding state flag to false.
func (ex *Extractor) DeleteTruncatedLineData() {
	if ex.hasPartialLineData {
		ex.lastBuffPartLineData = ""
		ex.hasPartialLineData = false
	}
}

// IncrementIndex increases the internal index of the Extractor by 1.
func (ex *Extractor) IncrementIndex() {
	ex.index++
}

// MoveIndex adjusts the internal index of the Extractor by adding the given value n.
func (ex *Extractor) MoveIndex(n int) {
	ex.index += n
}

// ColumnData returns the data stored in the columnData field of the Extractor instance.
func (ex *Extractor) ColumnData() string {
	return ex.columnData
}

// ClearColumnData resets the columnData field to an empty string, clearing any previously stored data.
func (ex *Extractor) ClearColumnData() {
	ex.columnData = ""
}

// BeginningOfLastLine returns the starting position of the last line in the associated content.
func (ex *Extractor) BeginningOfLastLine() int {
	return ex.beginningOfLastLine
}

// BeginningOfLastColumn returns the index of the first character of the last column in the structure.
func (ex *Extractor) BeginningOfLastColumn() int {
	return ex.beginningOfLastColumn
}

// EndOfLastColumn returns the ending offset of the last column in the extractor.
func (ex *Extractor) EndOfLastColumn() int {
	return ex.endOfLastColumn
}

// BufSize returns the current size of the buffer used by the Extractor instance.
func (ex *Extractor) BufSize() int {
	return ex.bufSize
}

// Row returns the current row index maintained by the Extractor instance.
func (ex *Extractor) Row() int {
	return ex.row
}

// Column returns the current column number in the Extractor.
func (ex *Extractor) Column() int {
	return ex.column
}

// Index returns the current index value of the Extractor instance.
func (ex *Extractor) Index() int {
	return ex.index
}

// ParsedBytes extracts and returns a slice of bytes from the Buffer field within the specified range [x, y).
func (ex *Extractor) ParsedBytes(x, y int) []byte {
	return ex.Buffer[x:y]
}

// ParsedString extracts a substring from the Buffer field within the specified range [x:y] and returns it as a string.
func (ex *Extractor) ParsedString(x, y int) string {
	return string(ex.Buffer[x:y])
}

// TestShow processes the current buffer and prints row, column, and data information based on specific character cases.
func (ex *Extractor) TestShow() {
	switch ex.Buffer[ex.Index()] {
	case '\t':
		ex.IndexAtEndOfColumn(ex.Column())
		fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), ex.Column(), ex.ColumnData())
	case '\n':
		ex.IndexAtEndOfColumn(ex.Column())
		fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", ex.Row(), ex.Column(), ex.ColumnData())
		ex.IndexAtEndOfLine(ex.Column())
	default:
		ex.IndexAtColumnData()
	}
}
