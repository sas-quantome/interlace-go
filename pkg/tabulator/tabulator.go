/*
 * Copyright 2026 SAS Quantome
 * SPDX-License-Identifier: Apache-2.0
 */

package tabulator

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/sas-quantome/interlace-go/pkg/runner"
)

// MatchMethod represents the type of matching strategy used when searching for table entries.
type MatchMethod int8

const (
	ExactMatch     MatchMethod = iota // code for array search method: exact string match
	MatchContains                     // code for array search method: partial string match
	PrefixMatch                       // code for array search method: prefix string match
	SuffixMatch                       // code for array search method: suffix string match
	UnMatchShowAll                    // code for array search method: no string compare is performed
)

// Tabulator represents a data structure for tabular data (column-oriented).
type Tabulator struct {
	// Data is stored as [ColumnIndex][RowIndex]string
	Data      [][]string
	Rows      int
	ColCursor int // Used for manual appending
}

// NewTabulator initializes and returns a new Tabulator instance.
func NewTabulator() *Tabulator {
	return &Tabulator{
		Data: make([][]string, 0),
	}
}

// LoadTable loads a tabular data file.
func (t *Tabulator) LoadTable(fn string) error {
	scanner := runner.NewScanner(0) // 0 = default buffer size
	if err := scanner.Start(fn); err != nil {
		return runner.ValidateErrors(err, scanner.Close())
	}

	// Reset state
	t.Data = nil
	t.Rows = 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		nCols := len(parts)

		// Initialize columns on the first row
		if t.Rows == 0 {
			t.Data = make([][]string, nCols)
		} else {
			// Validate column consistency
			if len(t.Data) != nCols {
				return runner.ValidateErrors(errors.New("number of columns in tsv file is not constant"), scanner.Close())
			}
		}

		// Append data to each column
		for i, val := range parts {
			t.Data[i] = append(t.Data[i], val)
		}
		t.Rows++
	}

	return runner.ValidateErrors(scanner.Err(), scanner.Close())
}

// Columns return the number of columns.
func (t *Tabulator) Columns() int {
	return len(t.Data)
}

// TabWriteTableSearch writes rows to output using a custom delimiter (e.g., CSV/TSV style).
func (t *Tabulator) TabWriteTableSearch(output io.Writer, columnName, keyword, delimiter string, method MatchMethod) error {
	w := bufio.NewWriter(output)
	err := t.writeFiltered(w, columnName, keyword, delimiter, method)
	if err != nil {
		return err
	}
	return w.Flush()
}

// PadWriteTableSearch writes rows to output using elastic tab-stops (aligned visualization).
func (t *Tabulator) PadWriteTableSearch(output io.Writer, columnName, keyword string, method MatchMethod) error {
	w := tabwriter.NewWriter(output, 0, 0, 1, ' ', 0)
	// Use explicit tab for the tabwriter to align correctly
	err := t.writeFiltered(w, columnName, keyword, "\t", method)
	if err != nil {
		return err
	}
	return w.Flush()
}

// writeFiltered is the core logic shared by both writing methods.
// It writes the header, selects the predicate, filters rows, and writes matches.
func (t *Tabulator) writeFiltered(w io.Writer, colName, keyword, sep string, method MatchMethod) error {
	if len(t.Data) == 0 || t.Rows == 0 {
		return nil
	}

	// 1. Write Header (Row 0)
	if err := t.writeRow(w, 0, sep); err != nil {
		return err
	}

	// 2. Determine Search Column
	queryColIdx := -1
	if method != UnMatchShowAll {
		queryColIdx = t.SearchColumn(colName)
		// If column not found or the keyword empties, we might choose to return or print nothing (matching original behavior)
		if queryColIdx < 0 || keyword == "" {
			return nil
		}
	}

	// 3. Select Predicate Function
	// Optimization: Define the check once, outside the loop.
	var matchFunc func(string) bool

	switch method {
	case ExactMatch:
		matchFunc = func(s string) bool { return s == keyword }
	case PrefixMatch:
		matchFunc = func(s string) bool { return strings.HasPrefix(s, keyword) }
	case SuffixMatch:
		matchFunc = func(s string) bool { return strings.HasSuffix(s, keyword) }
	case MatchContains:
		matchFunc = func(s string) bool { return strings.Contains(s, keyword) }
	case UnMatchShowAll:
		matchFunc = func(s string) bool { return true }
	default:
		return nil
	}

	// 4. Iterate Rows (Skip header at 0)
	// We assume all columns have the same number of rows due to LoadTable checks.
	// Accessing t.Data[queryColIdx][r] is safe if the queryColIdx is valid.

	for r := 1; r < t.Rows; r++ {
		// If we are showing all, or if the cell matches the predicate
		if method == UnMatchShowAll || matchFunc(t.Data[queryColIdx][r]) {
			if err := t.writeRow(w, r, sep); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeRow writes a single row across all columns joined by the separator.
func (t *Tabulator) writeRow(w io.Writer, rowIdx int, sep string) error {
	for c := 0; c < len(t.Data); c++ {
		if c > 0 {
			if _, err := io.WriteString(w, sep); err != nil {
				return err
			}
		}
		// fmt.Fprint handles various types if we ever change []string,
		// but io.WriteString is faster for known strings.
		if _, err := io.WriteString(w, t.Data[c][rowIdx]); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// SearchTable searches for a key in the KeyColumn and returns the value from ValueColumn.
// Returns an empty string if not found.
func (t *Tabulator) SearchTable(keyColName, valColName, key string) string {
	if keyColName == valColName {
		return ""
	}

	kIdx := t.SearchColumn(keyColName)
	vIdx := t.SearchColumn(valColName)

	if kIdx < 0 || vIdx < 0 {
		return ""
	}

	// Linear scan
	// Optimization: Since data is columnar, we iterate the Key Column directly.
	keyCol := t.Data[kIdx]
	valCol := t.Data[vIdx]

	// Start at 1 to skip the header
	for r := 1; r < t.Rows; r++ {
		if keyCol[r] == key {
			return valCol[r]
		}
	}
	return ""
}

// SearchColumn returns the index of the column with the specified header name.
func (t *Tabulator) SearchColumn(name string) int {
	if name == "" {
		return -1
	}
	for i, col := range t.Data {
		// Header is at row 0
		if len(col) > 0 && col[0] == name {
			return i
		}
	}
	return -1
}

// AppendColumn appends a string to the current column pointed to by ColCursor.
// This is used for manual table construction.
func (t *Tabulator) AppendColumn(s string) {
	// Ensure we have enough columns
	if t.ColCursor >= len(t.Data) {
		// Expand data slice
		// Assuming we simply append to the existing column at ColCursor.
		// If ColCursor is out of bounds, we append new columns.
		for len(t.Data) <= t.ColCursor {
			t.Data = append(t.Data, []string{})
		}
	}

	t.Data[t.ColCursor] = append(t.Data[t.ColCursor], s)
	t.ColCursor++
}

// ResetCursor resets the manual column cursor to 0 (e.g., for a new row).
func (t *Tabulator) ResetCursor() {
	t.ColCursor = 0
}
