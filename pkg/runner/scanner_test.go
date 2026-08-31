/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestScanner_NewScanner(t *testing.T) {
	scanner1 := NewScanner(0)
	require.NotNil(t, scanner1)
	require.Nil(t, scanner1.buf) // bufio.Scanner will set the default size (64KB)
	scanner2 := NewScanner(1024 * 1024)
	require.NotNil(t, scanner2)
	require.NotNil(t, scanner2.buf) // overwrites the default bufio.Scanner size (64KB)
}

func TestScanner_StartEnd(t *testing.T) {
	scanner := NewScanner(0)
	require.NoError(t, scanner.Start(path.Join(tester.TestDataPath(), "senescence.tsv")))
	for row := 1; scanner.Scan(); row++ {
		s := strings.Split(scanner.Text(), "\t")
		if len(s) == 7 {
			for col := range s {
				fmt.Printf("row:\t%d\tcolumn: %d\tdata: %s\n", row, col, s[col])
			}
		}
	}
	require.NoError(t, ValidateErrors(scanner.Err(), scanner.Close()))
}
