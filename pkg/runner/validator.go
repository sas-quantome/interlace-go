/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// IsGzipped checks if the given byte slice is gzipped by verifying the presence of the Gzip magic number (0x1F, 0x8B).
func IsGzipped(data []byte) bool {
	// Gzip files start with the magic bytes 0x1F, 0x8B
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x1F && data[1] == 0x8B
}

// ValidTimeTag function returns true if string is YYYYMMDDHHMMSS.
func ValidTimeTag(s string) bool {
	if CheckAllAreDigits(s) != 14 {
		return false
	}
	if s[0] != '2' && s[0] != '1' {
		return false
	}
	if i, e := strconv.Atoi(s[4:6]); e != nil || i < 1 || i > 12 {
		return false
	}
	if i, e := strconv.Atoi(s[6:8]); e != nil || i < 1 || i > 31 {
		return false
	}
	if i, e := strconv.Atoi(s[8:10]); e != nil || i < 0 || i > 23 {
		return false
	}
	if i, e := strconv.Atoi(s[10:12]); e != nil || i < 0 || i > 59 {
		return false
	}
	if i, e := strconv.Atoi(s[12:]); e != nil || i < 0 || i > 59 {
		return false
	}
	return true
}

// TimeTagNow returns a string tag formatted as [YYYYMMDDHHMMSS] with the current time.
func TimeTagNow() string {
	return time.Now().Format("20060102150405")
}

// SearchLatestTimeTag scans a directory for files with valid time tag names and returns the latest one, or an empty string if none found.
func SearchLatestTimeTag(fp string) string {
	files, err := os.ReadDir(fp)
	if err != nil {
		return ""
	}
	latestVersion := ""
	for _, file := range files {
		if ValidTimeTag(file.Name()) {
			if latestVersion < file.Name() {
				latestVersion = file.Name()
			}
		}
	}
	return latestVersion
}

// CheckAllAreDigits returns string length if all runes are digits.
func CheckAllAreDigits(s string) int {
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
	}
	return len(s)
}

// AllAreDigits returns true if all runes are digits.
func AllAreDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// CheckAllAreLowerCased returns string length if all runes are lower-cased characters.
func CheckAllAreLowerCased(s string) int {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return 0
		}
	}
	return len(s)
}

// CheckAllAreUpperCased returns string length if all runes are upper-cased characters.
func CheckAllAreUpperCased(s string) int {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return 0
		}
	}
	return len(s)
}

// IsNumeric returns true if the string contains only digits allowing a single dot and a minus sign at the start.
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}
	dots := 0
	for i := 0; i < len(s); i++ {
		if i == 0 && s[i] == '-' {
			continue
		}
		if s[i] == '.' {
			dots++
			if dots > 1 {
				return false
			}
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// MustHaveArrPrefix validates that the given string starts with one of the required prefixes: "suk", "uuk", "sdk", or "udk".
func MustHaveArrPrefix(binFn string) error {
	if !strings.HasPrefix(binFn, "suk_") && !strings.HasPrefix(binFn, "uuk_") &&
		!strings.HasPrefix(binFn, "sdk_") && !strings.HasPrefix(binFn, "udk_") {
		return errors.New("invalid array prefix")
	}
	return nil
}

// ValidAppID function returns true if string is a valid application ID such as a10101 and t10111 etc.
func ValidAppID(s string) bool {
	if len(s) != 6 {
		return false
	}
	if s[0] != 'a' && s[0] != 't' && s[0] != 'r' && s[0] != 's' && s[0] != 'b' && s[0] != 'd' && s[0] != 'w' {
		return false
	}
	if CheckAllAreDigits(strings.TrimPrefix(s, string(s[0]))) != 5 {
		return false
	}
	return true
}

// ValidProjectID function returns true if the string is a valid project ID such as p. 10101.
func ValidProjectID(s string) bool {
	if len(s) != 6 {
		return false
	}
	if s[0] != 'p' {
		return false
	}
	if CheckAllAreDigits(strings.TrimPrefix(s, string(s[0]))) != 5 {
		return false
	}
	return true
}

// ValidateManager validates the given agentName and timeTag strings.
func ValidateManager(agentName, timeTag string) error {
	if len(agentName) < 3 {
		return fmt.Errorf("agentName string length is less than 3: %s", agentName)
	}
	if ValidTimeTag(timeTag) == false {
		return fmt.Errorf("invalid time tag: %s", timeTag)
	}
	return nil
}

// ValidateDispenser validates the given concept, timeTag, and context strings.
func ValidateDispenser(entityConceptName, timeTag, assetIdentityName string, mustBeAppID bool) error {
	if len(entityConceptName) < 3 {
		return fmt.Errorf("concept string length is less than 3: %s", entityConceptName)
	}
	if mustBeAppID {
		if ValidAppID(assetIdentityName) == false {
			return fmt.Errorf("invalid WorkflowID: %s", assetIdentityName)
		}
	} else {
		if len(assetIdentityName) < 3 {
			return fmt.Errorf("asset string length is less than 3: %s", assetIdentityName)
		}
	}
	if ValidTimeTag(timeTag) == false {
		return fmt.Errorf("invalid time tag: %s", timeTag)
	}
	return nil
}

// ValidateErrors returns concatenated errors or nil if none.
func ValidateErrors(er ...error) error {
	var e []error
	for _, err := range er {
		if err != nil {
			e = append(e, err)
		}
	}
	return errors.Join(e...)
}

func ValidatePathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func ValidateSame(n ...int) bool {
	for i := 1; i < len(n); i++ {
		if n[i] != n[i-1] {
			return false
		}
	}
	return true
}

// ErrorInLine returns true if it detects one of the possible known error patterns found in log messages.
func ErrorInLine(s string) bool {
	if !strings.Contains(s, "rrors :") &&
		!strings.Contains(s, "ERROR StatusLogger") &&
		!strings.Contains(s, "-maxErrors") &&
		!strings.Contains(s, "-printableErrors") &&
		!strings.Contains(s, "There were a total of 0 errors.") {
		if strings.Contains(s, "ERROR") ||
			strings.Contains(s, "Err") ||
			strings.Contains(s, "error") ||
			strings.Contains(s, "Exception:") ||
			strings.Contains(s, "Killed ") ||
			strings.Contains(s, "Segmentation fault") ||
			strings.Contains(s, "end of process: incomplete") ||
			(strings.Contains(s, "Aborted") && strings.Contains(s, "core dumped ")) {
			return true
		}
	}
	return false
}

// ExtractTimeTag returns the valid timestamp YYYYMMDDHHMMSS in filepath or empty string.
func ExtractTimeTag(fileName string) string {
	s := strings.Split(fileName, "/")
	for i := range s {
		if s[i] != "" && s[i][0] == '2' {
			if ValidTimeTag(s[i]) {
				return s[i]
			}
		}
	}
	return ""
}
