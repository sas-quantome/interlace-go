/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"bufio"
	"cmp"
	"fmt"
	"hash"
	"hash/fnv"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sas-quantome/interlace-go/pkg/runner"
)

// CatalogItem represents a single entry in the catalog for sorting.
type CatalogItem struct {
	Hash uint64
	Text string
}

// Cataloguer is responsible for managing persistent catalogs of unique text strings.
// It deduplicates strings by hash and handles hash collisions by chaining.
type Cataloguer struct {
	tmpFile *os.File
	fnv     hash.Hash64

	// Paths and File Names
	binPath string
	hidFn   string
	txtFn   string

	// In-memory storage
	items []CatalogItem

	// State trackers
	newRecords    int
	oldRecords    int
	newCollisions int
	fileOpen      bool
}

// NewCataloguer initializes a new Cataloguer.
func NewCataloguer() *Cataloguer {
	return &Cataloguer{
		fnv:   fnv.New64a(),
		items: make([]CatalogItem, 0, 1024),
	}
}

// SetBinPathAndFileNames sets the configuration for file storage.
func (c *Cataloguer) SetBinPathAndFileNames(binPath, catKeyName string) {
	c.binPath = binPath
	c.hidFn = fmt.Sprintf("suk_%s_hash_id_m.bin", catKeyName)
	c.txtFn = fmt.Sprintf("uuk_%s_text_s.bin", catKeyName)
}

// CreateTmpFile creates a temporary file for staging new entries.
func (c *Cataloguer) CreateTmpFile(tmpDir string) error {
	var err error
	if c.tmpFile, err = os.CreateTemp(tmpDir, "hash_text_*.tsv"); err != nil {
		return err
	}
	c.fileOpen = true
	return nil
}

// HashText computes the FNV-1a 64-bit hash of a string.
func (c *Cataloguer) HashText(s string) uint64 {
	c.fnv.Reset()
	_, _ = c.fnv.Write([]byte(s))
	return c.fnv.Sum64()
}

// Add computes the hash and writes the tuple to the temporary staging file.
func (c *Cataloguer) Add(text string) error {
	h := c.HashText(text)
	return c.WriteTuple(h, text)
}

// WriteTuple writes a raw hash-text pair to the temporary staging file.
func (c *Cataloguer) WriteTuple(hashId uint64, text string) error {
	if hashId == 0 || text == "" {
		return fmt.Errorf("invalid input: hash=[%d] txt=[%s]", hashId, text)
	}
	// Use Fprintln for cleaner writing (handles newlines automatically)
	_, err := fmt.Fprintf(c.tmpFile, "%d\t%s\n", hashId, text)
	return err
}

// Save merges new data with the existing catalog, deduplicates it, and persists it.
func (c *Cataloguer) Save() error {
	// 1. Close the writing stream to the temp file so we can read it back.
	if err := c.CloseTmpFile(); err != nil {
		return err
	}
	// Ensure a temp file is removed eventually
	// defer c.RemoveTmpFile()

	// 2. Load the existing catalog (if any) into memory.
	if err := c.loadOldCatalog(); err != nil {
		// If the error is other than "Not Exist", return it.
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load old catalog: %w", err)
		}
		// If files don't exist, we just start fresh.
	}

	// 3. Read new items from the temp file and append them to memory.
	if err := c.ingestTempFile(); err != nil {
		return fmt.Errorf("failed to ingest temp file: %w", err)
	}

	// 4. Sort and Deduplicate in memory.
	c.processCatalog()

	// 5. Back up old files if this was an update.
	// (Check if oldRecords > 0 to imply previous existence or check file existence)
	if c.oldRecords > 0 {
		if err := c.backup(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	// 6. Save the new state.
	return c.saveToDisk()
}

// ingestTempFile reads the temporary TSV file and appends items to the internal slice.
func (c *Cataloguer) ingestTempFile() error {
	f, err := os.Open(c.tmpFile.Name())
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		hid, per := strconv.ParseUint(parts[0], 10, 64)
		if per != nil {
			continue
		}
		c.items = append(c.items, CatalogItem{
			Hash: hid,
			Text: parts[1],
		})
	}
	return runner.ValidateErrors(scanner.Err(), f.Close())
}

// loadOldCatalog loads the existing bin files into the items slice.
func (c *Cataloguer) loadOldCatalog() error {
	// Check if a hash file exists
	if _, err := os.Stat(c.hashFp()); err != nil {
		return err
	}

	var e1, e2 error

	t := NewContainer(c.txtFn)
	if err := t.OpenStrArr(c.textFp(), false); err != nil {
		return err
	}
	defer func() {
		e1 = t.CloseMappedArr()
	}()

	h := NewContainer(c.hidFn)
	if err := h.OpenNumArr(c.hashFp(), false); err != nil {
		return runner.ValidateErrors(e1, err)
	}
	defer func() {
		e2 = h.CloseMappedArr()
	}()

	if t.Len != h.Len {
		return fmt.Errorf("corrupt catalog: hash count %d != text count %d", h.Len, t.Len)
	}

	c.oldRecords = h.Len

	// Efficiently extend the slice
	if cap(c.items) < h.Len {
		newItems := make([]CatalogItem, 0, h.Len+1024)
		c.items = append(newItems, c.items...)
	}

	for i := 0; i < h.Len; i++ {
		c.items = append(c.items, CatalogItem{
			Hash: h.MArr[i],
			Text: t.SArr[i],
		})
	}

	return runner.ValidateErrors(e1, e2)
}

// processCatalog sorts the items by Hash and merges duplicates (collision handling).
func (c *Cataloguer) processCatalog() {
	if len(c.items) == 0 {
		return
	}

	// 1. Sort by Hash.
	slices.SortFunc(c.items, func(a, b CatalogItem) int {
		return cmp.Compare(a.Hash, b.Hash)
	})

	// 2. Compact (Deduplicate and Handle Collisions).
	// We rewrite the slice in-place to save memory.
	writeIdx := 0
	for readIdx := 1; readIdx < len(c.items); readIdx++ {
		current := c.items[readIdx]
		prev := &c.items[writeIdx]

		if current.Hash == prev.Hash {
			// Collision detected
			if current.Text != prev.Text {
				// True collision or just a duplicate of the same text?
				// The original logic assumes chaining with Pipe for ANY collision of hash.
				// We must check if the text is actually inside to avoid "Text|Text".
				if !strings.Contains(prev.Text, current.Text) {
					prev.Text = fmt.Sprintf("%s|%s", prev.Text, current.Text)
					c.newCollisions++
				}
			}
			// If texts are identical, we do nothing (it's a duplicate record), effectively skipping it.
		} else {
			// No collision, move write index forward and copy
			writeIdx++
			c.items[writeIdx] = current
		}
	}

	// Reslice to the correct length
	c.items = c.items[:writeIdx+1]
	c.newRecords = len(c.items)
}

// saveToDisk writes the optimized in-memory slices to bin files.
func (c *Cataloguer) saveToDisk() error {
	// Separate struct into parallel arrays for storage
	hashes := make([]uint64, len(c.items))
	texts := make([]string, len(c.items))

	for i, item := range c.items {
		hashes[i] = item.Hash
		texts[i] = item.Text
	}

	t := NewContainer(c.txtFn)
	h := NewContainer(c.hidFn)
	return runner.ValidateErrors(t.SaveStrArr(c.textFp(), texts), h.SaveNumArr(c.hashFp(), hashes))
}

// backup moves current files to a timestamped backup folder.
func (c *Cataloguer) backup() error {
	// Ensure backup root exists
	backupRoot := path.Join(c.binPath, "backup")
	if err := os.MkdirAll(backupRoot, 0775); err != nil {
		return err
	}

	// Create a timestamped folder
	// Using generic time format since TimeTagNow() was not provided in context
	tsFolder := path.Join(backupRoot, time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tsFolder, 0775); err != nil {
		return err
	}

	// Rename files
	if err := os.Rename(c.hashFp(), path.Join(tsFolder, c.hidFn)); err != nil {
		return err
	}
	return os.Rename(c.textFp(), path.Join(tsFolder, c.txtFn))
}

func (c *Cataloguer) CloseTmpFile() error {
	if c.fileOpen && c.tmpFile != nil {
		c.fileOpen = false
		return c.tmpFile.Close()
	}
	return nil
}

func (c *Cataloguer) RemoveTmpFile() error {
	_ = c.CloseTmpFile()
	if c.tmpFile != nil {
		return os.Remove(c.tmpFile.Name())
	}
	return nil
}

// Helpers for file paths
func (c *Cataloguer) hashFp() string { return path.Join(c.binPath, c.hidFn) }
func (c *Cataloguer) textFp() string { return path.Join(c.binPath, c.txtFn) }
