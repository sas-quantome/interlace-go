/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCIndexer_NewCIndexer(t *testing.T) {
	cid := NewCIndexer()
	require.Equal(t, uint8('A'), cid.a)
	require.Equal(t, uint8('A'), cid.b)
	require.Equal(t, uint8('A'), cid.c)
}

func TestCIndexer_String(t *testing.T) {
	cid := NewCIndexer()
	require.Equal(t, "AAA", cid.String())
}

func TestCIndexer_LowerCase(t *testing.T) {
	cid := NewCIndexer()
	if cid.LowerCase() != "aaa" {
		t.Errorf("Expected aaa, got %s", cid.LowerCase())
	}
}

func TestCIndexer_Set(t *testing.T) {
	cid := NewCIndexer()
	require.NoError(t, cid.Set('A', 'B', 'C'))
	require.Equal(t, "ABC", cid.String())
}

func TestCIndexer_Next(t *testing.T) {
	cid := NewCIndexer()
	for i := range 26 * 26 * 26 {
		//fmt.Println(i+1, cid.String())
		cid.Next()
		i++
	}
	require.Equal(t, "ZZZ", cid.String())
}

func TestCIndexer_Back(t *testing.T) {
	cid := NewCIndexer()
	require.NoError(t, cid.Set('Z', 'Z', 'Z'))
	for i := range 26 * 26 * 26 {
		//fmt.Println(i+1, cid.String())
		cid.Back()
		i++
	}
	require.Equal(t, "AAA", cid.String())
}

func TestLIndexer_NewLIndexer(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'A')
	require.Equal(t, "AAA", lid.cid)
	require.Equal(t, uint32(1234567890), lid.pos)
	require.Equal(t, uint64(123456789012345), lid.uid)
	require.Equal(t, uint8('A'), lid.cat)
}

func TestLIndexer_Reset(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'A')
	lid.Reset("XYZ", 123, 123456789, 'X')
	require.NotZero(t, lid.cat)
	require.Equal(t, "XYZ", lid.cid)
	require.Equal(t, uint32(123), lid.pos)
	require.Equal(t, uint64(123456789), lid.uid)
	require.Equal(t, uint8('X'), lid.cat)
	lid.Reset("123", 123, 123456789, 'A')
	require.Zero(t, lid.cat) // invalid CID
	lid.Reset("XYZ", 123, 123456789, '{')
	require.Zero(t, lid.cat) // invalid cat
}

func TestLIndexer_Update(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'A')
	lid.Update(123, 123456789)
	require.Equal(t, "AAA", lid.cid)
	require.Equal(t, uint32(123), lid.pos)
	require.Equal(t, uint64(123456789), lid.uid)
	require.Equal(t, uint8('A'), lid.cat)
}

func TestLIndexer_formatID(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'x')
	s := fmt.Sprintf("%s%010d%c%020d", lid.cid, lid.pos, lid.cat, lid.uid)
	lid.formatID()
	require.Equal(t, s, string(lid.buf[:]))
	fmt.Println(s)
	fmt.Println(lid.String())
}

func TestLIndexer_String(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'y')
	s := fmt.Sprintf("%s%010d%c%020d", lid.cid, lid.pos, lid.cat, lid.uid)
	lid.formatID()
	require.Equal(t, s, lid.String())
	fmt.Println(s)
	fmt.Println(lid.String())
}

func TestLIndexer_Factor(t *testing.T) {
	lid := NewLIndexer("AAA", 1234567890, 123456789012345, 'A')
	cid, pos, cat, uid := lid.Factor()
	require.Equal(t, "AAA", cid)
	require.Equal(t, uint32(1234567890), pos)
	require.Equal(t, uint64(123456789012345), uid)
	require.Equal(t, uint8('A'), cat)
}

func TestValidCID(t *testing.T) {
	require.False(t, ValidCID("AG"))
	require.False(t, ValidCID("AGr"))
	require.True(t, ValidCID("AGR"))
	require.False(t, ValidCID("Z0Z"))
	require.False(t, ValidCID("SOSS"))
}

func TestValidCat(t *testing.T) {
	require.True(t, ValidCat('A'))
	require.True(t, ValidCat('M'))
	require.True(t, ValidCat('Z'))
	require.True(t, ValidCat('a'))
	require.True(t, ValidCat('n'))
	require.True(t, ValidCat('Z'))
	require.False(t, ValidCat('0'))
	require.False(t, ValidCat('['))
	require.False(t, ValidCat('{'))
}

func TestGIndexer_NewGIndexer(t *testing.T) {
	gid := NewGIndexer(3)
	require.Equal(t, float64(3), gid.mapSize)
}

func TestGIndexer_Set(t *testing.T) {
	var cidIndex int

	geo := make([]string, 3)
	gid := NewGIndexer(3)
	cid := NewCIndexer()

	geo[0] = cid.String()
	cid.Next()
	geo[1] = cid.String()
	cid.Next()
	geo[2] = cid.String()

	cidIndex = 0
	require.NoError(t, gid.Set(geo[cidIndex], 99, float64(cidIndex)))

	cidIndex = 1
	require.NoError(t, gid.Set(geo[cidIndex], 75, float64(cidIndex)))

	cidIndex = 2
	require.NoError(t, gid.Set(geo[cidIndex], 50, float64(cidIndex)))
}

func TestGIndexer_String(t *testing.T) {
	var cidIndex int64

	geo := make([]string, 3)
	gid := NewGIndexer(3)
	cid := NewCIndexer()

	geo[0] = cid.String()
	cid.Next()
	geo[1] = cid.String()
	cid.Next()
	geo[2] = cid.String()

	// Ordering from largest (equator) to smallest (pole)

	fmt.Println("cidIndex = 0")
	cidIndex = 0
	a1 := []rune{'A', 'C', 'C', 'N', 'A', 'A', 'G', 'A', 'C', 'C', 'N', 'A', 'A', 'G'}
	z1 := len(a1)
	require.NoError(t, gid.Set(geo[cidIndex], float64(z1), float64(cidIndex)))
	for i := range z1 {
		s := gid.String(float64(i), float64(a1[i]))
		v := fmt.Sprintf("%g, %g", gid.Latitude(float64(a1[i])), gid.Longitude(float64(i)))
		fmt.Printf("%s %s\n", geo[cidIndex], s)
		require.Equal(t, s, v)
	}

	fmt.Println("cidIndex = 1")
	cidIndex = 1
	a2 := []rune{'A', 'C', 'C', 'N', 'A', 'A', 'G'}
	z2 := len(a2)
	require.NoError(t, gid.Set(geo[cidIndex], float64(z2), float64(cidIndex)))
	for i := range z2 {
		fmt.Printf("%s %s\n", geo[cidIndex], gid.String(float64(i), float64(a2[i])))
	}

	fmt.Println("cidIndex = 2")
	cidIndex = 2
	a3 := []rune{'A', 'C', 'G', 'T'}
	z3 := len(a3)
	require.NoError(t, gid.Set(geo[cidIndex], float64(z3), float64(cidIndex)))
	for i := range z3 {
		fmt.Printf("%s %s\n", geo[cidIndex], gid.String(float64(i), float64(a3[i])))
	}
}
