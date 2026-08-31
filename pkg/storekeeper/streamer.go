/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

const CmdMaxLen = 1024

// DatasetNamespace represents a unique namespace for a specific asset dataset.
type DatasetNamespace string

// Streamer represents a generic data streamer for transferring and processing sets of parallel arrays.
// It supports carrying multiple distinct datasets simultaneously.
type Streamer struct {
	mu    sync.RWMutex
	Err   error    // The caller handles this err.
	Msg   string   // The caller handles this msg.
	Paths []string // Metadata: <entity>/<asset>/<property>/<name>
	Items [][]any  // Columns: One slice of data per Path entry.
}

// NewStreamer initializes a new Streamer.
func NewStreamer() *Streamer {
	return &Streamer{
		Paths: make([]string, 0),
		Items: make([][]any, 0),
	}
}

// KeyNamespace extracts the dataset namespace (Entity/Asset) from a path string.
func KeyNamespace(path string) (DatasetNamespace, error) {
	entity, rest, ok := strings.Cut(path, "/")
	if !ok {
		return "", errors.New("invalid metadata path: " + path)
	}
	asset, _, ok := strings.Cut(rest, "/")
	if !ok {
		// If there is no property after the asset, the asset is the rest of the string.
		// However, standard paths usually have properties. We allow loose parsing here.
		return DatasetNamespace(entity + "/" + rest), nil
	}
	return DatasetNamespace(entity + "/" + asset), nil
}

// Reset clears all fields to recycle the Streamer instance.
func (s *Streamer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Err = nil
	s.Msg = ""
	s.Paths = s.Paths[:0]
	s.Items = s.Items[:0]
}

// Add appends a metadata path and its corresponding data column safely.
func (s *Streamer) Add(path string, col []any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if col == nil {
		return errors.New("cannot add: item column is nil")
	}

	// We cannot strictly enforce length consistency here easily without reparsing
	// the namespace of every existing path (O(N)).
	// We rely on IntegrityCheck() to validate consistency before transmission.

	s.Paths = append(s.Paths, path)
	s.Items = append(s.Items, col)
	return nil
}

// String returns a formatted row representation for a specific dataset namespace.
func (s *Streamer) String(key DatasetNamespace, row int, delimiter string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var b strings.Builder
	first := true
	keyPrefix := string(key)

	for i, path := range s.Paths {
		// Only include columns belonging to the requested dataset namespace
		if strings.HasPrefix(path, keyPrefix) {
			if row >= len(s.Items[i]) {
				continue
			}

			if !first {
				b.WriteString(delimiter)
			}
			if _, err := fmt.Fprintf(&b, "%v", s.Items[i][row]); err != nil {
				return ""
			}
			first = false
		}
	}
	return b.String()
}

// IntegrityCheck validates that columns strictly adhere to length consistency
// WITHIN their specific dataset namespace (Entity/Asset).
func (s *Streamer) IntegrityCheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Paths == nil || s.Items == nil {
		return errors.New("integrity check failed: streamer is not initialized")
	}
	if len(s.Paths) != len(s.Items) {
		return fmt.Errorf("integrity check failed: paths count (%d) != items count (%d)", len(s.Paths), len(s.Items))
	}

	// Map to track the expected row count for each Namespace found.
	// Key = "Entity/Asset", Value = Row Count
	datasetLengths := make(map[string]int)

	for i, path := range s.Paths {
		// 1. Extract Namespace: "Entity/Asset"
		entity, rest, ok := strings.Cut(path, "/")
		if !ok {
			return fmt.Errorf("integrity check failed: invalid path format '%s'", path)
		}

		asset, _, _ := strings.Cut(rest, "/") // If no 3rd part, Asset is the rest
		namespace := entity + "/" + asset

		// 2. Validate Row Count Consistency
		currentLen := len(s.Items[i])

		if expectedLen, seen := datasetLengths[namespace]; seen {
			if currentLen != expectedLen {
				return fmt.Errorf("integrity check failed: namespace '%s' has inconsistent row counts: expected %d, got %d (at path %s)",
					namespace, expectedLen, currentLen, path)
			}
		} else {
			// First time seeing this dataset, record the expected length
			datasetLengths[namespace] = currentLen
		}
	}
	return nil
}

// --- Network Encoding/Decoding ---

func (s *Streamer) CmdEncode(conn net.Conn, cmd string) error {
	s.mu.RLock()
	defer s.mu.RUnlock() // Defer unlock to ensure safety during encode

	if len(cmd) < 2 || len(cmd) > CmdMaxLen {
		return fmt.Errorf("invalid command length: %d", len(cmd))
	}

	bw := bufio.NewWriter(conn)

	if _, err := bw.WriteString(cmd + "\n"); err != nil {
		return fmt.Errorf("write cmd failed: %w", err)
	}

	if err := gob.NewEncoder(bw).Encode(s); err != nil {
		return fmt.Errorf("gob encode failed: %w", err)
	}

	return bw.Flush()
}

func CmdDecode(conn net.Conn) (*Streamer, string, error) {
	br := bufio.NewReader(conn)

	cmdStr, err := readBoundedLine(br, CmdMaxLen)
	if err != nil {
		return nil, "", fmt.Errorf("read cmd failed: %w", err)
	}

	streamer := NewStreamer()
	if err = gob.NewDecoder(br).Decode(streamer); err != nil {
		return nil, cmdStr, fmt.Errorf("gob decode failed: %w", err)
	}

	return streamer, cmdStr, nil
}

func readBoundedLine(r *bufio.Reader, maxLen int) (string, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		if errors.Is(err, bufio.ErrBufferFull) {
			return "", errors.New("command exceeds buffer size")
		}
		if err == io.EOF && len(line) > 0 {
			// Allow EOF if line exists
		} else {
			return "", err
		}
	}

	s := strings.TrimSpace(string(line))
	if len(s) > maxLen {
		return "", fmt.Errorf("command exceeds max length %d", maxLen)
	}
	if len(s) < 2 {
		return "", errors.New("command too short")
	}
	return s, nil
}

func (s *Streamer) NetEncode(conn net.Conn) error {
	bw := bufio.NewWriter(conn)

	s.mu.RLock()
	err := gob.NewEncoder(bw).Encode(s)
	s.mu.RUnlock()

	if err != nil {
		return err
	}
	return bw.Flush()
}

func NetDecode(conn net.Conn) (*Streamer, error) {
	br := bufio.NewReader(conn)
	streamer := NewStreamer()
	if err := gob.NewDecoder(br).Decode(streamer); err != nil {
		return nil, err
	}
	return streamer, nil
}

func (s *Streamer) BufEncode(buf *bytes.Buffer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return gob.NewEncoder(buf).Encode(s)
}

func BufDecode(buf *bytes.Buffer) (*Streamer, error) {
	s := NewStreamer()
	err := gob.NewDecoder(buf).Decode(s)
	return s, err
}
