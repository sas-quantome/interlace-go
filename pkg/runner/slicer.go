/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
)

// FilterItems applies a predicate function to a slice and returns a new slice with items that satisfy the predicate.
// Note: For in-place filtering (zero allocation), consider slices.DeleteFunc.
func FilterItems[T any](items []T, f func(T) bool) []T {
	var res []T
	for _, item := range items {
		if f(item) {
			res = append(res, item)
		}
	}
	return res
}

// FindItem searches for the first occurrence of the target in the item slice and returns its index or -1 if not found.
func FindItem[T comparable](items []T, target T) int {
	return slices.Index(items, target) // Replaced manual loop with standard lib
}

// BinarySearchItem performs a binary search on a sorted slice and returns the index of the target if found, or -1 otherwise.
func BinarySearchItem[T cmp.Ordered](items []T, target T) int {
	if i, found := slices.BinarySearch(items, target); found {
		return i
	}
	return -1
}

// InsertOnce inserts an item into a sorted slice only if it does not already exist.
// Returns the slice (possibly reallocated).
func InsertOnce[T cmp.Ordered](ts []T, t T) []T {
	i, found := slices.BinarySearch(ts, t)
	if found {
		return ts
	}
	return slices.Insert(ts, i, t)
}

// Insert inserts an item into a sorted slice, allowing duplicates.
// Returns the slice (possibly reallocated).
func Insert[T cmp.Ordered](ts []T, t T) []T {
	i, _ := slices.BinarySearch(ts, t)
	return slices.Insert(ts, i, t)
}

// SortItems sorts a slice of items in ascending order based on their natural ordering.
func SortItems[T cmp.Ordered](items []T) {
	slices.Sort(items)
}

// WriteSlice writes all elements of a slice to a file, one per line.
// Uses generics (no reflection) and buffered I/O.
func WriteSlice[T any](fn string, items []T) (err error) {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}

	// Ensure a file is closed and errors are captured properly
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("write error: %v, close error: %v", err, closeErr)
			}
		}
	}()

	// Use buffered writer for performance
	w := bufio.NewWriter(f)
	for _, item := range items {
		// fmt.Fprintln works with any type
		if _, err = fmt.Fprintln(w, item); err != nil {
			return err
		}
	}

	return w.Flush()
}

// FileLinesToSlice reads a file line by line and returns a slice of non-empty lines.
func FileLinesToSlice(fp string, sortIt bool, omitLinePrefix string) ([]string, error) {
	var lines []string
	ska := NewScanner(0)
	if err := ska.Start(fp); err != nil {
		return nil, ValidateErrors(err, ska.Close())
	}
	for ska.Scan() {
		line := ska.Text()
		if len(line) > 0 {
			if omitLinePrefix != "" && strings.HasPrefix(line, omitLinePrefix) {
				continue
			}
			lines = append(lines, line)
		}
	}
	if sortIt {
		slices.Sort(lines)
	}
	return lines, ska.Close()
}

// FillCols is a zero-allocation helper that fills a pre-allocated slice with tab-separated values.
// Returns the number of columns successfully extracted.
func FillCols(line string, dst []string) int {
	start := 0
	idx := 0
	length := len(line)
	maxCols := len(dst)

	for i := 0; i < length && idx < maxCols; i++ {
		if line[i] == '\t' {
			dst[idx] = line[start:i]
			idx++
			start = i + 1
		}
	}
	// Add the final column if there's room
	if idx < maxCols {
		dst[idx] = line[start:]
		idx++
	}
	return idx
}
