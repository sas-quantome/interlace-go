/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/sas-quantome/interlace-go/pkg/runner"
)

// Epsilon for float64 comparisons.
const Epsilon = 0.000001

type OperatorType int8

const (
	IG OperatorType = iota // Ignore
	EQ                     // Equal
	GT                     // Greater Than
	LT                     // Less Than
	GE                     // Greater or Equal
	LE                     // Less or Equal
	NE                     // Not Equal
	IN                     // Contains (String only)
	NI                     // Not In (String only)
)

// Container represents a data structure for handling array data from binary files with metadata and searching capabilities.
type Container struct {
	Name                              string
	Path                              string
	SUK, SDK, UUK, UDK, Fit, Found    bool
	IsCat, IsSet, IsEnu, IsBit, IsNuc bool
	FindOneOnly, FromAt, Expand       bool
	Epsilon                           float64
	Len, At                           int
	Ope                               OperatorType
	DataType                          byte
	Index                             []int
	Property, Term, EnuKw             string
	Err                               error

	// Explicit clean slices, one to hold the underlying mMap.
	SArr []string
	IArr []int64
	NArr []int32
	DArr []uint8
	UArr []uint32
	MArr []uint64
	FArr []float64

	// The raw memory-mapped file bytes
	mMap []byte
}

// epoch is used to convert YYYYMMDDHHMMSS into minutes starting from 2026 required for the 8-byte magic string
// in every bin file: B I T # * * * * where # is the data type and **** is the timestamp in number of minutes.
var epoch = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// NewContainer initializes and returns a new Container, the BinFile full path is preferred.
func NewContainer(name string) *Container {
	c := &Container{
		Epsilon: Epsilon,
		Expand:  false,
		Index:   make([]int, 0),
	}

	if strings.ContainsRune(name, '/') {
		c.Name = path.Base(name)
		c.Path = path.Dir(name)
	} else {
		c.Name = name
	}

	// Determine sorting/uniqueness from prefix
	switch {
	case strings.HasPrefix(name, "suk_"):
		c.SUK = true
	case strings.HasPrefix(name, "sdk_"):
		c.SDK = true
	case strings.HasPrefix(name, "uuk_"):
		c.UUK = true
	case strings.HasPrefix(name, "udk_"):
		c.UDK = true
	}

	// Determine type from suffix (e.g., _s.bin)
	if idx := strings.LastIndex(name, "_"); idx != -1 && idx+1 < len(name) {
		c.DataType = name[idx+1]
	} else {
		c.Err = errors.New("invalid stream type in name")
	}

	return c
}

// CloseMappedArr clears memory used, e.g., use defer co.CloseMappedArr() right after co.OpenNumArr() or co.OpenStrArr().
func (co *Container) CloseMappedArr() error {
	if co.mMap != nil {
		err := syscall.Munmap(co.mMap)
		switch co.DataType {
		case 's':
			co.SArr = nil
		case 'i':
			co.IArr = nil
		case 'f':
			co.FArr = nil
		case 'm':
			co.MArr = nil
		case 'd':
			co.DArr = nil
		case 'n':
			co.NArr = nil
		case 'u':
			co.UArr = nil
		}
		return err
	}
	return nil
}

// GetDataFromFile loads bin-encoded data from a file, if dirPath is empty will use the container's Path.
func (co *Container) GetDataFromFile(dirPath string, ignoreVersion bool) {
	if co.Err != nil || co.Fit {
		return
	}

	var fullPath string
	if dirPath == "" {
		fullPath = path.Join(dirPath, co.Name)
	} else if co.Path != "" {
		fullPath = path.Join(co.Path, co.Name)
	} else {
		fullPath = co.Name
	}

	switch co.DataType {
	case SDT:
		co.Err = co.OpenStrArr(fullPath, ignoreVersion)
	case IDT, NDT, DDT, UDT, MDT, FDT:
		co.Err = co.OpenNumArr(fullPath, ignoreVersion)
	default:
		co.Err = fmt.Errorf("unknown data type: %c", co.DataType)
	}
}

// Search routes the search request to the appropriate generic handler.
func (co *Container) Search(term string, operator OperatorType, findOneOnly, fromAt bool) {
	if co.Err != nil {
		return
	}
	co.Term = term
	co.Ope = operator
	co.FindOneOnly = findOneOnly
	co.FromAt = fromAt
	co.Found = false
	co.Index = co.Index[:0] // Reuse capacity
	if !fromAt {
		co.At = 0
	}

	if co.IsNuc {
		co.Found = false
		return
	}

	// Dispatch based on Type
	switch co.DataType {
	case SDT:
		searchGeneric(co, co.SArr, term)
	case IDT:
		n, err := strconv.ParseInt(term, 10, 64)
		if err != nil {
			co.Err = err
			return
		}
		searchGeneric(co, co.IArr, n)
	case NDT:
		n, err := strconv.ParseInt(term, 10, 32)
		if err != nil {
			co.Err = err
			return
		}
		searchGeneric(co, co.NArr, int32(n))
	case DDT:
		n, err := strconv.ParseUint(term, 10, 8)
		if err != nil {
			co.Err = err
			return
		}
		searchGeneric(co, co.DArr, uint8(n))
	case UDT:
		n, err := strconv.ParseUint(term, 10, 32)
		if err != nil {
			co.Err = err
			return
		}
		searchGeneric(co, co.UArr, uint32(n))
	case MDT:
		n, err := strconv.ParseUint(term, 10, 64)
		if err != nil {
			co.Err = err
			return
		}
		searchGeneric(co, co.MArr, n)
	case FDT:
		n, err := strconv.ParseFloat(term, 64)
		if err != nil {
			co.Err = err
			return
		}
		searchFloat(co, co.FArr, n)
	}
}

// searchGeneric handles both sorted and unsorted searches for any ordered type.
func searchGeneric[T cmp.Ordered](co *Container, data []T, term T) {
	if co.SUK || co.SDK {
		searchSorted(co, data, term)
	} else {
		searchUnsorted(co, data, term)
	}
}

// searchSorted performs binary search (O(log n)).
func searchSorted[T cmp.Ordered](co *Container, data []T, term T) {
	// Binary Search
	i := sort.Search(len(data), func(i int) bool { return data[i] >= term })

	// If FromAt flag is set, we might need to ignore matches before 'At'.
	startIdx := i
	if co.FromAt && !co.SUK {
		// In sorted duplicate keys, we might have found a match *before* our current At position.
		// However, standard binary search finds the *first* occurrence.
		// If we need to resume, we must ensure we are past co.At.
		if startIdx <= co.At {
			startIdx = co.At + 1
		}
	}

	// Check if we found anything valid
	if startIdx < len(data) && data[startIdx] == term {
		// Scan forward to find all duplicates
		for j := startIdx; j < len(data); j++ {
			if data[j] != term {
				break
			}
			co.addIndex(j)
			if co.FindOneOnly {
				return
			}
		}
	}
}

// searchUnsorted performs linear search (O(n)).
func searchUnsorted[T cmp.Ordered](co *Container, data []T, term T) {
	start := 0
	if co.FromAt {
		start = co.At + 1
	}

	for i := start; i < len(data); i++ {
		match := false
		val := data[i]

		switch co.Ope {
		case EQ:
			match = val == term
		case NE:
			match = val != term
		case GT:
			match = val > term
		case LT:
			match = val < term
		case GE:
			match = val >= term
		case LE:
			match = val <= term
		// String-specific operators
		case IN, NI:
			// Generic function can't handle string-specific Contains easily without reflection or casting.
			// Handling strictly for a string type:
			if vStr, ok := any(val).(string); ok {
				tStr := any(term).(string)
				contains := strings.Contains(vStr, tStr)
				if co.Ope == IN {
					match = contains
				}
				if co.Ope == NI {
					match = !contains
				}
			}
		case IG:
			match = false
		}

		if match {
			co.addIndex(i)
			if co.FindOneOnly {
				return
			}
		}
	}
}

// IndexNuc computes the encoded sequence length at each index of the container's unsigned array (UArr).
func (co *Container) IndexNuc() {
	co.Found = true
	na := NewNucleotideEncoder()
	co.Index = make([]int, len(co.UArr))
	j := 0
	for i := range co.UArr {
		j += len(na.Decode(co.UArr[i]))
		co.Index[i] = j
	}
	co.At = j
}

// SearchNucLocation returns the Nuc index where the decoded sequence length is greater than or equal to x.
func (co *Container) SearchNucLocation(x int) int {
	return sort.Search(len(co.Index), func(i int) bool { return co.Index[i] >= x })
}

// searchFloat handles float-specific logic (epsilon).
func searchFloat(co *Container, data []float64, term float64) {
	start := 0
	if co.FromAt {
		start = co.At + 1
	}

	// Optimization: If sorted, use Binary Search for EQ/GE/LE?
	// Floats are tricky with sorting due to epsilon. We stick to linear for safety unless exact equality.
	// If the user provided sorted floats, we assume standard sort.Float64s was used.

	if (co.SUK || co.SDK) && co.Ope == EQ {
		// Binary search for approximate float? Hard.
		// Falling back to linear scan for floats to guarantee Epsilon correctness.
		// Use standard loop.
	}

	for i := start; i < len(data); i++ {
		val := data[i]
		match := false

		switch co.Ope {
		case EQ:
			match = math.Abs(val-term) < co.Epsilon
		case NE:
			match = math.Abs(val-term) >= co.Epsilon
		case GT:
			match = val > term
		case LT:
			match = val < term
		case GE:
			match = val >= term
		case LE:
			match = val <= term
		default:
			match = false
		}

		if match {
			co.addIndex(i)
			if co.FindOneOnly {
				return
			}
		}
	}
}

func (co *Container) addIndex(i int) {
	co.At = i
	co.Found = true
	co.Index = append(co.Index, i)
}

// UpdateFlags updates the state flags based on filename conventions.
func (co *Container) UpdateFlags(binFn string) {
	co.IsSet = strings.Contains(binFn, "_set_s.")

	// Check Cat
	catSuffix := "_hash_id_m.bin"
	if co.IsSet {
		catSuffix = "_hash_id_set_s.bin"
	}
	co.IsCat = strings.HasSuffix(binFn, catSuffix) && strings.HasPrefix(binFn, "udk_")

	// Check Enu
	enuSuffix := "_cv_d.bin"
	if co.IsSet {
		enuSuffix = "_cv_set_s.bin"
	}
	if strings.Contains(binFn, "udk_bit_flags_") {
		co.IsNuc = false
		co.IsBit = true
		co.IsEnu = true
		co.EnuKw = "uuk_bit_flags"
	} else if strings.HasSuffix(binFn, enuSuffix) {
		co.IsNuc = false
		co.IsBit = false
		co.IsEnu = true
		prefix := strings.TrimSuffix(binFn, enuSuffix)
		co.EnuKw = strings.Replace(prefix, "udk_", "uuk_", 1)
	} else if IsNuc(binFn) {
		co.IsBit = false
		co.IsEnu = false
		co.IsNuc = true
	} else {
		co.IsBit = false
		co.IsEnu = false
		co.IsNuc = false
	}
	co.Property = strings.TrimSuffix(binFn, ".bin")
}

func IsNuc(binFn string) bool {
	return strings.HasPrefix(binFn, "udk_nucleotide_") && strings.HasSuffix(binFn, "_u.bin")
}

// OperatorCode returns the OperatorType for the given name.
func OperatorCode(name string) OperatorType {
	switch strings.ToLower(name) {
	case "eq":
		return EQ
	case "ne":
		return NE
	case "gt":
		return GT
	case "lt":
		return LT
	case "ge":
		return GE
	case "le":
		return LE
	case "in":
		return IN
	case "ni":
		return NI
	default:
		return IG
	}
}

func (co *Container) OpenNumArr(filename string, ignoreVersion bool) error {
	if co.DataType == 's' {
		return fmt.Errorf("invalid data type for numeric array: %s", filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	// We can close the file immediately after mapping!
	// The OS keeps the memory map alive until we call Munmap.
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	size := int(fi.Size())

	// FIX: Minimum size must be 12 (8-byte header + 4-byte CRC)
	if size < 12 {
		return fmt.Errorf("file too small: %s", filename)
	}

	// 1. Ask the OS to map the file into virtual memory
	// Note: syscall.Mmap is Unix-specific (Linux/macOS).
	co.mMap, err = syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("mMap failed: %w", err)
	}

	// 2. Slice the mapped memory into regions
	// We don't read from a stream; we just look at the byte indexes!
	header := co.mMap[0:8]
	payload := co.mMap[8 : size-4]
	crcBytes := co.mMap[size-4:]

	// 3. Verify Header
	if err = co.verifyHeader(file, header, ignoreVersion); err != nil {
		return err
	}

	// 4. Validate Checksum in a single pass
	// This is blazing fast because crc32 is just scanning virtual memory.
	expectedCRC := binary.LittleEndian.Uint32(crcBytes)
	if crc32.ChecksumIEEE(payload) != expectedCRC {
		_ = syscall.Munmap(co.mMap) // clean up on failure
		return fmt.Errorf("data corruption detected")
	}

	// 5. Zero-Copy conversion from []byte to []int64
	var count int
	switch co.DataType {
	case 'i', 'f', 'm':
		count = len(payload) / 8
	case 'd':
		count = len(payload)
	case 'n', 'u':
		count = len(payload) / 4
	}

	// Go 1.20+ safe-unsafe casting
	// We take the pointer to the start of the payload bytes, cast it to an int64 pointer,
	// and tell Go to treat that memory as a []int64 slice.
	switch co.DataType {
	case 'i':
		ptr := (*int64)(unsafe.Pointer(unsafe.SliceData(payload)))
		co.IArr = unsafe.Slice(ptr, count)
		co.Len = len(co.IArr)
	case 'f':
		ptr := (*float64)(unsafe.Pointer(unsafe.SliceData(payload)))
		co.FArr = unsafe.Slice(ptr, count)
		co.Len = len(co.FArr)
	case 'm':
		ptr := (*uint64)(unsafe.Pointer(unsafe.SliceData(payload)))
		co.MArr = unsafe.Slice(ptr, count)
		co.Len = len(co.MArr)
	case 'd':
		ptr := unsafe.SliceData(payload)
		co.DArr = unsafe.Slice(ptr, count)
		co.Len = len(co.DArr)
	case 'n':
		ptr := (*int32)(unsafe.Pointer(unsafe.SliceData(payload)))
		co.NArr = unsafe.Slice(ptr, count)
		co.Len = len(co.NArr)
	case 'u':
		ptr := (*uint32)(unsafe.Pointer(unsafe.SliceData(payload)))
		co.UArr = unsafe.Slice(ptr, count)
		co.Len = len(co.UArr)
	}
	co.Fit = true

	return nil
}

func (co *Container) OpenStrArr(filename string, ignoreVersion bool) error {
	if co.DataType != 's' {
		return fmt.Errorf("invalid data type for string array: %s", filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	// We can close the file descriptor immediately after mmap-ing
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	size := int(fi.Size())

	// FIX: Minimum size must be 12 (8-byte header + 4-byte CRC)
	if size < 12 {
		return fmt.Errorf("file too small")
	}

	// 1. Map the file directly into virtual RAM
	co.mMap, err = syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("mmap failed: %w", err)
	}

	// 2. Define the payload region
	// FIX: Updated comment to reflect the 8-byte header skip
	header := co.mMap[0:8]
	payload := co.mMap[8 : size-4]

	// 3. Verify Header
	if err = co.verifyHeader(file, header, ignoreVersion); err != nil {
		return err
	}

	// 4. One-pass CRC Check across the entire payload
	storedCRC := binary.LittleEndian.Uint32(co.mMap[size-4:])
	if crc32.ChecksumIEEE(payload) != storedCRC {
		_ = syscall.Munmap(co.mMap)
		return fmt.Errorf("data corruption detected: checksum failed")
	}

	// FIX: Ensure payload is large enough to contain at least `count` and `dataSize` (4+4 = 8 bytes)
	if len(payload) < 8 {
		_ = syscall.Munmap(co.mMap)
		return fmt.Errorf("file structure corrupted: missing payload metadata")
	}

	// 5. Parse Header Data using slice arithmetic (Zero-copy)
	// payload[0:4] -> string count
	count := int32(binary.LittleEndian.Uint32(payload[0:4]))
	// 248 956 422
	if count < 0 || count > 128*1024*1024 {
		_ = syscall.Munmap(co.mMap)
		return fmt.Errorf("invalid count exceeds 128 * 1024 * 1024: %d", count)
	}

	// 6. Calculate slice boundaries
	offsetStart := 4
	offsetEnd := offsetStart + int(count)*4

	// FIX: Prevent panic when reading dataSize by checking bounds against payload size
	if len(payload) < offsetEnd+4 {
		_ = syscall.Munmap(co.mMap)
		return fmt.Errorf("file truncated: unable to read data size")
	}

	// 7. Read Data Size (immediately follows offsets)
	dataSize := binary.LittleEndian.Uint32(payload[offsetEnd : offsetEnd+4])

	bytesStart := offsetEnd + 4
	bytesEnd := bytesStart + int(dataSize)

	// 8. Final sanity check on file structure integrity
	if bytesEnd != len(payload) {
		_ = syscall.Munmap(co.mMap)
		return fmt.Errorf("file structural length mismatch")
	}

	// 9. Zero-copy map the Offsets array
	// We cast the memory region containing offsets directly into a []int32
	offsetsPtr := (*int32)(unsafe.Pointer(&payload[offsetStart]))
	offsetsView := unsafe.Slice(offsetsPtr, count)

	// 10. Map the raw characters
	rawStringBytes := payload[bytesStart:bytesEnd]

	// 11. Construct the []string array (1 Heap Allocation)
	co.SArr = make([]string, count)
	var start int32 = 0

	for i := range count {
		end := offsetsView[i]

		// FIX: Sanity check offset values to prevent panic from corrupt data
		if end < start || int(end) > len(rawStringBytes) {
			_ = syscall.Munmap(co.mMap) // cleanup on abort
			return fmt.Errorf("corrupt offset array at string index %d", i)
		}

		// Map the string directly to the OS mmap bytes
		co.SArr[i] = unsafe.String(unsafe.SliceData(rawStringBytes[start:end]), end-start)
		start = end
	}
	co.Fit = true
	co.Len = len(co.SArr)

	return nil
}

// verifyHeader reads and validates the header, protecting against type mismatch
func (co *Container) verifyHeader(file *os.File, header []byte, ignoreVersion bool) error {
	if len(header) < 8 || header[0] != 'B' || header[1] != 'I' || header[2] != 'N' {
		return fmt.Errorf("invalid header found in: %s", file.Name())
	}
	if header[3] != co.DataType {
		return fmt.Errorf("type mismatch: file contains type token [%c.bin], but reader expected [%c.bin]: %s",
			header[3], co.DataType, file.Name())
	}
	if ignoreVersion {
		return nil
	}
	var err error
	var t time.Time
	t, err = DecodeTimestamp(header[4:8])
	if err != nil {
		return err
	}
	tt := t.Format("200601021504")
	if !strings.Contains(file.Name(), tt) && !strings.Contains("20260101000000", tt) {
		return fmt.Errorf("invalid binary version: %s from: %s", tt, file.Name())
	}
	return nil
}

// writeHeader prefixes the file with the validation signature and type token: B I T # * * * *
// where # is the data type and **** is the timestamp in number of minutes.
func (co *Container) writeHeader(file *os.File) error {
	var err error
	if _, err = file.Write([]byte{'B', 'I', 'N'}); err == nil {
		if _, err = file.Write([]byte{co.DataType}); err == nil {
			_, err = file.Write(EncodeTimestamp(FileNameToTime(file.Name())))
		}
	}
	return err
}

// FileNameToTime extracts the YYYYMMDDHHMMSS from the filepath and converts back to time else defaults to epoch time.
func FileNameToTime(fileName string) time.Time {
	s := strings.Split(fileName, "/")
	for i := range s {
		if s[i] != "" && s[i][0] == '2' {
			if runner.ValidTimeTag(s[i]) {
				year, _ := strconv.Atoi(s[i][:4])
				month, _ := strconv.Atoi(s[i][4:6])
				day, _ := strconv.Atoi(s[i][6:8])
				hour, _ := strconv.Atoi(s[i][8:10])
				minute, _ := strconv.Atoi(s[i][10:12])
				second, _ := strconv.Atoi(s[i][12:14])
				return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
			}
		}
	}
	return epoch
}

// SaveNumArr writes any primitive slice to disk with an 8-byte header and a 4-byte trailing CRC32.
func (co *Container) SaveNumArr(filename string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// 1. Write the 8-byte safety header ("BIN" + Type Token + version-in-minutes)
	if err = co.writeHeader(file); err != nil {
		return err
	}

	// 2. Initialize the CRC32 IEEE hashing system
	crcTable := crc32.MakeTable(crc32.IEEE)
	hasher := crc32.New(crcTable)

	// 3. Create a MultiWriter that duplicates the stream to both the file and the hasher
	multiWriter := io.MultiWriter(file, hasher)

	// 4. Stream the raw numeric data directly through the MultiWriter
	if err = binary.Write(multiWriter, binary.LittleEndian, data); err != nil {
		return err
	}

	// 5. Extract the finalized checksum signature
	finalChecksum := hasher.Sum32()

	// 6. Append the 4-byte checksum to the very tail of the file
	return binary.Write(file, binary.LittleEndian, finalChecksum)
}

// SaveStrArr packs a string slice into a contiguous block on disk with an 8-byte header and a 4-byte trailing CRC32.
func (co *Container) SaveStrArr(filename string, strings []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	// 1. Write the 8-byte safety header ("BIN" + Type Token + version-in-minutes)
	if err = co.writeHeader(file); err != nil {
		return err
	}

	// 2. Initialize the single-pass CRC32 hashing engine
	crcTable := crc32.MakeTable(crc32.IEEE)
	hasher := crc32.New(crcTable)

	// Create a MultiWriter so everything written goes to disk AND updates the hash
	multiWriter := io.MultiWriter(file, hasher)

	count := int32(len(strings))
	offsets := make([]int32, count)

	// Calculate the index string layout boundaries
	var totalBytes int32 = 0
	for i, s := range strings {
		totalBytes += int32(len(s))
		offsets[i] = totalBytes
	}

	// 3. Write the Total Count through the hasher into the file
	if err = binary.Write(multiWriter, binary.LittleEndian, count); err != nil {
		return err
	}

	// 4. Write the Array of End-Offsets through the hasher into the file
	if err = binary.Write(multiWriter, binary.LittleEndian, offsets); err != nil {
		return err
	}

	// 5. Write the Total Data Size through the hasher into the file
	if err = binary.Write(multiWriter, binary.LittleEndian, totalBytes); err != nil {
		return err
	}

	// 6. Stream the actual raw string characters sequentially through the hasher into the file
	for _, s := range strings {
		// Evaluated natively by the compiler to prevent extra string copying
		if _, err = multiWriter.Write([]byte(s)); err != nil {
			return err
		}
	}

	// 7. Extract the final computed checksum for everything written to the multiWriter
	finalChecksum := hasher.Sum32()

	// 8. Append the 4-byte checksum to the very tail of the file
	// CRITICAL: Write directly to 'file', NOT 'multiWriter' so the checksum isn't hashed into itself!
	return binary.Write(file, binary.LittleEndian, finalChecksum)
}

// EncodeTimestamp converts a time to 4 bytes based on minutes passed since epoch
func EncodeTimestamp(t time.Time) []byte {
	duration := t.Sub(epoch)
	minutes := uint32(duration.Minutes())

	// Pack uint32 into a 4-byte temporary buffer (big-endian)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, minutes)

	return buf
}

// DecodeTimestamp converts 4 bytes back into a full time.Time object
func DecodeTimestamp(b []byte) (time.Time, error) {
	// Validate length to prevent out-of-bounds runtime panics
	if len(b) < 4 {
		return time.Time{}, errors.New("malformed magic bits: input must be at least 4 bytes")
	}

	minutes := binary.BigEndian.Uint32(b)

	// Add the minutes back to the base epoch
	return epoch.Add(time.Duration(minutes) * time.Minute), nil
}
