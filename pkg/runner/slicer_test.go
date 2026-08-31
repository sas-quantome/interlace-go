/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlicer_InsertOnce(t *testing.T) {
	var s = []string{"a", "d", "c"}
	require.Equal(t, 3, len(s))
	s = InsertOnce(s, "a")
	require.Equal(t, 3, len(s))
	s = InsertOnce(s, "e")
	require.Equal(t, 4, len(s))
}

func TestSlicer_Insert(t *testing.T) {
	var s = []string{"a", "d", "c"}
	require.Equal(t, 3, len(s))
	s = Insert(s, "a")
	require.Equal(t, 4, len(s))
}

func TestSlicer_SortItems(t *testing.T) {
	var s Stack[float64]
	s.Reset()
	s.Push(1000)
	s.Push(0.2)
	s.Push(0.003009)
	s.Push(2e3)
	SortItems(s.Items)
	require.Equal(t, "[0.003009 0.2 1000 2000]", s.String())
}

func TestSlicer_FilterItems(t *testing.T) {
	var s Stack[float64]
	s.Reset()
	s.Push(1000)
	s.Push(0.2)
	s.Push(0.003009)
	s.Push(2e3)
	r := FilterItems(s.Items, func(item float64) bool { return item > 100 })
	require.Equal(t, "[1000 2000]", fmt.Sprintf("%v", r))
}

func TestSlicer_FindItem(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("Testing!")
	s.Push("ABC")
	s.Push("End of test.")
	require.Equal(t, 1, FindItem(s.Items, "ABC"))
	require.Equal(t, -1, FindItem(s.Items, ""))
}

func TestSlicer_BinarySearchItem(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("Testing!")
	s.Push("ABC")
	s.Push("End of test.")
	SortItems(s.Items)
	require.Equal(t, 0, BinarySearchItem(s.Items, "ABC"))
	require.Equal(t, -1, BinarySearchItem(s.Items, ""))
	require.Equal(t, -1, BinarySearchItem(s.Items, "ABCD"))
}

func TestSlicer_PushOnce(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("Testing!")
	s.Push("End of test.")
	s.Push("ABC")
	require.Equal(t, 3, s.Len())
	SortItems(s.Items)
	PushOnce(s.Items, "ABC")
	require.Equal(t, 3, s.Len())
	s.Items = PushOnce(s.Items, "XYZ")
	s.Items = PushOnce(s.Items, "XYZ")
	require.Equal(t, 4, s.Len())
	require.Equal(t, 0, BinarySearchItem(s.Items, "ABC"))
	require.Equal(t, -1, BinarySearchItem(s.Items, ""))
	require.Equal(t, -1, BinarySearchItem(s.Items, "ABCD"))
	require.Equal(t, 3, BinarySearchItem(s.Items, "XYZ"))
}

func TestSlicer_WriteSlice(t *testing.T) {
	var ar []string
	ar = append(ar, "one", "two", "uno", "dos")
	require.NoError(t, WriteSlice(TmpRunPath(t)+"/s2aTest.txt", ar))
}

func TestSlicer_FillCols(t *testing.T) {
	dst := make([]string, 3)
	require.Equal(t, 3, FillCols("aa\tbb\tcc", dst))
	require.Equal(t, 3, len(dst))
	require.Equal(t, "aa", dst[0])
	require.Equal(t, "bb", dst[1])
	require.Equal(t, "cc", dst[2])

}
