/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
)

// Framer manages a thread-safe collection of slot-filler pairs (Key-Value store).
// It uses a standard map for O(1) lookups.
type Framer struct {
	mu    sync.RWMutex
	store map[string]any
}

// NewFramer initializes a new Framer instance.
func NewFramer() *Framer {
	return &Framer{
		store: make(map[string]any),
	}
}

// LoadFrameFromTable reads a tab-delimited file and populates the frame.
// colSlot: column index for the Key. colFiller: column index for the Value.
func (f *Framer) LoadFrameFromTable(fn string, colSlot, colFiller, maxLineSize int) error {
	scanner := NewScanner(maxLineSize)
	if err := scanner.Start(fn); err != nil {
		return ValidateErrors(err, scanner.Close())
	}

	// Use a local map for batch loading to minimize lock contention
	batch := make(map[string]any)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		if len(parts) > colSlot && len(parts) > colFiller {
			slot := parts[colSlot]
			filler := parts[colFiller]
			batch[slot] = filler
		}
	}

	if scanner.Err() != nil {
		return ValidateErrors(scanner.Err(), scanner.Close())
	}

	// Bulk insert protected by lock
	f.InsertBatch(batch)
	return scanner.Close()
}

// LoadFrame populates the frame assuming col 0 is Slot and col 1 is Filler.
// If the fn contains ".gob", it loads from the gob file.
func (f *Framer) LoadFrame(fn string) error {
	if strings.Contains(fn, ".gob") {
		return LoadDataFromGobFile(fn, &f.store)
	}
	return f.LoadFrameFromTable(fn, 0, 1, 0)
}

// Records return the total number of items.
func (f *Framer) Records() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.store)
}

// Insert adds or updates a slot-filler pair.
func (f *Framer) Insert(slot string, filler any) {
	if slot == "" {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[slot] = filler
}

// InsertBatch adds multiple items efficiently.
func (f *Framer) InsertBatch(items map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for k, v := range items {
		if k != "" {
			f.store[k] = v
		}
	}
}

// SearchSlot retrieves the raw value for a slot. Returns nil if not found.
func (f *Framer) SearchSlot(slot string) any {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.store[slot]
}

// SearchSlotS retrieves a string value safely.
func (f *Framer) SearchSlotS(slot string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if val, ok := f.store[slot]; ok {
		if str, isString := val.(string); isString {
			return str
		}
	}
	return ""
}

// SearchSlotI retrieves an int value safely.
func (f *Framer) SearchSlotI(slot string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if val, ok := f.store[slot]; ok {
		if i, isInt := val.(int); isInt {
			return i
		}
	}
	return 0
}

// SearchSlotF retrieves a float64 value safely.
func (f *Framer) SearchSlotF(slot string) float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if val, ok := f.store[slot]; ok {
		if fl, isFloat := val.(float64); isFloat {
			return fl
		}
	}
	return math.NaN()
}

// SlotContains searches for the first slot that contains the substring 'pattern'.
// Note: This requires a full scan O(N).
func (f *Framer) SlotContains(pattern string) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var matches []string
	for slot, val := range f.store {
		if strings.Contains(slot, pattern) {
			if str, ok := val.(string); ok {
				matches = append(matches, str)
			}
		}
	}
	sort.Strings(matches)
	return matches
}

// Delete removes a slot.
func (f *Framer) Delete(slot string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.store, slot)
}

// Clear empties the frame efficiently.
func (f *Framer) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Optimized clear: allow GC to reclaim the old map
	f.store = make(map[string]any)
}

// PadWrite writes the frame formatted with a tabwriter.
// Keys are sorted alphabetically for a consistent output.
func (f *Framer) PadWrite(output io.Writer) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	keys := f.sortedKeys()
	w := tabwriter.NewWriter(output, 0, 0, 1, ' ', 0)

	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s:\t%v\n", k, f.store[k]); err != nil {
			return err
		}
	}
	return w.Flush()
}

// TabWrite writes the frame using a delimiter.
// Keys are sorted alphabetically for a consistent output.
func (f *Framer) TabWrite(output io.Writer, delimiter string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	keys := f.sortedKeys()
	w := bufio.NewWriter(output)

	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s%s%v\n", k, delimiter, f.store[k]); err != nil {
			return err
		}
	}
	return w.Flush()
}

// Log writes all key-value pairs from the Framer's store through the logger.
func (f *Framer) Log(mgr *Manager) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	exclude := []string{AppName, AppVersion, AppProject, AppAim}
	keys := f.sortedKeys()
	for _, k := range keys {
		if slices.Index(exclude, k) == -1 {
			mgr.Say(fmt.Sprintf("%s: %v", k, f.store[k]))
		}
	}
}

// Save writes the frame to a file using TabWrite.
func (f *Framer) Save(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return ValidateErrors(f.TabWrite(file, "\t"), file.Close())
}

// Grep searches for a pattern in filler and returns the matching slots.
func (f *Framer) Grep(pattern string) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	keys := make([]string, 0, len(f.store))
	for k, v := range f.store {
		if strings.Contains(fmt.Sprintf("%v", v), pattern) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// ReplaceVariables replaces all occurrences of {{variable}} in the dictionary frame.
func (f *Framer) ReplaceVariables(dictionary *Framer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for key, val := range f.store {
		switch val.(type) {
		case string:
			filler := val.(string)
			if strings.Contains(filler, "{{") {
				var u []string
				s := strings.SplitSeq(filler, "{{")
				for v := range s {
					t := strings.Split(v, "}}") // has been checked before
					u = InsertOnce(u, t[0])
				}
				for _, variableName := range u {
					parameter := dictionary.SearchSlotS(variableName) // it exists, has been checked before
					filler = strings.ReplaceAll(filler, "{{"+variableName+"}}", parameter)
				}
				f.store[key] = filler
			}
		}
	}
}

// sortedKeys returns a sorted slice of all keys in the map.
// Caller must hold the lock.
func (f *Framer) sortedKeys() []string {
	keys := make([]string, 0, len(f.store))
	for k := range f.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
