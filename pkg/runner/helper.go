/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// OverWriteGobFile saves data to disk encoded as a gob file, overwriting any existing file.
func OverWriteGobFile(fn string, data any) error {
	// os.Create automatically truncates the file if it exists, so explicit removal is not needed.
	return WriteGobDataToFile(fn, data)
}

// WriteGobDataToFile writes data to a file in Gob format, compressing if the extension is ".gz".
func WriteGobDataToFile(fn string, data any) error {
	// 1. Create/Truncate File
	fw, err := os.Create(fn)
	if err != nil {
		return err
	}
	// 2. Setup Writer Chain: Gob -> Bufio -> Gzip (Optional) -> File
	var outputWriter io.Writer = fw
	var gzw *gzip.Writer
	// Add Gzip layer if needed
	if filepath.Ext(fn) == ".gz" {
		gzw = gzip.NewWriter(fw)
		outputWriter = gzw
	}
	// Add Buffer layer (reduces syscalls for small Gob writes)
	bw := bufio.NewWriter(outputWriter)
	// 3. Encode
	if err = gob.NewEncoder(bw).Encode(data); err != nil {
		err = fmt.Errorf("gob encode failed: %w", err)
	}
	if err == nil {
		err = bw.Flush()
	}
	var ger error
	if gzw != nil {
		ger = gzw.Close()
	}
	return ValidateErrors(err, ValidateErrors(ger, fw.Close()))
}

// LoadDataFromGobFile reads Gob-encoded data.
// It creates a robust fallback: if 'file.gz' is requested but missing, it checks for 'file' (and vice versa),
// adjusting the decompression logic automatically.
func LoadDataFromGobFile(fn string, data any) error {
	// 1. Resolve actual filename and compression state
	finalPath, isCompressed, err := resolveFile(fn)
	if err != nil {
		return err
	}
	// 2. Open File
	fw, fer := os.Open(finalPath)
	if fer != nil {
		return fer
	}
	// 3. Setup Reader Chain: File -> Gzip (Optional) -> Bufio -> Gob
	var inputReader io.Reader = fw
	var gzr *gzip.Reader
	if isCompressed {
		var ger error
		gzr, ger = gzip.NewReader(fw)
		if ger != nil {
			return ValidateErrors(ger, fw.Close())
		}
		inputReader = gzr
	}
	// Buffer the input for performance
	br := bufio.NewReader(inputReader)
	// 4. Decode
	if err = gob.NewDecoder(br).Decode(data); err != nil {
		err = fmt.Errorf("gob decode failed for %s: %w", finalPath, err)
	}
	var ger error
	if gzr != nil {
		ger = gzr.Close()
	}
	return ValidateErrors(err, ValidateErrors(ger, fw.Close()))
}

// resolveFile handles the logic of finding the correct file and determining if it needs decompression.
func resolveFile(requested string) (actual string, isGz bool, err error) {
	// Case 1: File exists exactly as requested
	if _, ret := os.Stat(requested); ret == nil {
		return requested, strings.HasSuffix(requested, ".gz"), nil
	}

	// Case 2: File missing. Check alternates.
	if before, ok := strings.CutSuffix(requested, ".gz"); ok {
		// Requested .gz, try plain
		alt := before
		if _, ret := os.Stat(alt); ret == nil {
			return alt, false, nil // Found a plain file
		}
	} else {
		// Requested plain, try .gz
		alt := requested + ".gz"
		if _, ret := os.Stat(alt); ret == nil {
			return alt, true, nil // Found a compressed file
		}
	}

	return "", false, os.ErrNotExist
}

// ReplaceAtIndex replaces the rune at index i.
// Safe for Unicode strings.
func ReplaceAtIndex(in string, r rune, i int) string {
	runes := []rune(in)
	if i < 0 || i >= len(runes) {
		return in // Or return "" depending on desired safety, but returning original is safer than empty
	}
	runes[i] = r
	return string(runes)
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	destination, err := os.Create(dst)
	if err != nil {
		return ValidateErrors(err, source.Close())
	}
	// io.Copy handles buffering automatically and efficiently
	_, err = io.Copy(destination, source)
	return ValidateErrors(err, ValidateErrors(destination.Close(), source.Close()))
}

// StringToGzipBytes compresses a string.
func StringToGzipBytes(s string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write([]byte(s))
	if err != nil {
		if ret := gz.Close(); ret != nil {
			return nil, errors.Join(err, ret)
		}
		return nil, err
	}
	if err = gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GzipBytesToString decompresses bytes.
func GzipBytesToString(z []byte) (string, error) {
	if len(z) == 0 {
		return "", nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(z))
	if err != nil {
		return "", err
	}
	// Use strings.Builder for efficiency
	var sb strings.Builder
	_, err = io.Copy(&sb, zr)
	err = ValidateErrors(err, zr.Close())
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// SplitAccessionVersion returns an empty version when no number; a format is at least 5 consecutive chars + 1 dot + number.
func SplitAccessionVersion(acn string) (accession string, version string) {
	accession = acn
	version = ""
	sp := strings.Split(acn, ".")
	if len(sp) == 2 && len(sp[0]) > 4 {
		if CheckAllAreDigits(sp[1]) == len(sp[1]) {
			accession = sp[0]
			version = sp[1]
		}
	}
	return accession, version
}

// HelpDoc returns a formatted help document.
func HelpDoc(fileNames []string, body string) string {
	if fileNames == nil {
		return ""
	}
	var buf bytes.Buffer
	if len(fileNames) > 0 {
		_, _ = fmt.Fprintf(&buf, "Sources:\n\n")
	}
	if body != "" {
		_, _ = fmt.Fprintf(&buf, "%s\n\n", body)
	}
	for i, fn := range fileNames {
		_, _ = fmt.Fprintf(&buf, "  %d) %s\n", i+1, fn)
	}
	buf.WriteString("\n")
	return buf.String()
}

// DataOrNothing returns an empty string if the input string is empty or matches the noData value, otherwise it returns the input string.
func DataOrNothing(s, noData string) string {
	if s == "" || s == noData {
		return ""
	}
	return s
}

// ExtractNumber removes a specific prefix and leading zeros e.g., GO:000123 returns 123 if the prefix is "GO:"
func ExtractNumber(s, noData, prefix string) string {
	s = DataOrNothing(s, noData)
	if s == "" {
		return ""
	}
	return strings.TrimLeft(strings.TrimPrefix(s, prefix), "0")
}

// PopOutString returns the string between prefix and suffix, or an empty string if not found.
func PopOutString(s, prefix, suffix string) string {
	i := strings.Index(s, prefix)
	j := strings.Index(s, suffix)
	if i == -1 || j < i {
		return ""
	}
	return s[i+len(prefix) : j]
}
