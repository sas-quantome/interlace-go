/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func tmpBinFile() string {
	repoPath := tester.GitRepoPath(tester.GitProjectName, tester.GitRepoName)
	return path.Join(repoPath, "test", "suk_hola_s.bin")
}

func TestNewEnumerator(t *testing.T) {
	require.NotNil(t, NewEnumerator())
}

func TestEnumerator_Create(t *testing.T) {
	e := NewEnumerator()
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	fmt.Println("Displaying in name sort order:")
	e.Print("\t")
}

func TestEnumerator_Append(t *testing.T) {
	e := NewEnumerator()
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	e.Append("rola")
	z := e.Len()
	require.Equal(t, z, e.Len())
	e.Append("rola") // should not append since is already added
	require.Equal(t, z, e.Len())
	fmt.Println("Displaying in name sort order after appending rola:")
	e.Print("\t")
	fmt.Printf("Appended rola with code: %d\n", e.GetCode("rola"))
}

func TestEnumerator_Save(t *testing.T) {
	e := NewEnumerator()
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	e.Append("rola")
	fmt.Println("Saving enumerator in enum order...")
	require.NoError(t, e.Save(tmpBinFile())) // writes on code sort order
	fmt.Printf("File created: %s\n", tmpBinFile())
	e.Print("\t")
	require.NoError(t, os.Remove(tmpBinFile()))
}

func TestEnumerator_Load(t *testing.T) {
	e := NewEnumerator()
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	require.NoError(t, e.Save(tmpBinFile()))
	n := NewEnumerator()
	require.NoError(t, n.Load(tmpBinFile())) // loads on code sort order and sorts by name
	require.Equal(t, e.Len(), n.Len())
	/*
		require.Equal(t, "bola", n.GetName(0))
		require.Equal(t, "Cola", n.GetName(1))
		require.Equal(t, "Lola", n.GetName(2))
		require.Equal(t, "sola", n.GetName(3))
		require.Equal(t, "ola", n.GetName(4))

	*/
	n.Print("\t")
	require.NoError(t, os.Remove(tmpBinFile()))
}

func TestEnumerator_GetName(t *testing.T) {
	var e Enumerator
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	name := e.GetName(4)
	fmt.Printf("Get name from code 4: %s\n", name)
}

func TestEnumerator_GetCode(t *testing.T) {
	var e Enumerator
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	code := e.GetCode("Lola")
	require.Equal(t, code, 2)
	fmt.Printf("Get the code number from Lola: %d\n", code)
}

func TestEnumerator_OrderByEnumCode(t *testing.T) {
	var e Enumerator
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	e.Print("\t")
}

func TestEnumerator_OrderByEnumName(t *testing.T) {
	var e Enumerator
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	a := []string{"bola", "Cola", "Lola", "sola", "ola"}
	for i := 0; i < e.Len(); i++ {
		require.Equal(t, a[i], e.GetName(i))
	}
}

func TestEnumerator_Print(t *testing.T) {
	var e Enumerator
	require.NoError(t, e.Create([]string{"bola", "Cola", "Lola", "sola", "ola"}))
	e.Print("|")
}
