/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package randommer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	x1 := RandomInt(0, 1024) // returns random number >= 0
	var x2 int64 = -2
	require.Greater(t, x1, x2, "max-min+1 <= 0")
	x1 = RandomInt(1024, 0) // fails and returns -1
	x2 = 0
	require.Greater(t, x2, x1, "max-min+1 <= 0")
}

func TestRandomDNA(t *testing.T) {
	high := RandomDNA(100, false)
	low := RandomDNA(100, true)
	require.Equal(t, 100, len(high))
	require.Equal(t, 100, len(low))
	t.Logf("High complexity random DNA:\t%s", high)
	t.Logf("Low complexity random DNA:\t%s", low)
}

func TestRandomString(t *testing.T) {
	require.NotEmpty(t, RandomString(1), "failed creating random string size 1")
	require.NotEmpty(t, RandomString(64), "failed creating random string size 1024")
	require.Empty(t, RandomString(0), "creating random empty string should not be possible")
	require.Empty(t, RandomString(-2), "creating random negative size string should not be possible")
	t.Logf("random string: %s", RandomString(6))
}

func TestRandomTitle(t *testing.T) {
	require.NotEmpty(t, RandomTitle())
}

func TestRandomText(t *testing.T) {
	require.NotEmpty(t, RandomText())
}

func TestRandomMonth(t *testing.T) {
	require.NotEmpty(t, RandomMonth())
}

func TestRandomDate(t *testing.T) {
	require.NotEmpty(t, RandomDate(2022, 8))
}

func TestRandomStringNumber(t *testing.T) {
	require.NotEmpty(t, RandomStringNumber(1), "failed creating random string size 1")
	require.NotEmpty(t, RandomStringNumber(64), "failed creating random string size 1024")
	require.Empty(t, RandomStringNumber(0), "creating random empty string should not be possible")
	require.Empty(t, RandomStringNumber(-2), "creating random negative size string should not be possible")
	t.Logf("random number: %s", RandomStringNumber(11))
}

func TestRandomAlias(t *testing.T) {
	s := RandomAlias()
	require.NotEmpty(t, s, "failed creating random alias")
	t.Logf("random alias: %s", s)
}

func TestRandomFKey(t *testing.T) {
	s := RandomFKey("", 6)
	require.NotEmpty(t, s, "failed creating random FK")
	t.Logf("random FK: %s", s)
	s = RandomFKey("NFTme", 6)
	require.NotEmpty(t, s, "failed returning FK")
	t.Logf("non random FK: %s", s)
}

func TestRandomForenames(t *testing.T) {
	s := RandomForenames()
	require.NotEmpty(t, s, "failed creating random forenames")
	t.Logf("random forenames: %s", s)
}

func TestRandomSurname(t *testing.T) {
	s := RandomSurname()
	require.NotEmpty(t, s, "failed creating random surname")
	t.Logf("random forename: %s", s)
}

func TestRandomEmail(t *testing.T) {
	s := RandomEmail()
	require.NotEmpty(t, s, "failed creating random email")
	t.Logf("random email: %s", s)
}

func TestCreateEmail(t *testing.T) {
	brands := []string{"genofy", "telomenhancer", "cryptoomics", "nftme", "genexpat"}
	for i := range brands {
		email := CreateEmail(RandomForenames(), RandomSurname(), brands[i])
		require.NotEmpty(t, email, "failed creating email")
		t.Logf("random email: %s", email)
	}
}

func TestRandomYearRange(t *testing.T) {
	y1, y2 := RandomYearRange(4)
	require.LessOrEqual(t, y1, y2)
}

func TestRandomMonthRange(t *testing.T) {
	m1, m2 := RandomMonthRange()
	require.LessOrEqual(t, m1, m2)
}
