/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package randommer

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

// RandomDialCode returns random dial code from code map.
func RandomDialCode(codes map[int]string) int {
	dialCode := 0
	n := RandomInt(1, int64(len(codes)))
	for dialCode = range codes {
		n--
		if n == 0 {
			break
		}
	}
	return dialCode
}

// RandomPhoneEntry returns an 11-digit random phone number.
func RandomPhoneEntry(dialCode, dialCodeLen int) string {
	return fmt.Sprintf("%d-%s", dialCode, RandomStringNumber(11-dialCodeLen))
}

func RandomDNA(n int, lowComplexity bool) string {
	if n <= 0 {
		return ""
	}
	const alpha = "ACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTA"
	const runoff = "aaacccgggttt"
	var sb strings.Builder

	for range n {
		if lowComplexity {
			sb.WriteByte(runoff[RandomInt(0, int64(len(runoff)-1))])
		} else {
			sb.WriteByte(alpha[RandomInt(0, int64(len(alpha)-1))])
		}
	}

	return sb.String()
}

// RandomInt generates a random positive integer between min and max or returns -1 if it fails.
func RandomInt(min, max int64) int64 {
	n := max - min + 1
	if n <= 0 {
		return -1
	}
	return min + rand.Int63n(n)
}

// RandomString generates a random string of length n or returns empty string if fails.
func RandomString(n int) string {
	if n <= 0 {
		return ""
	}
	const alpha = "bcdfghjklmnpqrstvwxyz"
	const aeiou = "aeiou"
	var sb strings.Builder
	k := len(alpha)

	flag := true
	for i := 0; i <= n; i++ {
		if flag {
			flag = false
			sb.WriteByte(alpha[rand.Intn(k)])
		} else {
			flag = true
			sb.WriteByte(aeiou[rand.Intn(5)])
		}
	}

	return sb.String()
}

// RandomStringNumber generates a random string of digits of length n or returns empty string if fails.
func RandomStringNumber(n int) string {
	if n <= 0 {
		return ""
	}
	const alphabet = "0123456789"
	var sb strings.Builder
	k := len(alphabet)

	for range n {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomAlias generates a random alias name.
func RandomAlias() string {
	return RandomString(6)
}

// RandomFKey returns fk or generates fk.
func RandomFKey(fk string, sz int) string {
	if fk != "" {
		return fk
	}
	return strings.ToUpper(RandomString(sz))
}

// RandomForenames generates a random forename.
func RandomForenames() string {
	one := RandomString(int(RandomInt(2, 6)))
	two := RandomString(int(RandomInt(2, 6)))
	a := byte(unicode.ToUpper(rune(one[0])))
	b := byte(unicode.ToUpper(rune(two[0])))
	return string(a) + one[1:] + " " + string(b) + two[1:]
}

// RandomSurname generates a random surname.
func RandomSurname() string {
	one := RandomString(int(RandomInt(2, 6)))
	a := byte(unicode.ToUpper(rune(one[0])))
	return string(a) + one[1:]
}

// RandomEmail generates a random email
func RandomEmail() string {
	return CreateEmail(RandomForenames(), RandomSurname(), RandomString(int(RandomInt(3, 7))))
}

// CreateEmail generates email from a subscription.
func CreateEmail(forenames, surname, organizationBrand string) string {
	x := []string{".com", ".io", ".xyz", ".edu", ".org", ".fr", ".jp", ".us", ".ca", ".biz"}
	n := len(x)
	s := strings.ToLower(strings.Split(forenames, " ")[0])
	s += "."
	s += strings.ToLower(surname)
	s += "@"
	s += strings.ToLower(organizationBrand)
	s += x[rand.Intn(n)]
	return s
}

// RandomTitle returns nonsense strings as a title.
func RandomTitle() string {
	var title string
	n := int(RandomInt(2, 4))
	for i := range n {
		if i == 0 {
			title = RandomString(int(RandomInt(3, 7)))
		} else {
			title = fmt.Sprintf("%s %s", title, RandomString(int(RandomInt(3, 7))))
		}
	}
	return title
}

// RandomText returns nonsense text.
func RandomText() string {
	var text string
	n := int(RandomInt(7, 13))
	for i := range n {
		if i == 0 {
			text = RandomString(int(RandomInt(3, 7)))
		} else {
			text = fmt.Sprintf("%s %s", text, RandomString(int(RandomInt(3, 7))))
		}
	}
	return text
}

// RandomMonth returns time month.
func RandomMonth() time.Month {
	switch int(RandomInt(1, 12)) {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	default:
		return time.December
	}
}

// RandomDate returns random date.
func RandomDate(year int, month time.Month) time.Time {
	return time.Date(year,
		month,
		int(RandomInt(1, 8)),
		int(RandomInt(0, 24)),
		int(RandomInt(0, 60)),
		int(RandomInt(0, 60)),
		int(RandomInt(0, 3600)),
		time.UTC)
}

// RandomYearRange returns a random year range at min year distance of separation.
func RandomYearRange(maxIntervalSize int) (int, int) {
	td := max(maxIntervalSize, 1)
	t0 := time.Now()
	t1 := int64(t0.Year())
	t2 := t1 + int64(td)
	y1 := RandomInt(t1, t2)
	y2 := RandomInt(y1, t2)
	return int(y1), int(y2)
}

// RandomMonthRange returns month one and two where month one is the earliest.
func RandomMonthRange() (time.Month, time.Month) {
	m1 := RandomMonth()
	m2 := RandomMonth()
	if m1 >= m2 {
		m0 := m1
		m1 = m2
		m2 = m0
	}
	return m1, m2
}
