package downloader

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

// DownloadInfo holds metadata about a URL retrieved before the download starts.
type DownloadInfo struct {
	URL         string `json:"url"`
	FileName    string `json:"fileName"`
	TotalSize   int64  `json:"totalSize"`
	ContentType string `json:"contentType"`
	Resumable   bool   `json:"resumable"`
}

var driveFileRe = regexp.MustCompile(`/file/d/([^/?#]+)`)

// normalizeURL converts well-known share/viewer URLs to direct download URLs.
func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if strings.Contains(u.Host, "drive.google.com") {
		// /file/d/FILE_ID/view → direct download
		if m := driveFileRe.FindStringSubmatch(u.Path); len(m) == 2 {
			return "https://drive.google.com/uc?export=download&id=" + m[1]
		}
		// /uc without export=download
		if strings.HasPrefix(u.Path, "/uc") {
			q := u.Query()
			if q.Get("export") == "" {
				q.Set("export", "download")
				u.RawQuery = q.Encode()
				return u.String()
			}
		}
	}
	return rawURL
}

const probeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Probe fetches metadata for rawURL without downloading the file body.
// It normalises the URL (e.g. Google Drive share → direct download link),
// follows redirects, detects HTML protection/viewer pages, and extracts
// the filename from Content-Disposition or the URL path.
func Probe(ctx context.Context, rawURL string) (DownloadInfo, error) {
	rawURL = normalizeURL(rawURL)

	client := &http.Client{Timeout: 15 * time.Second}

	// Try HEAD first; fall back to a minimal byte-range GET for servers that
	// reject HEAD (some CDNs, GCS signed URLs, etc.).
	var resp *http.Response
	for _, method := range []string{http.MethodHead, http.MethodGet} {
		req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
		if err != nil {
			return DownloadInfo{}, err
		}
		req.Header.Set("User-Agent", probeUserAgent)
		if method == http.MethodGet {
			req.Header.Set("Range", "bytes=0-0")
		}
		resp, err = client.Do(req)
		if err != nil {
			return DownloadInfo{}, err
		}
		resp.Body.Close()
		if resp.StatusCode < 400 {
			break
		}
	}

	// Detect HTML response — Cloudflare challenge, Google Drive viewer, login
	// pages, etc. None of these are actual files.
	ct := resp.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if mediaType == "text/html" {
		return DownloadInfo{}, fmt.Errorf(
			"URL returned an HTML page instead of a file — it may be behind a " +
				"login, a bot-protection page, or a file viewer. Try a direct download link",
		)
	}

	// Resolve filename: Content-Disposition wins, then URL path.
	fileName := fileNameFromHeader(resp.Header.Get("Content-Disposition"))
	if fileName == "" {
		fileName = extractFileName(rawURL)
	}
	fileName = sanitizeFileName(fileName)
	if fileName == "" {
		fileName = "download_" + time.Now().Format("20060102_150405")
	}

	// Resolve total size.
	// For a 206 Partial Content response to "Range: bytes=0-0", the actual
	// total is in Content-Range (e.g. "bytes 0-0/1048576"), not Content-Length.
	size := resp.ContentLength
	if resp.StatusCode == http.StatusPartialContent {
		if cr := resp.Header.Get("Content-Range"); cr != "" {
			if idx := strings.LastIndex(cr, "/"); idx != -1 {
				var total int64
				if n, _ := fmt.Sscanf(cr[idx+1:], "%d", &total); n == 1 && total > 0 {
					size = total
				}
			}
		}
	}

	return DownloadInfo{
		URL:         rawURL,
		FileName:    fileName,
		TotalSize:   size,
		ContentType: ct,
		Resumable:   strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes"),
	}, nil
}

// fileNameFromHeader parses a Content-Disposition header and returns the
// suggested filename, preferring the RFC 5987 filename* parameter.
func fileNameFromHeader(cd string) string {
	if cd == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(cd)
	if err != nil {
		return ""
	}
	if name, ok := params["filename*"]; ok {
		// Strip charset and language prefix: UTF-8''encoded-value
		if idx := strings.LastIndex(name, "'"); idx != -1 {
			name = name[idx+1:]
		}
		if dec, err := url.QueryUnescape(name); err == nil {
			return dec
		}
		return name
	}
	return params["filename"]
}

// sanitizeFileName strips path separators and OS-illegal characters from a
// candidate file name.
func sanitizeFileName(name string) string {
	name = path.Base(name)
	if idx := strings.Index(name, "?"); idx != -1 {
		name = name[:idx]
	}
	name = strings.TrimSpace(name)
	for _, c := range `\/:*?"<>|` {
		name = strings.ReplaceAll(name, string(c), "_")
	}
	return name
}
