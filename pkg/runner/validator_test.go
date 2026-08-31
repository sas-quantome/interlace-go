/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestValidator_CheckAllAreDigits(t *testing.T) {
	require.Equal(t, 0, CheckAllAreDigits("1abc"))
	require.Equal(t, 3, CheckAllAreDigits("123"))
}

func TestValidator_CheckAllAreUpperCased(t *testing.T) {
	require.Equal(t, 0, CheckAllAreUpperCased("1aBc"))
	require.Equal(t, 0, CheckAllAreUpperCased("AB3"))
	require.Equal(t, 3, CheckAllAreUpperCased("ABC"))
}

func TestValidator_CheckAllAreLowerCased(t *testing.T) {
	require.Equal(t, 0, CheckAllAreLowerCased("1aBc"))
	require.Equal(t, 0, CheckAllAreLowerCased("ab3"))
	require.Equal(t, 3, CheckAllAreLowerCased("abc"))
}

func TestValidator_IsGzipped(t *testing.T) {
	s := "This is a test string."
	gz, err := StringToGzipBytes(s)
	require.NoError(t, err)
	require.True(t, IsGzipped(gz))
	require.False(t, IsGzipped([]byte(s)))
}

func TestValidator_ValidTimeTag(t *testing.T) {
	require.Equal(t, true, ValidTimeTag(ApolloTimeTag))
}

func TestValidator_SearchLatestTimeTag(t *testing.T) {
	require.Equal(t, ApolloTimeTag, SearchLatestTimeTag(tester.TestTimeTags()))
}

func TestValidator_ValidAppID(t *testing.T) {
	ids := []string{"a10101", "t10102", "r12345", "w54321", "s20102", "b01011", "d12300"}
	for i := range ids {
		require.True(t, ValidAppID(ids[i]), ids[i])
	}
	ids = []string{" ", "ahd", "", "0", "s123456", "x12345"}
	for i := range ids {
		require.False(t, ValidAppID(ids[i]), ids[i])
	}
}

func TestValidator_ValidProjectID(t *testing.T) {
	ids := []string{"p10101", "p10102", "p12345"}
	for i := range ids {
		require.True(t, ValidProjectID(ids[i]), ids[i])
	}
	ids = []string{" ", "ahd", "", "0", "p123456", "x12345"}
	for i := range ids {
		require.False(t, ValidProjectID(ids[i]), ids[i])
	}
}

func TestValidator_ValidateManager(t *testing.T) {
	require.NoError(t, ValidateManager("123", ApolloTimeTag))
	require.Error(t, ValidateManager("12", ApolloTimeTag))
	require.Error(t, ValidateManager("123", "30260720201740"))
}

func TestValidator_ValidateDispenser(t *testing.T) {
	require.NoError(t, ValidateDispenser("123", ApolloTimeTag, "123", false))
	require.Error(t, ValidateDispenser("12", ApolloTimeTag, "123", false))
	require.Error(t, ValidateDispenser("123", ApolloTimeTag, "12", false))
	require.NoError(t, ValidateDispenser("123", ApolloTimeTag, "w10101", true))
	require.NoError(t, ValidateDispenser("123", ApolloTimeTag, "a10101", true))
	require.NoError(t, ValidateDispenser("123", ApolloTimeTag, "b10101", true))
	require.Error(t, ValidateDispenser("123", ApolloTimeTag, "w123", true))
	require.Error(t, ValidateDispenser("123", ApolloTimeTag, "123456", true))
}

func TestValidator_IsNumeric(t *testing.T) {
	require.True(t, IsNumeric("123"))
	require.True(t, IsNumeric("123.4"))
	require.False(t, IsNumeric("123_4"))
	require.False(t, IsNumeric("123a"))
}

func TestValidator_ValidateSame(t *testing.T) {
	require.True(t, ValidateSame(1, 1, 1, 1))
	require.False(t, ValidateSame(1, 2, 1, 2))
}

func TestValidator_ExtractTimeTag(t *testing.T) {
	require.Equal(t, ApolloTimeTag, ExtractTimeTag(path.Join("aaa", "bbb", ApolloTimeTag, "ccc")))
}
