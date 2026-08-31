/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"errors"
	"path"
	"strconv"

	"github.com/sas-quantome/interlace-go/pkg/extractor"
	"github.com/sas-quantome/interlace-go/pkg/runner"
)

const (
	SDT = byte('s') // code for array data type: string
	IDT = byte('i') // code for array data type: int64
	NDT = byte('n') // code for array data type: int32
	DDT = byte('d') // code for array data type: uint8
	UDT = byte('u') // code for array data type: uint32
	MDT = byte('m') // code for array data type: uint64
	FDT = byte('f') // code for array data type: float64
)

// ItemOrderType provides the sorting order of items.
type ItemOrderType int8

const (
	UDK ItemOrderType = iota // unsorted duplicate keys
	UUK                      // unsorted unique keys
	SDK                      // sorted duplicate keys
	SUK                      // sorted unique keys
)

// Interlacer represents a structure for creating bin-encoded data assets from a TSV file.
type Interlacer struct {
	HeadOut int    // Number of header rows to skip.
	BufSize int    // Buffer size for read.
	BinPath string // Output directory.
	TsvFile string // Input file path.

	// Internal optimization
	idxMap        map[int]int // Used only during setup map[col]index of []BinAsset
	colIndexCache []int       // Used during processing (fast lookup)

	BinData []BinAsset // Assets to process
}

// BinAsset manages Stacks and generates bin-encoded slices, column count starts at 0.
type BinAsset struct {
	ItemColN int
	FileName string
	DataType byte
	IdxOrder ItemOrderType

	// Explicit stacks (Same as before)
	StackZDT runner.Stack[[]byte]
	StackSDT runner.Stack[string]
	StackIDT runner.Stack[int64]
	StackNDT runner.Stack[int32]
	StackDDT runner.Stack[uint8]
	StackUDT runner.Stack[uint32]
	StackMDT runner.Stack[uint64]
	StackFDT runner.Stack[float64]
}

func NewInterlacer(bufSize, nHeadRowsToSkip int, tsvFile, binPath string) *Interlacer {
	return &Interlacer{
		HeadOut: nHeadRowsToSkip,
		BufSize: bufSize,
		BinPath: binPath,
		TsvFile: tsvFile,
		idxMap:  make(map[int]int),
	}
}

// AddBinAsset adds a new asset.
func (interlacer *Interlacer) AddBinAsset(itemColN int, fileName string, dType byte, order ItemOrderType) error {
	if itemColN < 0 {
		return errors.New("item column number must be greater or equal to zero")
	}
	if _, found := interlacer.idxMap[itemColN]; found {
		return errors.New("item column number already entered")
	}
	asset := BinAsset{
		ItemColN: itemColN,
		FileName: fileName,
		DataType: dType,
		IdxOrder: order,
	}
	// Initialize the specific stack to ensure it's ready (clearing previous garbage if any)
	switch dType {
	case SDT:
		asset.StackSDT.Clear()
	case IDT:
		asset.StackIDT.Clear()
	case NDT:
		asset.StackNDT.Clear()
	case DDT:
		asset.StackDDT.Clear()
	case UDT:
		asset.StackUDT.Clear()
	case MDT:
		asset.StackMDT.Clear()
	case FDT:
		asset.StackFDT.Clear()
	}

	interlacer.idxMap[itemColN] = len(interlacer.BinData)
	interlacer.BinData = append(interlacer.BinData, asset)
	return nil
}

// compileLookupTable prepares the fast-lookup slice.
func (interlacer *Interlacer) compileLookupTable() {
	maxCol := 0
	for col := range interlacer.idxMap {
		if col > maxCol {
			maxCol = col
		}
	}
	// Create cache: index=ColumnNumber, value=BinDataIndex
	// Init with -1 (meaning "ignore this column")
	interlacer.colIndexCache = make([]int, maxCol+2)
	for i := range interlacer.colIndexCache {
		interlacer.colIndexCache[i] = -1
	}
	for col, idx := range interlacer.idxMap {
		interlacer.colIndexCache[col] = idx
	}
}

// TsvToBin processing loop.
func (interlacer *Interlacer) TsvToBin() error {
	interlacer.compileLookupTable() // Build optimization table
	ex := extractor.NewExtractor(interlacer.BufSize)
	if err := ex.Start(interlacer.TsvFile); err != nil {
		return err
	}
	defer func(ex *extractor.Extractor) {
		err := ex.End()
		if err != nil {
			panic(err)
		}
	}(ex)

	var err error
	bytesRead := ex.Read()
	for bytesRead > 0 {
		for err == nil && ex.Index() < bytesRead {
			b := ex.Buffer[ex.Index()]
			if b == '\t' || b == '\n' {
				// CRITICAL: Capture column index *before* Extractor increments it
				colID := ex.Column()
				// Extractor updates the state and increments its internal column counter
				if b == '\t' {
					ex.IndexAtEndOfColumn(colID)
				} else {
					ex.IndexAtEndOfLine(colID)
				}
				// Pass the captured ID (colID), not ex.Column()
				err = interlacer.parser(ex.Row(), colID, ex.ColumnData())
				// FIXED: Adjusts the internal state of Extractor to the next column.
				ex.IndexAtColumnData()
			} else {
				ex.IndexAtColumnData()
			}
			ex.IncrementIndex()
		}
		if err != nil {
			return err
		}
		ex.CatchTruncatedColumnData()
		bytesRead = ex.ReadNext()
	}
	return nil
}

// parser processes the data. Now uses fast slice lookup.
func (interlacer *Interlacer) parser(row, column int, data string) error {
	if interlacer.HeadOut > 0 && row < interlacer.HeadOut {
		return nil
	}
	// Boundary check for lookup table
	if column >= len(interlacer.colIndexCache) {
		return nil
	}
	binIdx := interlacer.colIndexCache[column]
	if binIdx == -1 {
		return nil // ignore this column
	}
	// Use pointer to avoid copying the struct
	asset := &interlacer.BinData[binIdx]
	switch asset.DataType {
	case SDT:
		asset.StackSDT.Push(data)
	case IDT:
		return interlacer.pushI(asset, data)
	case NDT:
		return interlacer.pushN(asset, data)
	case DDT:
		return interlacer.pushD(asset, data)
	case UDT:
		return interlacer.pushU(asset, data)
	case MDT:
		return interlacer.pushM(asset, data)
	case FDT:
		return interlacer.pushF(asset, data)
	}
	return nil
}

// --- Helper Push Methods (Updated to take *BinAsset pointer directly) ---

func (interlacer *Interlacer) pushI(asset *BinAsset, data string) error {
	n, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return err
	}
	asset.StackIDT.Push(n)
	return nil
}

func (interlacer *Interlacer) pushN(asset *BinAsset, data string) error {
	n, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		return err
	}
	asset.StackNDT.Push(int32(n))
	return nil
}

func (interlacer *Interlacer) pushD(asset *BinAsset, data string) error {
	n, err := strconv.ParseUint(data, 10, 8)
	if err != nil {
		return err
	}
	asset.StackDDT.Push(uint8(n))
	return nil
}

func (interlacer *Interlacer) pushU(asset *BinAsset, data string) error {
	n, err := strconv.ParseUint(data, 10, 32)
	if err != nil {
		return err
	}
	asset.StackUDT.Push(uint32(n))
	return nil
}

func (interlacer *Interlacer) pushM(asset *BinAsset, data string) error {
	n, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return err
	}
	asset.StackMDT.Push(n)
	return nil
}

func (interlacer *Interlacer) pushF(asset *BinAsset, data string) error {
	n, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return err
	}
	asset.StackFDT.Push(n)
	return nil
}

// Save writes data to disk.
func (interlacer *Interlacer) Save() error {
	var err error
	for i := range interlacer.BinData {
		asset := &interlacer.BinData[i]
		fullPath := path.Join(interlacer.BinPath, asset.FileName)

		co := NewContainer(asset.FileName)
		switch asset.DataType {
		case SDT:
			err = co.SaveStrArr(fullPath, asset.StackSDT.Items)
		case IDT:
			err = co.SaveNumArr(fullPath, asset.StackIDT.Items)
		case NDT:
			err = co.SaveNumArr(fullPath, asset.StackNDT.Items)
		case DDT:
			err = co.SaveNumArr(fullPath, asset.StackDDT.Items)
		case UDT:
			err = co.SaveNumArr(fullPath, asset.StackUDT.Items)
		case MDT:
			err = co.SaveNumArr(fullPath, asset.StackMDT.Items)
		case FDT:
			err = co.SaveNumArr(fullPath, asset.StackFDT.Items)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
