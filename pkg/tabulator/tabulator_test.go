/*
 * Copyright 2026 SAS Quantome
 * SPDX-License-Identifier: Apache-2.0
 */

package tabulator

import (
	"os"
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestNewTabulator(t *testing.T) {
	tabulator := NewTabulator()
	require.NotNil(t, tabulator)
	require.NotNil(t, tabulator.Data)
}

func TestTabulator_LoadTable(t *testing.T) {
	sampleTsv := path.Join(tester.TestDataPath(), "bitcoin.tsv")
	tabulator := NewTabulator()
	err := tabulator.LoadTable(sampleTsv)
	require.NoError(t, err)
	require.Equal(t, tabulator.Columns(), len(tabulator.Data))
	require.Equal(t, tabulator.Rows, len(tabulator.Data[0]))
}

func TestTabulator_TabWriteTableSearch(t *testing.T) {
	sampleTsv := path.Join(tester.TestDataPath(), "bitcoin.tsv")
	tabulator := NewTabulator()
	err := tabulator.LoadTable(sampleTsv)
	require.NoError(t, err)
	require.NoError(t, tabulator.TabWriteTableSearch(os.Stdout, "Year", "2023", "|", ExactMatch))
	require.NoError(t, tabulator.TabWriteTableSearch(os.Stdout, "Year", "201", "|", PrefixMatch))
}

func TestTabulator_PadWriteTableSearch(t *testing.T) {
	sampleTsv := path.Join(tester.TestDataPath(), "bitcoin.tsv")
	tabulator := NewTabulator()
	err := tabulator.LoadTable(sampleTsv)
	require.NoError(t, err)
	require.NoError(t, tabulator.PadWriteTableSearch(os.Stdout, "Year", "2023", ExactMatch))
	require.NoError(t, tabulator.PadWriteTableSearch(os.Stdout, "Year", "201", PrefixMatch))
}

func TestTabulator_SearchTable(t *testing.T) {
	sampleTsv := path.Join(tester.TestDataPath(), "bitcoin.tsv")
	tabulator := NewTabulator()
	err := tabulator.LoadTable(sampleTsv)
	require.NoError(t, err)
	require.Equal(t, "27100", tabulator.SearchTable("Year", "High", "2023"))
}

func TestTabulator_SearchColumn(t *testing.T) {
	sampleTsv := path.Join(tester.TestDataPath(), "bitcoin.tsv")
	tabulator := NewTabulator()
	err := tabulator.LoadTable(sampleTsv)
	require.NoError(t, err)
	require.Equal(t, -1, tabulator.SearchColumn(""))
	require.Equal(t, -1, tabulator.SearchColumn("NoName"))
	require.Equal(t, 0, tabulator.SearchColumn("Year"))
	require.Equal(t, 1, tabulator.SearchColumn("Timestamp"))
	require.Equal(t, 2, tabulator.SearchColumn("Open"))
	require.Equal(t, 3, tabulator.SearchColumn("High"))
	require.Equal(t, 4, tabulator.SearchColumn("Low"))
	require.Equal(t, 5, tabulator.SearchColumn("Close"))
	require.Equal(t, 6, tabulator.SearchColumn("Volume"))
}
