/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package cowriter

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"

	"github.com/sas-quantome/interlace-go/pkg/runner"
)

type WtrBitFlag uint8

const (
	WtrIsOpen WtrBitFlag = 1 << iota
	WtrIsGzip
	WtrIsTemp
	WtrRemoved
	WtrAliasA
	WtrAliasB
	WtrAliasC
	WtrAliasD
)

type Cowriter struct {
	File    *os.File
	Buf     bytes.Buffer
	gz      *gzip.Writer
	mu      sync.RWMutex
	bitFlag runner.BitMask[WtrBitFlag]
	nBytes  int
}

// NewCowriter creates and returns a new instance of Cowriter.
func NewCowriter() *Cowriter {
	return new(Cowriter)
}

// Open opens a file for writing. If isTemporary is true and the dirpath is empty, it uses the default temporary path
// returned by os.TempDir. If the temporary filename includes a "*", the random string replaces the last "*"; else,
// it is appended at the end. If isTemporary is false, the os.Create function is used instead of os.CreateTemp and the
// filename is used as-is. If isGzip is true, the file is compressed using gzip, and the filename is appended with
// ".gz" if not already present.
func (cw *Cowriter) Open(dirpath, filename string, isTemporary, isGzip bool) error {
	if cw.File != nil {
		return errors.New("cannot open: cowrite file is not nil")
	}
	if cw.bitFlag.Has(WtrIsOpen) {
		return errors.New("cannot open: cowrite file flags open: " + cw.File.Name())
	}
	var err error
	var fn string
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if filename != "" && isGzip && filepath.Ext(filename) != ".gz" {
		fn = filename + ".gz"
	} else {
		fn = filename
	}
	if fn == "" || fn == "-" || fn == "." {
		cw.File = os.Stdout
		cw.bitFlag.Add(WtrIsOpen)
		return nil
	} else if isTemporary {
		cw.bitFlag.Add(WtrIsTemp)
		cw.File, err = os.CreateTemp(dirpath, fn)
	} else {
		cw.File, err = os.Create(path.Join(dirpath, fn))
	}
	if cw.File == nil {
		return errors.New("failed open to write: cowrite file is nil: " + fn)
	}
	if err != nil {
		return nil
	}
	if isGzip {
		cw.bitFlag.Add(WtrIsGzip)
		if cw.gz, err = gzip.NewWriterLevel(cw.File, gzip.DefaultCompression); err != nil {
			return nil
		}
	}
	cw.bitFlag.Add(WtrIsOpen)
	return nil
}

// Close closes the file and flushes the buffer.
func (cw *Cowriter) Close() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.File == nil {
		return nil
	}
	var err error
	if cw.bitFlag.Has(WtrIsGzip) {
		err = runner.ValidateErrors(cw.gz.Flush(), cw.gz.Close())
	}
	if cw.File == os.Stdout {
		cw.File = nil
		cw.bitFlag.Clear(WtrIsOpen)
		return nil
	}
	err = runner.ValidateErrors(err, cw.File.Close())
	if err == nil {
		cw.bitFlag.Clear(WtrIsOpen)
	}
	return err
}

// Write writes the given bytes to the file.
func (cw *Cowriter) Write(p []byte) error {
	var err error
	var n int
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.File == nil {
		return errors.New("cannot write: cowrite file is nil")
	}
	if cw.bitFlag.Has(WtrIsGzip) {
		n, err = cw.gz.Write(p)
	} else {
		n, err = cw.File.Write(p)
	}
	cw.nBytes += n
	return err
}

// WriteString writes the given string to the file.
func (cw *Cowriter) WriteString(s string) error {
	var err error
	var n int
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.File == nil {
		return errors.New("cannot write: cowrite file is nil")
	}
	if cw.bitFlag.Has(WtrIsGzip) {
		n, err = cw.gz.Write([]byte(s))
	} else {
		n, err = cw.File.WriteString(s)
	}
	cw.nBytes += n
	return err
}

// BytesWritten returns the number of bytes written to the file.
func (cw *Cowriter) BytesWritten() int {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.nBytes
}

// BytesToDisk flushes the buffer to disk if it exceeds the minimum chunk size.
func (cw *Cowriter) BytesToDisk(minChunkSize int) error {
	if cw.Buf.Len() == 0 || cw.Buf.Len() < minChunkSize || minChunkSize <= 0 {
		return nil
	}
	if err := cw.Write(cw.Buf.Bytes()); err != nil {
		return err
	}
	cw.Buf.Reset()
	return nil
}

// FileSize returns the size of the file in bytes File.Stat().
func (cw *Cowriter) FileSize() int64 {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	if cw.File == nil {
		return 0
	}
	fi, _ := cw.File.Stat()
	return fi.Size()
}

// IsOpen returns true if the file is open.
func (cw *Cowriter) IsOpen() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.bitFlag.Has(WtrIsOpen)
}

// IsTemporary returns true if the file is temporary.
func (cw *Cowriter) IsTemporary() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.bitFlag.Has(WtrIsTemp)
}

// IsGzip returns true if the file is compressed using gzip.
func (cw *Cowriter) IsGzip() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.bitFlag.Has(WtrIsGzip)
}

// IsRemoved returns true if the file has been removed.
func (cw *Cowriter) IsRemoved() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.bitFlag.Has(WtrRemoved)
}

// Path returns the full path of the file.
func (cw *Cowriter) Path() string {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	if cw.File == nil {
		return ""
	}
	return cw.File.Name()
}

// Remove closes the file if not done, removes it, clears, and resets the bit flags; lives the writer ready for reuse.
func (cw *Cowriter) Remove() error {
	if cw.File == nil {
		return nil
	}
	if cw.bitFlag.Has(WtrRemoved) {
		return nil
	}
	if cw.bitFlag.Has(WtrIsOpen) {
		if err := cw.Close(); err != nil {
			return err
		}
	}
	cw.mu.Lock()
	defer cw.mu.Unlock()
	err := os.Remove(cw.File.Name())
	if err == nil {
		cw.nBytes = 0
		cw.File = nil
		cw.gz = nil
		cw.bitFlag.ClearAll()
		cw.bitFlag.Add(WtrRemoved)
		cw.Buf.Reset()
	}
	return err
}

// Clear resets the writer to its initial state and makes it ready for reuse.
func (cw *Cowriter) Clear() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.nBytes = 0
	cw.File = nil
	cw.gz = nil
	cw.bitFlag.ClearAll()
	cw.Buf.Reset()
}

// Sort sorts the file using the given arguments.
func (cw *Cowriter) Sort(args string) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.File == nil {
		return errors.New("cannot sort: cowrite file is nil")
	}
	if cw.bitFlag.Has(WtrIsOpen) {
		return errors.New("cannot sort: file is still open: " + cw.File.Name())
	}
	inp := cw.File.Name()
	out := inp + ".sorted"
	par := ""
	if cw.bitFlag.Has(WtrIsGzip) {
		if args == "" {
			par = fmt.Sprintf("gzip -dc %s | sort | gzip -c > %s", inp, out)
		} else {
			par = fmt.Sprintf("gzip -dc %s | sort %s | gzip -c > %s", inp, args, out)
		}
	} else {
		if args == "" {
			par = fmt.Sprintf("sort -o %s %s", out, inp)
		} else {
			par = fmt.Sprintf("sort %s -o %s %s", args, out, inp)
		}
	}
	cmd := exec.Command("bash", "-c", par)
	err := cmd.Run()
	if err == nil {
		if err = os.Remove(inp); err == nil {
			return os.Rename(out, inp)
		}
	}
	return err
}

// SetBitFlag sets the bit flag to the given value.
func (cw *Cowriter) SetBitFlag(flag WtrBitFlag, value bool) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if value {
		cw.bitFlag.Add(flag)
	} else {
		cw.bitFlag.Clear(flag)
	}
}

// GetBitFlag returns the bit flag value.
func (cw *Cowriter) GetBitFlag(flag WtrBitFlag) bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.bitFlag.Has(flag)
}
