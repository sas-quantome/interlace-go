/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// Enumerator represents a bidirectional map of string names to integer codes.
// It is thread-safe and optimized for O(1) lookups in both directions.
type Enumerator struct {
	mu      sync.RWMutex
	nameMap map[string]int
	codeMap map[int]string
	nextID  int
}

// NewEnumerator initializes a new, empty Enumerator.
func NewEnumerator() *Enumerator {
	return &Enumerator{
		nameMap: make(map[string]int),
		codeMap: make(map[int]string),
		nextID:  0,
	}
}

// Len returns the number of items in the enumerator.
func (e *Enumerator) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.nameMap)
}

// Load loads enumeration data from a bin-encoded string slice file.
// The index of the string in the file becomes its code.
func (e *Enumerator) Load(binFile string) error {

	co := NewContainer(binFile)
	if err := co.OpenStrArr(binFile, false); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.reset()
	for code, name := range co.SArr {
		e.addUnsafe(strings.Clone(name), code)
	}
	// Ensure the nextID is strictly greater than the highest-loaded code
	e.nextID = len(co.SArr)

	return co.CloseMappedArr()
}

// Save stores the enumerator's elements in a bin-encoded file.
// Elements are sorted by code to ensure deterministic file output (0, 1, 2...).
func (e *Enumerator) Save(binFile string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.codeMap) == 0 {
		return errors.New("no enumerations to save")
	}

	// Reconstruct the slice based on codes.
	// Note: We assume codes are dense (0...N) because Load/Append creates them that way.
	// If codes are sparse, we must sort the keys first.
	codes := make([]int, 0, len(e.codeMap))
	for code := range e.codeMap {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	output := make([]string, len(codes))
	for i, code := range codes {
		output[i] = e.codeMap[code]
	}
	co := NewContainer(binFile)
	return co.SaveStrArr(binFile, output)
}

// Create initializes the Enumerator with a list of enum names.
func (e *Enumerator) Create(names []string) error {
	if len(names) == 0 {
		return errors.New("cannot create empty enumerator")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.reset()
	for i, name := range names {
		e.addUnsafe(name, i)
	}
	e.nextID = len(names)
	return nil
}

// Append adds a new enumeration item.
// If the name already exists, it does nothing.
// If it is new, it assigns the next available integer code.
func (e *Enumerator) Append(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.nameMap[name]; exists {
		return
	}

	// Assign the next available ID
	e.addUnsafe(name, e.nextID)
	e.nextID++
}

// Assure retrieves the code for the given name. If the name does not exist,
// it creates it, assigns a new code, and returns the new code.
func (e *Enumerator) Assure(name string) int {
	e.mu.RLock()
	code, exists := e.nameMap[name]
	e.mu.RUnlock()

	if exists {
		return code
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	// Double-check locking to prevent race conditions
	if code, exists = e.nameMap[name]; exists {
		return code
	}

	code = e.nextID
	e.addUnsafe(name, code)
	e.nextID++
	return code
}

// GetName retrieves the name associated with the given code.
// Returns an empty string if not found.
func (e *Enumerator) GetName(code int) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.codeMap[code]
}

// GetCode retrieves the code associated with the given name.
// Returns math.MinInt64 if not found.
func (e *Enumerator) GetCode(name string) int {
	if name == "" {
		return math.MinInt64
	}
	e.mu.RLock()
	defer e.mu.RUnlock()

	if code, ok := e.nameMap[name]; ok {
		return code
	}
	return math.MinInt64
}

// Print prints each element sorted by code.
func (e *Enumerator) Print(delimiter string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Sort keys for consistent printing
	codes := make([]int, 0, len(e.codeMap))
	for code := range e.codeMap {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	for _, code := range codes {
		fmt.Printf("%d%s%s\n", code, delimiter, e.codeMap[code])
	}
}

// --- Internal Helper Methods ---

// reset clears the maps. Caller must hold Lock.
func (e *Enumerator) reset() {
	e.nameMap = make(map[string]int)
	e.codeMap = make(map[int]string)
	e.nextID = 0
}

// addUnsafe adds an entry without locking. Caller must hold Lock.
func (e *Enumerator) addUnsafe(name string, code int) {
	e.nameMap[name] = code
	e.codeMap[code] = name
}
