/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestInterlacer_NewInterlacer(t *testing.T) {
	inpTsvFile := path.Join(tester.TestDataPath(), "senescence.tsv")
	binPath := TmpRunPath(t)
	interlacer := NewInterlacer(4*1024, 0, inpTsvFile, binPath)
	require.NotNil(t, interlacer)
	require.Equal(t, 0, interlacer.HeadOut)
	require.Equal(t, 4*1024, interlacer.BufSize)
	require.FileExists(t, interlacer.TsvFile)
	require.DirExists(t, interlacer.BinPath)
}

func TestInterlacer_BinDataIndex(t *testing.T) {
	interlacer := NewInterlacer(4*1024, 0, path.Join(tester.TestDataPath(), "senescence.tsv"), TmpRunPath(t))
	require.NotNil(t, interlacer)
	require.NoError(t, interlacer.AddBinAsset(0, "sdk_entrez_gene_id_u.bin", IDT, SDK))
	require.NoError(t, interlacer.AddBinAsset(1, "udk_gene_symbol_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(3, "udk_is_cancer_cell_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(4, "udk_senescence_type_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(6, "udk_pubmed_id_u.bin", IDT, UDK))
	interlacer.compileLookupTable()
	require.Equal(t, 5, len(interlacer.BinData))
	require.Equal(t, 0, interlacer.colIndexCache[0])
	require.Equal(t, 1, interlacer.colIndexCache[1])
	require.Equal(t, 2, interlacer.colIndexCache[3])
	require.Equal(t, 3, interlacer.colIndexCache[4])
	require.Equal(t, 4, interlacer.colIndexCache[6])
	require.NotEqual(t, UUK, SUK) // to silence
}

func TestInterlacer_AddBinAsset(t *testing.T) {
	interlacer := NewInterlacer(4*1024, 0, path.Join(tester.TestDataPath(), "senescence.tsv"), TmpRunPath(t))
	require.NotNil(t, interlacer)
	require.NoError(t, interlacer.AddBinAsset(0, "sdk_entrez_gene_id_u.bin", IDT, SDK))
	require.NoError(t, interlacer.AddBinAsset(1, "udk_gene_symbol_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(2, "udk_gene_name_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(3, "udk_is_cancer_cell_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(4, "udk_senescence_type_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(5, "udk_senescence_effect_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(6, "udk_pubmed_id_u.bin", IDT, UDK))
	require.Equal(t, 7, len(interlacer.BinData))
	interlacer.compileLookupTable()
	for i := range 7 {
		if idx := interlacer.colIndexCache[i]; idx >= 0 {
			require.Equal(t, i, idx)
		}
	}
	require.Equal(t, IDT, interlacer.BinData[interlacer.colIndexCache[0]].DataType)
	for i := 1; i < 6; i++ {
		require.Equal(t, SDT, interlacer.BinData[interlacer.colIndexCache[i]].DataType)
	}
	require.Equal(t, IDT, interlacer.BinData[interlacer.colIndexCache[6]].DataType)
	require.Equal(t, SDK, interlacer.BinData[interlacer.colIndexCache[0]].IdxOrder)
	for i := 1; i < 7; i++ {
		require.Equal(t, UDK, interlacer.BinData[interlacer.colIndexCache[i]].IdxOrder)
	}
}

func TestInterlacer_TsvToBin(t *testing.T) {
	interlacer := createInterlacer(t)
	require.NoError(t, interlacer.TsvToBin())
	n := len(interlacer.BinData[interlacer.colIndexCache[0]].StackIDT.Items)
	require.Greater(t, n, 42)
	require.Equal(t, n, len(interlacer.BinData[interlacer.colIndexCache[1]].StackSDT.Items))
	require.Equal(t, n, len(interlacer.BinData[interlacer.colIndexCache[6]].StackIDT.Items))
	require.Equal(t, int64(7490), interlacer.BinData[interlacer.colIndexCache[0]].StackIDT.Items[42])
	require.Equal(t, "WT1", interlacer.BinData[interlacer.colIndexCache[1]].StackSDT.Items[42])
	require.Equal(t, int64(20972333), interlacer.BinData[interlacer.colIndexCache[6]].StackIDT.Items[42])
}

func TestInterlacer_Save(t *testing.T) {
	interlacer := createInterlacer(t)
	require.NoError(t, interlacer.TsvToBin())
	require.NoError(t, interlacer.Save())
	require.FileExists(t, path.Join(interlacer.BinPath, "sdk_entrez_gene_id_u.bin"))
	require.FileExists(t, path.Join(interlacer.BinPath, "udk_gene_symbol_s.bin"))
	require.FileExists(t, path.Join(interlacer.BinPath, "udk_pubmed_id_u.bin"))
}

func createInterlacer(t *testing.T) *Interlacer {
	interlacer := NewInterlacer(4*1024, 0, path.Join(tester.TestDataPath(), "senescence.tsv"), TmpRunPath(t))
	require.NotNil(t, interlacer)
	require.NoError(t, interlacer.AddBinAsset(0, "sdk_entrez_gene_id_u.bin", IDT, SDK))
	require.NoError(t, interlacer.AddBinAsset(1, "udk_gene_symbol_s.bin", SDT, UDK))
	require.NoError(t, interlacer.AddBinAsset(6, "udk_pubmed_id_u.bin", IDT, UDK))
	require.NoError(t, interlacer.TsvToBin())
	return interlacer
}
