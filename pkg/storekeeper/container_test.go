/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestContainer_NewContainer(t *testing.T) {
	r := NewContainer("sdk_synonym_s.bin")
	require.NotNil(t, r)
	require.Equal(t, "sdk_synonym_s.bin", r.Name)
	require.True(t, r.SDK)
	require.Equal(t, SDT, r.DataType)
}

func TestContainer_OperatorCode(t *testing.T) {
	require.Equal(t, EQ, OperatorCode("eq"))
	require.Equal(t, NE, OperatorCode("ne"))
	require.Equal(t, GT, OperatorCode("gt"))
	require.Equal(t, IG, OperatorCode("unknown"))
}

func TestContainer_Search(t *testing.T) {
	gobPath := TmpRunPath(t)
	interlacer := NewInterlacer(4*1024, 0, path.Join(tester.TestDataPath(), "senescence.tsv"), gobPath)
	require.NotNil(t, interlacer)
	require.NoError(t, interlacer.AddBinAsset(0, "sdk_entrez_gene_id_u.bin", UDT, SDK))
	require.NoError(t, interlacer.AddBinAsset(1, "udk_gene_symbol_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(3, "udk_is_cancer_cell_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(6, "udk_pubmed_id_u.bin", UDT, UDK))
	require.NoError(t, interlacer.TsvToBin())
	require.NoError(t, interlacer.Save())
	assets := make([]*Container, len(interlacer.BinData))
	idxMap := make(map[string]int)
	fmt.Println()
	for i := range interlacer.BinData {
		idxMap[interlacer.BinData[i].FileName] = i
		fmt.Printf("%s %s\n", gobPath, interlacer.BinData[i].FileName)
		assets[i] = NewContainer(path.Join(gobPath, interlacer.BinData[i].FileName))
		assets[i].GetDataFromFile(gobPath, true)
		require.NoError(t, assets[i].Err)
		fmt.Println("Property:", assets[i].Len, "rows:", assets[i].Name)
	}
	fmt.Println()
	search(t, assets, idxMap["sdk_entrez_gene_id_u.bin"], 3, uint32(7157), false, false)
	search(t, assets, idxMap["udk_gene_symbol_s.bin"], 2, "UTP6", false, false)
	search(t, assets, idxMap["udk_pubmed_id_u.bin"], 1, uint32(28394343), false, false)
	search(t, assets, idxMap["udk_is_cancer_cell_s.bin"], 1, "Yes", true, false)
	search(t, assets, idxMap["udk_is_cancer_cell_s.bin"], 1, "Yes", true, true)
	search(t, assets, idxMap["sdk_entrez_gene_id_u.bin"], 1, uint32(7157), true, false)
	search(t, assets, idxMap["sdk_entrez_gene_id_u.bin"], 1, uint32(7157), true, true)
	search(t, assets, idxMap["sdk_entrez_gene_id_u.bin"], 1, uint32(7157), true, true)
}

func search(t *testing.T, assets []*Container, qryIdx, expected int, term any, findOneOnly, fromAt bool) {
	assets[qryIdx].Term = term
	assets[qryIdx].SearchAny(term, EQ, findOneOnly, fromAt)
	fmt.Println("Finding At Index:", assets[qryIdx].Index)
	nFound := len(assets[qryIdx].Index)
	require.Equal(t, expected, nFound)
	for i := range nFound {
		for property := range assets {
			j := assets[qryIdx].Index[i]
			switch assets[property].DataType {
			case SDT:
				fmt.Print(assets[property].SArr[j], "\t")
			case IDT:
				fmt.Print(assets[property].IArr[j], "\t")
			case NDT:
				fmt.Print(assets[property].NArr[j], "\t")
			case DDT:
				fmt.Print(assets[property].DArr[j], "\t")
			case UDT:
				fmt.Print(assets[property].UArr[j], "\t")
			case MDT:
				fmt.Print(assets[property].MArr[j], "\t")
			case FDT:
				fmt.Print(assets[property].FArr[j], "\t")
			}
		}
		fmt.Print("\n\n")
	}
}

// Helper function to create randomized test data
func generateTestData() ([]int64, []string) {
	ints := make([]int64, 1000)
	strings := make([]string, 1000)
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	for i := range 1000 {
		ints[i] = rand.Int63()

		// Make a random 10-character string
		b := make([]rune, 10)
		for j := range b {
			b[j] = letters[rand.Intn(len(letters))]
		}
		strings[i] = string(b)
	}
	return ints, strings
}

func binFp() string {
	return "interlace-ex/entity/gene/function/20260807000009/property/sdk_entrez_gene_id_i.bin"
}

func TestFileNameToTime(t *testing.T) {
	x := FileNameToTime(binFp())
	// cut the seconds out
	require.Equal(t, "20260807000009", x.Format("20060102150405"))
	fmt.Printf("Encoded Magic Bits (Hex): %s %s\n", x.String(), x.Format("20060102150405"))
}

func TestEncodeTimestamp(t *testing.T) {
	b := EncodeTimestamp(FileNameToTime(binFp()))
	fmt.Printf("Encoded Magic Bits (Hex): %X %v\n", b, b)
}

func TestDecodeTimestamp(t *testing.T) {
	y, e := DecodeTimestamp(EncodeTimestamp(FileNameToTime(binFp())))
	require.NoError(t, e)
	fmt.Printf("Encoded Magic Bits (Hex) with TimeTag in path: %s %s\n", y.String(), y.Format("20060102150405"))
	// cut the seconds out
	require.Equal(t, "20260807000000", y.Format("20060102150405"))
	y, e = DecodeTimestamp(EncodeTimestamp(FileNameToTime("aa/bb/cc")))
	require.NoError(t, e)
	fmt.Printf("Encoded Magic Bits (Hex) without TimeTag in path: %s %s\n", y.String(), y.Format("20060102150405"))
	// cut the seconds out
	require.Equal(t, "20260101000000", y.Format("20060102150405"))
}

// ============================================================================
// INTEGRITY & CORRECTNESS TESTS
// ============================================================================

func TestSaveAndReadStrings(t *testing.T) {
	filename := "udk_test_s.bin"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.FailNow()
		}
	}(filename)

	_, originalData := generateTestData()
	co := NewContainer(filename)

	// 1. Save data
	if err := co.SaveStrArr(filename, originalData); err != nil {
		t.Fatalf("Failed to save string array: %v", err)
	}

	// 2. Read data back
	if err := co.OpenStrArr(filename, true); err != nil {
		t.Fatalf("Failed to read string array: %v", err)
	}
	defer func() {
		_ = co.CloseMappedArr()
	}()

	// 3. Verify strings match perfectly
	if co.Len != len(originalData) {
		t.Fatalf("Length mismatch! Expected %d, got %d", len(originalData), co.Len)
	}
	for i := range originalData {
		if co.SArr[i] != originalData[i] {
			t.Errorf("String corruption at index %d! Expected %q, got %q", i, originalData[i], co.SArr[i])
		}
	}
}

func TestSaveAndReadInt64(t *testing.T) {
	filename := "udk_test_i.bin"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.FailNow()
		}
	}(filename)

	originalData, _ := generateTestData()
	co := NewContainer(filename)

	// 1. Write
	if err := co.SaveNumArr(filename, originalData); err != nil {
		t.Fatalf("Failed to save int64 array: %v", err)
	}

	// 2. Read data back using memory pools
	if err := co.OpenNumArr(filename, false); err != nil {
		t.Fatalf("Failed to read int64 array: %v", err)
	}
	defer func() {
		_ = co.CloseMappedArr()
	}()

	// 3. Verify values match perfectly
	if co.Len != len(originalData) {
		t.Fatalf("Length mismatch! Expected %d, got %d", len(originalData), co.Len)
	}
	for i := range originalData {
		if co.IArr[i] != originalData[i] {
			t.Errorf("Data corruption at index %d! Expected %d, got %d", i, originalData[i], co.IArr[i])
		}
	}
}

func TestCorruptionDetection(t *testing.T) {
	filename := "suk_corrupted_i.bin"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.FailNow()
		}
	}(filename)

	originalData := []int64{1, 2, 3, 4, 5}
	co := NewContainer(filename)
	// CORRECTED: Call SaveNumericArray directly with the type token
	if err := co.SaveNumArr(filename, originalData); err != nil {
		t.Fatalf("Failed to save int64 array: %v", err)
	}

	// Open file and intentionally alter the very last byte (the CRC32 signature area)
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for tampering: %v", err)
	}
	fi, _ := file.Stat()

	// Corrupt the trailing checksum byte
	_, _ = file.WriteAt([]byte{0xFF}, fi.Size()-1)
	if file.Close() != nil {
		t.FailNow()
	}

	// Try reading it back; it must trigger our CRC verification failure
	err = co.OpenNumArr(filename, false)
	defer func() {
		_ = co.CloseMappedArr()
	}()
	if err == nil {
		t.Error("Security flaw: Reader successfully processed a corrupted file without throwing an error!")
	}
}

// ============================================================================
// PERFORMANCE BENCHMARKS (Gob vs Custom Binary)
// ============================================================================

func BenchmarkSaveCustomBinary(b *testing.B) {
	originalData, _ := generateTestData()
	filename := "udk_bench_i.bin"
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			b.FailNow()
		}
	}(filename)
	b.ResetTimer()
	co := NewContainer(filename)
	for i := 0; i < b.N; i++ {
		_ = co.SaveNumArr(filename, originalData)
	}
}

// BenchmarkReadGob-10    	                       67294	     16070 ns/op
// BenchmarkReadCustomBinaryZeroAlloc-10    	   22879	     51467 ns/op
func BenchmarkReadCustomBinaryZeroAlloc(b *testing.B) {
	originalData, _ := generateTestData()
	filename := "sdk_bench_i.bin"
	// CORRECTED: Call SaveNumericArray directly with the type token
	co := NewContainer(filename)
	_ = co.SaveNumArr(filename, originalData)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			b.FailNow()
		}
	}(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = co.OpenNumArr(filename, false)
		_ = co.CloseMappedArr()
	}
}

func BenchmarkSaveGob(b *testing.B) {
	originalData, _ := generateTestData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		_ = enc.Encode(originalData)
	}
}

func BenchmarkReadGob(b *testing.B) {
	originalData, _ := generateTestData()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(originalData)
	gobBytes := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target []int64
		reader := bytes.NewReader(gobBytes)
		dec := gob.NewDecoder(reader)
		_ = dec.Decode(&target)
	}
}
