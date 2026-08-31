/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"math"
	"strings"

	"github.com/sas-quantome/interlace-go/pkg/runner"
)

const (
	EncIsNA uint32 = 1 << 30
	EncIsUI uint32 = 1 << 31
)

const (
	EncNA byte = 'N'
	EncUI byte = 'U'
)

const maskSize = 30

type Encoder struct {
	err  error
	enum *Enumerator
}

type NucleotideEncoder struct {
	index   int
	nBases  int
	NSpan   uint32
	key     byte
	mask    []uint32
	encoder map[byte][]byte
	bitMask runner.LgBitMask
}

func ACGTEnum() []byte {
	return []byte{
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', '1', '0',
	}
}

func IUPACEnum() []byte {
	return []byte{
		'N', 'B', 'D', 'H',
		'A', 'T', 'C', 'G',
		'K', 'M', 'R', 'S',
		'A', 'T', 'C', 'G',
		'V', 'W', 'Y', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', 'C', 'G',
		'A', 'T', '0', '1',
	}
}

func TrinucleotideEnum() []string {
	return []string{"UNK",
		"AAA", "AAC", "AAG", "AAT", "ACA",
		"ACC", "ACG", "ACT", "AGA", "AGC",
		"AGG", "AGT", "ATA", "ATC", "ATG",
		"ATT", "CAA", "CAC", "CAG", "CAT",
		"CCA", "CCC", "CCG", "CCT", "CGA",
		"CGC", "CGG", "CGT", "CTA", "CTC",
		"CTG", "CTT", "GAA", "GAC", "GAG",
		"GAT", "GCA", "GCC", "GCG", "GCT",
		"GGA", "GGC", "GGG", "GGT", "GTA",
		"GTC", "GTG", "GTT", "TAA", "TAC",
		"TAG", "TAT", "TCA", "TCC", "TCG",
		"TCT", "TGA", "TGC", "TGG", "TGT",
		"TTA", "TTC", "TTG", "TTT", "GNT",
		"GNG", "GNC", "GNA", "CNT", "CNG",
		"CNC", "CNA", "TNA", "TNC", "TNG",
		"TNT", "NAA", "NAC", "NAG", "NAT",
		"NCA", "NCC", "NCG", "NCT", "TTN",
		"TGN", "TCN", "TAN", "GTN", "GGN",
		"GCN", "GAN", "CTN", "ANT", "ANG",
		"ANC", "ANA", "ATN", "AGN", "ACN",
		"AAN", "CCN", "NGA", "NGC", "NGG",
		"NGT", "CGN", "NTA", "NTC", "CAN",
		"NGN", "NCN", "NAN", "ANN", "GNN",
		"CNN", "TNN", "NTG", "NTT", "NTN",
		"NNA", "NNC", "NNG", "NNT", "NNN"}
}

func NewNucleotideEncoder() *NucleotideEncoder {
	na := new(NucleotideEncoder)
	na.Clear(true)
	na.mask = make([]uint32, maskSize)
	var n uint32 = 1
	na.mask[0] = n
	for i := 1; i < len(na.mask); i++ {
		n = 2 * n
		na.mask[i] = n
	}
	na.encoder = make(map[byte][]byte)
	na.encoder[EncNA] = ACGTEnum()
	na.encoder[EncUI] = IUPACEnum()
	return na
}

// NewEncoder initializes and returns a pointer to an Encoder using the provided enumeration of strings.
func NewEncoder(enum []string) *Encoder {
	ec := new(Encoder)
	ec.enum = NewEnumerator()
	ec.err = ec.enum.Create(enum)
	return ec
}

// Number returns int or -1 if not found.
func (ec *Encoder) Number(s string) int {
	x := ec.enum.GetCode(s)
	if x == math.MinInt64 {
		return -1
	}
	return x
}

// String returns string or empty string if not found.
func (ec *Encoder) String(n int) string {
	s := ec.enum.GetName(n)
	return s
}

func (ec *Encoder) Err() error {
	return ec.err
}

// Next searches for the next occurrence of the specified base starting from the given index within the encoder's mapping.
// If no encoder, it returns -1.
func (na *NucleotideEncoder) Next(index int, base byte) int {
	if na.NoEncoder() {
		return -1
	}
	for i := index + 1; i < maskSize; i++ {
		if na.encoder[na.key][i] == base {
			return i
		}
	}
	return -1
}

// Clear resets the NucleotideEncoder by clearing its bitmask, resetting index to -1, and initializing key and NSpan to zero.
func (na *NucleotideEncoder) Clear(nBasesCounterEnds bool) {
	na.bitMask.ClearAll()
	na.NSpan = 0
	na.index = -1
	na.key = 0
	if nBasesCounterEnds {
		na.nBases = 0
	}
}

// Reset initializes the encoder state based on the provided base, updating the bitmask and resetting core properties.
func (na *NucleotideEncoder) Reset(base byte) {
	na.Clear(false)
	if strings.ContainsRune("ACGT", rune(base)) {
		na.key = EncNA
		na.bitMask.Add(EncIsNA)
	} else {
		na.key = EncUI
		na.bitMask.Add(EncIsUI)
	}
}

// Encode encodes the specified base into the NucleotideEncoder, updating the bitmask and resetting NSpan if necessary.
func (na *NucleotideEncoder) Encode(base byte) {
	na.index = na.Next(na.index, base)
	if na.index >= 0 && na.index < maskSize {
		na.bitMask.Add(na.mask[na.index])
		if base != 'N' {
			na.nBases++
		}
	}
}

// UpdateMask clears the bitmask of the NucleotideEncoder, replaces it with the provided uint32 value, and updates nBases.
func (na *NucleotideEncoder) UpdateMask(b uint32) {
	na.bitMask.ClearAll()
	na.bitMask.Add(b)
	if b&EncIsNA == 0 && b&EncIsUI == 0 {
		na.nBases++
	}
}

// BitMask returns the bitmask-encoded sequence.
func (na *NucleotideEncoder) BitMask() uint32 {
	return na.bitMask.Flag
}

func (na *NucleotideEncoder) NBases() int {
	return na.nBases
}

func (na *NucleotideEncoder) NotIndexed() bool {
	return na.index == -1
}

func (na *NucleotideEncoder) NoEncoder() bool {
	return na.key != EncNA && na.key != EncUI
}

// Decode decodes the bitmask-encoded sequence into a string.
func (na *NucleotideEncoder) Decode(bitFlags uint32) string {
	var buf strings.Builder
	var key byte
	if bitFlags&EncIsNA != 0 {
		key = EncNA
	} else if bitFlags&EncIsUI != 0 {
		key = EncUI
	} else {
		for range bitFlags {
			buf.WriteByte('N')
		}
		return buf.String()
	}
	for i := range na.mask {
		if bitFlags&na.mask[i] != 0 {
			buf.WriteByte(na.encoder[key][i])
		}
	}
	return buf.String()
}
