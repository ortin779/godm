package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// downloadFile orchestrates single-part or multipart downloading.
func downloadFile(ctx context.Context, task *downloadTask, destPath string, maxParts int) error {
	totalSize, supportsRange, err := probeURL(ctx, task.download.URL)
	if err != nil {
		return fmt.Errorf("probing URL: %w", err)
	}

	task.mu.Lock()
	task.download.TotalSize = totalSize
	alreadyDownloaded := task.download.Downloaded
	task.mu.Unlock()

	if supportsRange && totalSize > 0 && maxParts > 1 {
		return multipartDownload(ctx, task, destPath, totalSize, alreadyDownloaded, maxParts)
	}
	return singlePartDownload(ctx, task, destPath, alreadyDownloaded)
}

// singlePartDownload downloads the file in one request, supporting resume via Range.
func singlePartDownload(ctx context.Context, task *downloadTask, destPath string, resumeFrom int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.download.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", probeUserAgent)

	flags := os.O_CREATE | os.O_WRONLY
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the server returns HTML the URL is behind a redirect, login, or
	// bot-protection page — writing that to disk would corrupt the file.
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "text/html") {
		return fmt.Errorf("server returned an HTML page instead of the file (redirect or bot-protection?). Try a direct download link")
	}

	f, err := os.OpenFile(destPath, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return copyWithProgress(ctx, task, f, resp.Body, resumeFrom)
}

// multipartDownload splits the file into N parts and downloads concurrently.
func multipartDownload(ctx context.Context, task *downloadTask, destPath string, totalSize, resumeFrom int64, parts int) error {
	// On a fresh download (not a resume), remove any stale part files left
	// behind by a previous failed attempt so we never append new data onto old.
	if resumeFrom == 0 {
		if matches, err := filepath.Glob(destPath + ".part*"); err == nil {
			for _, m := range matches {
				os.Remove(m)
			}
		}
	}

	// Determine actual number of parts based on remaining size
	remaining := totalSize - resumeFrom
	if int64(parts) > remaining {
		parts = int(remaining)
	}

	task.mu.Lock()
	task.download.Parts = parts
	task.mu.Unlock()

	partSize := (totalSize - resumeFrom) / int64(parts)
	var downloaded int64 = resumeFrom

	var wg sync.WaitGroup
	errCh := make(chan error, parts)
	partPaths := make([]string, parts)

	for i := 0; i < parts; i++ {
		start := resumeFrom + int64(i)*partSize
		end := start + partSize - 1
		if i == parts-1 {
			end = totalSize - 1
		}
		partPath := fmt.Sprintf("%s.part%d", destPath, i)
		partPaths[i] = partPath

		// Check if part file already exists (resume support)
		var partResumeFrom int64
		if info, err := os.Stat(partPath); err == nil {
			partResumeFrom = info.Size()
		}

		wg.Add(1)
		go func(partPath string, start, end, partResume int64) {
			defer wg.Done()
			n, err := downloadPart(ctx, task.download.URL, partPath, start+partResume, end, partResume > 0)
			atomic.AddInt64(&downloaded, n)
			task.mu.Lock()
			task.download.Downloaded = atomic.LoadInt64(&downloaded)
			task.mu.Unlock()
			if err != nil {
				errCh <- err
			}
		}(partPath, start, end, partResumeFrom)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return err
	}

	// Check context before merging
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return mergeParts(destPath, partPaths)
}

// downloadPart downloads a byte range into partPath and returns bytes written.
// resume=true means the part file already exists and we are appending to it.
func downloadPart(ctx context.Context, rawURL, partPath string, start, end int64, resume bool) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", probeUserAgent)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 206 Partial Content is the only valid response for a range request.
	// A 200 means the server ignored the Range header and returned the full
	// file — writing that into every part would produce a file N× too large.
	if resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("server returned HTTP %d for range request (expected 206)", resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resume {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(partPath, flags, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(f, resp.Body)
}

// mergeParts concatenates part files into the final destination file.
func mergeParts(destPath string, partPaths []string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, p := range partPaths {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, f)
		f.Close()
		if err != nil {
			return err
		}
		os.Remove(p)
	}
	return nil
}

// cleanPartFiles removes temporary part files for a given destination.
func cleanPartFiles(destPath string) {
	pattern := destPath + ".part*"
	matches, _ := filepath.Glob(pattern)
	for _, m := range matches {
		os.Remove(m)
	}
	os.Remove(destPath)
}

// copyWithProgress copies from src to dst while updating task.Downloaded and speed.
func copyWithProgress(ctx context.Context, task *downloadTask, dst io.Writer, src io.Reader, startOffset int64) error {
	buf := make([]byte, 32*1024)
	var totalWritten int64
	lastTime := time.Now()
	var lastBytes int64

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			totalWritten += int64(written)
			if writeErr != nil {
				return writeErr
			}

			now := time.Now()
			elapsed := now.Sub(lastTime).Seconds()
			if elapsed >= 0.5 {
				speed := int64(float64(totalWritten-lastBytes) / elapsed)
				lastBytes = totalWritten
				lastTime = now

				task.mu.Lock()
				task.download.Downloaded = startOffset + totalWritten
				task.download.Speed = speed
				task.mu.Unlock()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				task.mu.Lock()
				task.download.Downloaded = startOffset + totalWritten
				task.download.Speed = 0
				task.mu.Unlock()
				return nil
			}
			return readErr
		}
	}
}
