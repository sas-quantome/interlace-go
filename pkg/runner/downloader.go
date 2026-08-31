/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// Download fetches files concurrently in the assigned directory from URL addresses.
func (mgr *Manager) Download(dirPath string, urls []string) error {
	if mgr.GetBitFlag(MgrFlgNoDownload) == true {
		return nil
	}
	ch := make(chan string)
	for _, url := range urls {
		go mgr.fetch(dirPath, url, ch)
	}
	for range urls {
		ret := fmt.Sprint(<-ch)
		if strings.Contains(ret, "fail") {
			mgr.War(ret)
		} else {
			mgr.Say(ret)
		}
	}
	return nil
}

// Fetch downloads a file concurrently.
func (mgr *Manager) fetch(dirPath, url string, ch chan<- string) {
	filePath := path.Join(dirPath, path.Base(url))
	if _, err := os.Stat(filePath); err == nil {
		ch <- fmt.Sprintf("file exists: %s", filePath)
		return
	}
	start := time.Now()
	response, err1 := http.Get(url)
	if err1 != nil {
		ch <- fmt.Sprintf("http get failure %s: %v", url, err1)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			mgr.War(err.Error())
		}
	}(response.Body)

	outFile, err2 := os.Create(filePath)
	if err2 != nil {
		ch <- fmt.Sprintf("failed creating file %s: %v", url, err2)
		return
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			mgr.War(err.Error())
		}
	}(outFile)

	b, err4 := io.Copy(outFile, response.Body)
	if err4 != nil {
		ch <- fmt.Sprintf("failed writing %s: %v", url, err4)
		return
	}
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("downloaded: %.2fs  %7d  %s", secs, b, url)
}
