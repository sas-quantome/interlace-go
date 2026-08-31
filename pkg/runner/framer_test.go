/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"os"
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestFramer_NewFramer(t *testing.T) {
	f := NewFramer()
	require.NotNil(t, f)
	require.Equal(t, 0, f.Records())
}

func TestFramer_Insert(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", "value1")
	require.Equal(t, 1, f.Records())
	require.Equal(t, "value1", f.SearchSlot("key1"))
}

func TestFramer_Delete(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", "value1")
	f.Insert("key2", "value2")
	require.Equal(t, 2, f.Records())
	f.Delete("key1")
	require.Equal(t, 1, f.Records())
	f.Delete("key1")
	require.Equal(t, 1, f.Records())
}

func TestFramer_Clear(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", "value1")
	f.Insert("key2", "value2")
	f.Insert("key3", "value3")
	require.Equal(t, 3, f.Records())
	f.Clear()
	require.Equal(t, 0, f.Records())
}

func TestFramer_InsertInt(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", 123)
	require.Equal(t, 1, f.Records())
	require.Equal(t, 123, f.SearchSlot("key1"))
}

func TestFramer_InsertMix(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", 123)
	f.Insert("key2", "abc")
	require.Equal(t, 2, f.Records())
	require.Equal(t, 123, f.SearchSlot("key1"))
	require.Equal(t, "abc", f.SearchSlot("key2"))
	result1 := f.SearchSlot("key1")
	switch result1.(type) {
	case int:
		require.Equal(t, 123, f.SearchSlot("key1"))
	case string:
		require.Equal(t, "abc", f.SearchSlot("key2"))
	default:
		require.Fail(t, "unexpected type")
	}
	result2 := f.SearchSlot("key2")
	switch result2.(type) {
	case int:
		require.Equal(t, 123, f.SearchSlot("key1"))
	case string:
		require.Equal(t, "abc", f.SearchSlot("key2"))
	default:
		require.Fail(t, "unexpected type")
	}
}

func TestFramer_SearchSlot(t *testing.T) {
	f := NewFramer()
	f.Insert("key1", "value1")
	require.Equal(t, "value1", f.SearchSlot("key1"))
	require.Equal(t, nil, f.SearchSlot("key2"))
}

func TestFramer_SlotContains(t *testing.T) {
	f := NewFramer()
	f.Insert("key1_long", "value1")
	f.Insert("key9_long", "value1")
	require.Equal(t, 1, len(f.SlotContains("key1")))
	require.Equal(t, 2, len(f.SlotContains("key")))
}

func TestFramer_LoadFrame(t *testing.T) {
	tmpDir := TmpRunPath(t)
	filePath := path.Join(tmpDir, "test_frame.tsv")
	content := "key1\tvalue1\nkey2\tvalue2\n"
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	f := NewFramer()
	err = f.LoadFrame(filePath)
	require.NoError(t, err)
	require.Equal(t, 2, f.Records())
	require.Equal(t, "value1", f.SearchSlot("key1"))
	require.Equal(t, "value2", f.SearchSlot("key2"))
}

func TestFramer_LoadFrameFromTable(t *testing.T) {
	tmpDir := TmpRunPath(t)
	filePath := path.Join(tmpDir, "test_frame_custom.tsv")
	content := "col0\tkey1\tvalue1\ncol0\tkey2\tvalue2\n"
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	f := NewFramer()
	err = f.LoadFrameFromTable(filePath, 1, 2, 0)
	require.NoError(t, err)
	require.Equal(t, 2, f.Records())
	require.Equal(t, "value1", f.SearchSlot("key1"))
	require.Equal(t, "value2", f.SearchSlot("key2"))
}

func TestFramer_PadWrite(t *testing.T) {
	configFile := path.Join(tester.TestDataPath(), "config.tsv")
	framer := NewFramer()
	require.NoError(t, framer.LoadFrame(configFile))
	require.NoError(t, framer.PadWrite(os.Stdout))
}

func TestFramer_TabWrite(t *testing.T) {
	configFile := path.Join(tester.TestDataPath(), "config.tsv")
	framer := NewFramer()
	require.NoError(t, framer.LoadFrame(configFile))
	require.NoError(t, framer.TabWrite(os.Stdout, "|"))
}
