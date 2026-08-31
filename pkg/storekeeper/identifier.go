/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"fmt"
	"strconv"
)

/*

The Interlace Architecture uses three custom identifiers:

1) CID to organize categories
2) LID to localize features in ordinal position and lexicographical order for fast retrieval
3) GID to analyze features and annotations in geospatial space

1. CID
The Interlace Category ID is a 3-letter uppercase lexicographical identifier [A-Z].

2. LID
The Interlace Library ID is an alphanumeric lexicographical compound ID of 34 bytes.

+ 3-bytes string: the CID
+ 10-byte string: uint32 ordinal position 0-left padding
+ 20-byte string: uin64 number 0-left padding, this can be a DB unique ID or any 64-bit hash
+ 1-byte rune: uint8 ASCII character [a-z,A-Z]

This compound ID can be stored in disk in just 13 bytes, by partition folders AAA, AAB...ZZZ and as integer values:

- 4-byte uint32: ordinal position
- 8-byte uint64: feature ID
- 1-byte uint8: ASCII character [a-z,A-Z]

___ _ ___ __| ___ _ __ ___ |__ ___ ___ _|_ ___
123 4 567 890 123 4 56 789 012 345 678 901 234
AAA 4 294 967 295 A 18 446 744 073 709 551 615

3. GID

The Interlace Architecture maps standard one-dimensional arrays to a two-dimensional geospatial coordinate
system. In this paradigm, longitude strictly maps the ordinal array span (the physical structure), while latitude maps the
discrete feature track (e.g., biological function or category). This allows geospatial APIs to resolve complex positional
intersections mathematically rather than iteratively.

Longitude:

In a 360-degree range, the Interlace Architecture can map two 1D arrays, one ranging ordinal positions from
0 to 180 degrees East and the other 0 to 180 degrees West; where:

ArrLen = len(array).
Direction = 1 for East and -1 for West
Degree longitude = (array-index + 1) * 180.000000 * Direction / ArrLen

Latitude:

Goes from 0 degrees at the middle up to 90 degrees North at the North Pole and 90 degrees South at the South Pole.

MapSize = number of ordinal arrays or size of map["CID"]GIndexer.
Direction = 1 for the North Pole and -1 for the South Pole.
Degree latitude = (cid-index + 1) * 90.000000 * Direction / MapSize

*/

// CIndexer manages a 3-letter identifier with thread-safe modifications and lexicographical navigation.
type CIndexer struct {
	a, b, c byte
}

// LIndexer represents a lightweight identifier manager combining CID, position, category, and UID into a 34-byte buffer.
type LIndexer struct {
	cid string
	pos uint32
	uid uint64
	cat uint8
	// A fixed-size array allocates on the stack (zero Garbage Collection overhead).
	buf [34]byte
}

// GIndexer represents a geographical indexer.
type GIndexer struct {
	cid       string
	lat       float64
	arrLen    float64
	mapSize   float64
	cidIndex  float64
	direction float64
}

// NewLIndexer creates and returns a pointer to an LIndexer instance, validating that cid contains only [A-Z].
func NewLIndexer(cid string, pos uint32, uid uint64, cat uint8) *LIndexer {
	li := LIndexer{}
	li.Reset(cid, pos, uid, cat)
	return &li
}

// Reset reinitialize the LIndexer instance with the provided parameters if the CID has exactly 3 uppercase alphabetic characters.
func (li *LIndexer) Reset(cid string, pos uint32, uid uint64, cat uint8) {
	li.cid = cid
	li.pos = pos
	li.uid = uid
	li.cat = cat
	if ValidCID(cid) && ValidCat(cat) {
		li.formatID()
	} else {
		li.cid = ""
		li.pos = 0
		li.uid = 0
		li.cat = 0
	}
}

// ValidCID checks if the given string is a CID with exactly 3 uppercase alphabetic characters.
func ValidCID(cid string) bool {
	if len(cid) != 3 {
		return false
	}
	for i := range 3 {
		if cid[i] < 'A' || cid[i] > 'Z' {
			return false
		}
	}
	return true
}

// ValidCat checks if the given category is a valid character.
func ValidCat(cat uint8) bool {
	return (cat >= 'A' && cat <= 'Z') || (cat >= 'a' && cat <= 'z')
}

// Update updates the internal CID values of the LIndexer instance with the provided position and uid.
func (li *LIndexer) Update(pos uint32, uid uint64) {
	li.pos = pos
	li.uid = uid
	li.formatID()
}

// formatID constructs a 34-byte identifier combining CID, position, category, and UID into a single buffer.
// This code allocates nothing; until the final LID string is created with li.String().
func (li *LIndexer) formatID() {
	if li.cat == 0 {
		return
	}
	// 1. cid (3 bytes)
	_ = li.cid[2] // Bounds check elimination: tells the compiler to check length once
	li.buf[0] = li.cid[0]
	li.buf[1] = li.cid[1]
	li.buf[2] = li.cid[2]

	// 2. pos (10 bytes, uint32)
	// Write backwards from index 12 down to 3
	pos := li.pos
	for i := 12; i >= 3; i-- {
		li.buf[i] = byte(pos%10) + '0'
		pos /= 10
	}

	// 3. cat (1 byte, uint8)
	li.buf[13] = li.cat

	// 4. uid (20 bytes, uint64)
	// Write backwards from index 33 down to 14
	uid := li.uid
	for i := 33; i >= 14; i-- {
		li.buf[i] = byte(uid%10) + '0'
		uid /= 10
	}
}

// String returns the string representation of the LIndexer instance: CID[3] POS[10] CAT[1] UID[20] or empty if error.
func (li *LIndexer) String() string {
	if li.cid == "" {
		return ""
	}
	// Creates exactly one heap allocation for the final string
	return string(li.buf[:])
}

// Factor returns the CID, position, category, and UID values of the LIndexer instance.
func (li *LIndexer) Factor() (string, uint32, uint8, uint64) {
	return li.cid, li.pos, li.cat, li.uid
}

// NewCIndexer initializes and returns a new instance of CIndexer with the CID set to "AAA".
func NewCIndexer() *CIndexer {
	return &CIndexer{
		a: 'A',
		b: 'A',
		c: 'A',
	}
}

// String returns the 3-letter uppercase CID, this is the standard identifier form.
func (ci *CIndexer) String() string {
	return fmt.Sprintf("%c%c%c", ci.a, ci.b, ci.c)
}

// LowerCase returns the 3-letter lowercase CID used for naming partition directories.
func (ci *CIndexer) LowerCase() string {
	return fmt.Sprintf("%c%c%c", ci.a+32, ci.b+32, ci.c+32)
}

// Set updates the internal 3-letter uppercase identifier values with the provided runes a, b, and c.
// It returns an error if any of the input values are not uppercase letters (A-Z).
func (ci *CIndexer) Set(a, b, c uint8) error {
	if a < 'A' || a > 'Z' || b < 'A' || b > 'Z' || c < 'A' || c > 'Z' {
		return fmt.Errorf("invalid rune")
	}
	ci.a = a
	ci.b = b
	ci.c = c
	return nil
}

// Reset sets the internal 3-letter uppercase identifier values to "AAA".
func (ci *CIndexer) Reset() {
	ci.a = 'A'
	ci.b = 'A'
	ci.c = 'A'
}

// Next increments the internal 3-letter identifier to its next value in lexicographical order, wrapping when necessary.
func (ci *CIndexer) Next() {
	if ci.String() == "ZZZ" {
		return
	}
	if ci.c < 'Z' {
		ci.c++
	} else if ci.b < 'Z' {
		ci.c = 'A'
		ci.b++
	} else {
		ci.c = 'A'
		ci.b = 'A'
		ci.a++
	}
}

// Back decrements the internal 3-letter identifier to its previous value in lexicographical order, wrapping when necessary.
func (ci *CIndexer) Back() {
	if ci.String() == "AAA" {
		return
	}
	if ci.c != 'A' {
		ci.c--
	} else if ci.b != 'A' {
		ci.c = 'Z'
		ci.b--
	} else {
		ci.c = 'Z'
		ci.b = 'Z'
		ci.a--
	}
}

// NewGIndexer creates and initializes a new GIndexer instance with the size = len(map["CID"]GIndexer).
func NewGIndexer(mapSize float64) *GIndexer {
	gi := GIndexer{}
	gi.Reset(mapSize)
	return &gi
}

// Reset clears the GIndexer fields and resets the size = len(map["CID"]GIndexer).
func (gi *GIndexer) Reset(mapSize float64) {
	gi.cid = ""
	gi.lat = 0.0
	gi.arrLen = 0.0
	gi.mapSize = mapSize
	gi.cidIndex = 0
	gi.direction = 0.0
}

// Set assigns the CID, array length, and index to the GIndexer while validating inputs and updating the latitude field.
func (gi *GIndexer) Set(cid string, arrLen, cidIndex float64) error {
	if !ValidCID(cid) {
		return fmt.Errorf("invalid CID: %s", cid)
	}
	if cidIndex < 0 || cidIndex >= gi.mapSize {
		return fmt.Errorf("the CID index is out of range: %.0f", cidIndex)
	}
	gi.cid = cid
	gi.arrLen = arrLen
	gi.cidIndex = cidIndex
	gi.setDirection()
	gi.setLatitude()
	return nil
}

// setDirection returns to -1 or 1 based on the CID index.
func (gi *GIndexer) setDirection() {
	if (int(gi.cidIndex)+1)%2 == 0 {
		gi.direction = 1.0
	} else {
		gi.direction = -1.0
	}
}

// setLatitude calculates and assigns the latitude
// MapSize = number of ordinal arrays or size of map["CID"]GIndexer.
// Direction = 1 for the North Pole and -1 for the South Pole.
// Degree latitude = (cid-index + 1) * 90.000000 * Direction / MapSize
func (gi *GIndexer) setLatitude() {
	gi.lat = (gi.cidIndex + 1) * 90.0 * gi.direction / gi.mapSize
}

// Latitude computes the latitude for the given annotationTrack or returns the pre-calculated latitude if the offset is 0.
func (gi *GIndexer) Latitude(annotationTrack float64) float64 {
	if annotationTrack == 0 {
		return gi.lat
	}
	return gi.lat + 180*gi.direction/gi.mapSize/annotationTrack
}

// Longitude returns the longitude based on the position or 0 if the position is out of range [1,len(uint32)]
// ArrLen = len(array).
// Direction = 1 for East and -1 for West
// Degree longitude = (array-index + 1) * 180.000000 * Direction / ArrLen
func (gi *GIndexer) Longitude(pos float64) float64 {
	return min(pos, gi.arrLen) * 180 * gi.direction / gi.arrLen
}

// String returns decimal degrees, first latitude, then longitude; for the given position and annotation track.
// Example: "41.40338, 2.17403"
func (gi *GIndexer) String(pos, annotationTrack float64) string {
	b := []byte(", ")
	b = strconv.AppendFloat(b, gi.Longitude(pos), 'g', -1, 64)
	return strconv.FormatFloat(gi.Latitude(annotationTrack), 'g', -1, 64) + string(b)
}
