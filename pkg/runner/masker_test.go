/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMasker_Has(t *testing.T) {
	var x BitMask[uint16]
	x.Add(4)
	require.True(t, x.Has(4))
}

func TestMasker_Add(t *testing.T) {
	var m SmBitMask
	m.Add(1)
	require.True(t, m.Has(1))
}

func TestMasker_Clear(t *testing.T) {
	var m XlBitMask
	m.Add(18475298)
	require.True(t, m.Has(18475298))
	m.Clear(18475298)
	require.False(t, m.Has(18475298))
}

func TestMasker_ClearAll(t *testing.T) {
	var m XlBitMask
	m.Add(123)
	m.Add(1)
	require.True(t, m.Has(1))
	require.True(t, m.Has(123))
	m.ClearAll()
	require.False(t, m.Has(1))
	require.False(t, m.Has(123))
}

func TestMasker_Toggle(t *testing.T) {
	var x BitMask[uint]
	x.Add(1234)
	require.True(t, x.Has(1234))
	x.Toggle(1234)
	require.False(t, x.Has(1234))
}
