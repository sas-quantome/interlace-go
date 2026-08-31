/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"golang.org/x/exp/constraints"
)

// BitMask represents a generic bitmask for any unsigned integer type.
// T can be uint8, uint16, uint32, uint64, or uint.
type BitMask[T constraints.Unsigned] struct {
	Flag T
}

// Has checks if the specific bit flag(s) are set.
func (b *BitMask[T]) Has(key T) bool {
	return b.Flag&key != 0
}

// Add sets the specific bit flag(s).
func (b *BitMask[T]) Add(key T) {
	b.Flag |= key
}

// Clear removes the specific bit flag(s).
func (b *BitMask[T]) Clear(key T) {
	b.Flag &= ^key
}

// ClearAll resets all bits in the bitmask to 0, clearing all set flags.
func (b *BitMask[T]) ClearAll() {
	b.Flag = 0
}

// Toggle flips the state of the specific bit flag(s).
func (b *BitMask[T]) Toggle(key T) {
	b.Flag ^= key
}

// --- Usage Examples (Aliases for backward compatibility if needed) ---

// SmBitMask is an alias for an 8-bit mask (uint8).
type SmBitMask = BitMask[uint8]

// LgBitMask is an alias for a 32-bit mask (uint32).
type LgBitMask = BitMask[uint32]

// XlBitMask is an alias for a 64-bit mask (uint64).
type XlBitMask = BitMask[uint64]
