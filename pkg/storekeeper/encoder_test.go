/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder_NewEncoder(t *testing.T) {
	e := NewEncoder(TrinucleotideEnum())
	require.NoError(t, e.Err())
}

func TestEncoder_Number(t *testing.T) {
	e := NewEncoder(TrinucleotideEnum())
	require.NoError(t, e.Err())
	for i, s := range TrinucleotideEnum() {
		require.Equal(t, i, e.Number(s))
	}
}

func TestEncoder_String(t *testing.T) {
	e := NewEncoder(TrinucleotideEnum())
	require.NoError(t, e.Err())
	for i, s := range TrinucleotideEnum() {
		require.Equal(t, s, e.String(i))
	}
}

func TestEncoder_NewNucleotideEncoder(t *testing.T) {
	na := NewNucleotideEncoder()
	na.Reset('A')
	require.Equal(t, 0, na.Next(-1, 'A'))
	na.Reset('N')
	require.Equal(t, 0, na.Next(-1, 'N'))
}

func TestEncoder_Clear(t *testing.T) {
	na := NewNucleotideEncoder()
	na.key = EncNA
	na.NSpan = 100
	na.index = 42
	na.Clear(false)
	require.Equal(t, byte(0), na.key)
	require.Equal(t, uint32(0), na.NSpan)
	require.Equal(t, -1, na.index)
}

func TestEncoder_Reset(t *testing.T) {
	na := NewNucleotideEncoder()
	na.key = EncUI
	na.NSpan = 99
	na.index = 66
	na.Reset('A')
	require.Equal(t, EncNA, na.key)
	require.Equal(t, uint32(0), na.NSpan)
	require.Equal(t, -1, na.index)
	require.True(t, na.bitMask.Has(EncIsNA))
	require.False(t, na.bitMask.Has(EncIsUI))
	na.Reset('B')
	require.Equal(t, EncUI, na.key)
	require.Equal(t, uint32(0), na.NSpan)
	require.Equal(t, -1, na.index)
	require.True(t, na.bitMask.Has(EncIsUI))
	require.False(t, na.bitMask.Has(EncIsNA))
}

func TestEncoder_EncodeNucleotide(t *testing.T) {
	var buf strings.Builder
	na := NewNucleotideEncoder()
	sq := "ACCGTACTGAAGCTA"
	na.Reset(sq[0])
	for i := range sq {
		na.Encode(sq[i])
		if na.NotIndexed() {
			buf.WriteString(na.Decode(na.BitMask()))
			na.Reset(sq[i])
			na.Encode(sq[i])
		}
	}
	buf.WriteString(na.Decode(na.BitMask()))
	require.Equal(t, sq, buf.String())
}

func TestEncoder_Decode(t *testing.T) {
	var buf strings.Builder
	na := NewNucleotideEncoder()
	sq := "ACCGTACTGAAGCTANNNNNNNNNNBAGAVACGTTGACCACTGTGATCACGTGCNGACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGT"
	na.Reset(sq[0])
	for i := range sq {
		if sq[i] == 'N' {
			if na.NSpan == 0 {
				buf.WriteString(na.Decode(na.BitMask()))
				fmt.Println("1", buf.String(), na.Decode(na.BitMask()))
				na.Clear(false)
			}
			na.NSpan++
			na.UpdateMask(na.NSpan)
			continue
		} else if na.NSpan == 1 {
			na.Reset('N')
			na.Encode('N')
		} else if na.NSpan > 1 {
			buf.WriteString(na.Decode(na.BitMask()))
			fmt.Println("2", buf.String(), na.Decode(na.BitMask()))
			na.Reset(sq[i])
		}
		na.Encode(sq[i])
		if na.NotIndexed() {
			buf.WriteString(na.Decode(na.BitMask()))
			fmt.Println("3", buf.String(), na.Decode(na.BitMask()))
			na.Reset(sq[i])
			na.Encode(sq[i])
		}
	}
	buf.WriteString(na.Decode(na.BitMask()))
	fmt.Println("4", buf.String(), na.Decode(na.BitMask()))
	require.Equal(t, sq, buf.String())
	require.Equal(t, len(sq), na.NBases())
}
