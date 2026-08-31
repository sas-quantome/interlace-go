/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStacker_Push(t *testing.T) {
	var s Stack[int]
	s.Reset()
	s.Push(1)
	s.Push(2)
	require.Equal(t, 2, s.Len())
	var a Stack[any]
	a.Reset()
	a.Push(1.001)
	a.Push("two")
	require.Equal(t, 2, a.Len())
	require.Equal(t, "[1.001 two]", a.String())
}

func TestStacker_Pop(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("one")
	s.Push("two")
	item, found := s.Pop()
	require.True(t, found)
	require.Equal(t, "two", item)
}

func TestStacker_Peek(t *testing.T) {
	var s Stack[uint8]
	s.Reset()
	s.Push(1)
	s.Push(2)
	require.Equal(t, uint8(2), s.Peek())
}

func TestStacker_String(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("any")
	s.Push("one")
	s.Push("two")
	require.Equal(t, "[any one two]", s.String())
	var u Stack[uint32]
	u.Push(1)
	u.Push(2)
	u.Push(3)
	require.Equal(t, "[1 2 3]", u.String())
	var f Stack[float64]
	f.Push(1.000001)
	f.Push(2.000100)
	f.Push(3.100000)
	require.Equal(t, "[1.000001 2.0001 3.1]", f.String())
}

func TestStacker_Clear(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("any")
	s.Clear()
	require.True(t, s.Empty())
}

func TestStacker_Len(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("any")
	s.Push("any")
	s.Push("any")
	require.Equal(t, 3, s.Len())
}

func TestStacker_Empty(t *testing.T) {
	var s Stack[string]
	require.True(t, s.Empty())
	s.Reset()
	require.True(t, s.Empty())
}

func TestStacker_Copy(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("abc")
	s.Push("def")
	s.Push("ghi")
	c := s.Copy()
	require.Equal(t, s.String(), c.String())
}

func TestStacker_Reset(t *testing.T) {
	var s Stack[string]
	require.True(t, s.Empty())
	s.Reset()
	require.True(t, s.Empty())
	s.Push("abc")
	require.False(t, s.Empty())
	s.Reset()
	require.True(t, s.Empty())
}

func TestStacker_Reverse(t *testing.T) {
	var s Stack[string]
	s.Reset()
	s.Push("abc")
	s.Push("def")
	s.Push("ghi")
	s.Reverse()
	require.Equal(t, "[ghi def abc]", s.String())
}

func TestStacker_ListOf(t *testing.T) {
	list := ListOf("one", "two", "three")
	require.Equal(t, 3, len(list))
}
