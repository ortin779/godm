package downloader

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EmitFn is the function signature for emitting Wails events.
type EmitFn func(eventName string, data ...interface{})

// downloadTask holds the runtime state of a single download.
type downloadTask struct {
	download Download
	cancel   context.CancelFunc
	mu       sync.Mutex
}

func (t *downloadTask) snapshot() Download {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.download
}

// Manager orchestrates all downloads.
type Manager struct {
	ctx       context.Context
	cfg       Config
	cfgMu     sync.RWMutex
	tasks     map[string]*downloadTask
	tasksMu   sync.RWMutex
	semaphore chan struct{}
	emit      EmitFn
}

// NewManager creates a Manager, restores persisted state, and starts the progress emitter.
func NewManager(ctx context.Context, cfg Config, emit EmitFn) *Manager {
	m := &Manager{
		ctx:       ctx,
		cfg:       cfg,
		tasks:     make(map[string]*downloadTask),
		semaphore: make(chan struct{}, cfg.MaxConcurrentDownloads),
		emit:      emit,
	}
	m.loadState()
	go m.progressLoop()
	return m
}

// progressLoop emits progress events every 500ms.
func (m *Manager) progressLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			snapshots := m.GetAll()
			for _, d := range snapshots {
				if d.Status == StatusActive || d.Status == StatusPaused {
					m.emit("download:progress", d)
				}
			}
		}
	}
}

// ProbeDownload fetches metadata for rawURL without downloading the file.
func (m *Manager) ProbeDownload(rawURL string) (DownloadInfo, error) {
	return Probe(m.ctx, rawURL)
}

// RevealFile opens the file's containing folder in the system file manager
// and selects the file.
func (m *Manager) RevealFile(id string) error {
	task := m.getTask(id)
	if task == nil {
		return fmt.Errorf("download %s not found", id)
	}
	m.cfgMu.RLock()
	dir := m.cfg.DownloadDir
	m.cfgMu.RUnlock()

	filePath := filepath.Join(dir, task.download.FileName)
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-R", filePath).Start()
	case "windows":
		return exec.Command("explorer", "/select,", filePath).Start()
	default:
		return exec.Command("xdg-open", filepath.Dir(filePath)).Start()
	}
}

// Add enqueues a new download for the given URL and fileName.
// If fileName is empty the name is derived from the URL.
func (m *Manager) Add(rawURL, fileName string) (Download, error) {
	if fileName == "" {
		fileName = extractFileName(rawURL)
	}
	id := uuid.NewString()

	d := Download{
		ID:        id,
		URL:       rawURL,
		FileName:  fileName,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	task := &downloadTask{download: d}
	m.tasksMu.Lock()
	m.tasks[id] = task
	m.tasksMu.Unlock()

	go m.saveState()
	go m.runDownload(task)
	return d, nil
}

// runDownload acquires a semaphore slot then starts the download.
func (m *Manager) runDownload(task *downloadTask) {
	// Acquire slot
	select {
	case m.semaphore <- struct{}{}:
	case <-m.ctx.Done():
		return
	}
	defer func() { <-m.semaphore }()

	// If the download was paused or cancelled while waiting for a slot, do not start.
	task.mu.Lock()
	if task.download.Status == StatusPaused || task.download.Status == StatusCancelled {
		task.mu.Unlock()
		return
	}

	taskCtx, cancel := context.WithCancel(m.ctx)
	task.cancel = cancel
	task.download.Status = StatusActive
	task.mu.Unlock()

	m.cfgMu.RLock()
	cfg := m.cfg
	m.cfgMu.RUnlock()

	destPath := filepath.Join(cfg.DownloadDir, task.download.FileName)
	err := downloadFile(taskCtx, task, destPath, cfg.MaxPartsPerDownload)

	task.mu.Lock()
	if err != nil {
		if taskCtx.Err() != nil {
			// cancelled or paused externally — status already set
		} else {
			task.download.Status = StatusError
			task.download.Error = err.Error()
		}
	} else {
		task.download.Status = StatusCompleted
		if task.download.TotalSize > 0 {
			// Known size: snap Downloaded to TotalSize to avoid drift from rounding
			task.download.Downloaded = task.download.TotalSize
		} else {
			// Unknown Content-Length: copyWithProgress already set Downloaded to
			// the actual bytes written; mirror that into TotalSize for display
			task.download.TotalSize = task.download.Downloaded
		}
	}
	task.mu.Unlock()

	// Emit final state and persist
	m.emit("download:progress", task.snapshot())
	go m.saveState()
}

// Pause stops an active or pending download, preserving progress.
func (m *Manager) Pause(id string) error {
	task := m.getTask(id)
	if task == nil {
		return fmt.Errorf("download %s not found", id)
	}
	task.mu.Lock()
	defer task.mu.Unlock()
	if task.download.Status != StatusActive && task.download.Status != StatusPending {
		return fmt.Errorf("download cannot be paused")
	}
	task.download.Status = StatusPaused
	if task.cancel != nil {
		task.cancel()
	}
	go m.saveState()
	return nil
}

// Resume restarts a paused download from where it left off.
func (m *Manager) Resume(id string) error {
	task := m.getTask(id)
	if task == nil {
		return fmt.Errorf("download %s not found", id)
	}
	task.mu.Lock()
	if task.download.Status != StatusPaused {
		task.mu.Unlock()
		return fmt.Errorf("download is not paused")
	}
	task.download.Status = StatusPending
	task.mu.Unlock()

	go m.saveState()
	go m.runDownload(task)
	return nil
}

// Cancel stops and removes a download.
func (m *Manager) Cancel(id string) error {
	task := m.getTask(id)
	if task == nil {
		return fmt.Errorf("download %s not found", id)
	}
	task.mu.Lock()
	task.download.Status = StatusCancelled
	if task.cancel != nil {
		task.cancel()
	}
	task.mu.Unlock()

	m.tasksMu.Lock()
	delete(m.tasks, id)
	m.tasksMu.Unlock()

	go m.saveState()

	// Clean up partial files
	m.cfgMu.RLock()
	dir := m.cfg.DownloadDir
	m.cfgMu.RUnlock()
	cleanPartFiles(filepath.Join(dir, task.download.FileName))

	return nil
}

// GetAll returns a snapshot of all non-cancelled downloads.
func (m *Manager) GetAll() []Download {
	m.tasksMu.RLock()
	defer m.tasksMu.RUnlock()
	result := make([]Download, 0, len(m.tasks))
	for _, t := range m.tasks {
		d := t.snapshot()
		if d.Status != StatusCancelled {
			result = append(result, d)
		}
	}
	return result
}

// GetActive returns downloads that are active or paused.
func (m *Manager) GetActive() []Download {
	all := m.GetAll()
	result := make([]Download, 0)
	for _, d := range all {
		if d.Status == StatusActive || d.Status == StatusPaused || d.Status == StatusPending || d.Status == StatusError {
			result = append(result, d)
		}
	}
	return result
}

// GetCompleted returns finished downloads.
func (m *Manager) GetCompleted() []Download {
	all := m.GetAll()
	result := make([]Download, 0)
	for _, d := range all {
		if d.Status == StatusCompleted {
			result = append(result, d)
		}
	}
	return result
}

// GetConfig returns the current config.
func (m *Manager) GetConfig() Config {
	m.cfgMu.RLock()
	defer m.cfgMu.RUnlock()
	return m.cfg
}

// UpdateConfig applies and persists new config values.
func (m *Manager) UpdateConfig(cfg Config) error {
	if cfg.MaxConcurrentDownloads < 1 {
		cfg.MaxConcurrentDownloads = 1
	}
	if cfg.MaxPartsPerDownload < 1 {
		cfg.MaxPartsPerDownload = 1
	}
	m.cfgMu.Lock()
	m.cfg = cfg
	// Resize semaphore by replacing it (active downloads keep their old slots)
	m.semaphore = make(chan struct{}, cfg.MaxConcurrentDownloads)
	m.cfgMu.Unlock()
	return cfg.Save()
}

func (m *Manager) getTask(id string) *downloadTask {
	m.tasksMu.RLock()
	defer m.tasksMu.RUnlock()
	return m.tasks[id]
}

// extractFileName derives a filename from a URL.
func extractFileName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "download"
	}
	base := filepath.Base(parsed.Path)
	base = strings.Split(base, "?")[0]
	if base == "" || base == "." || base == "/" {
		return "download_" + time.Now().Format("20060102_150405")
	}

	// Check if the filename has no extension and it's likely a direct file
	if !strings.Contains(base, ".") {
		// Try to get from Content-Disposition via HEAD later; use as-is for now
		return base
	}
	return base
}

// probeURL sends a HEAD request to determine file size and multipart support.
func probeURL(ctx context.Context, rawURL string) (totalSize int64, supportsRange bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("User-Agent", probeUserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	totalSize = resp.ContentLength
	supportsRange = strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes")
	return totalSize, supportsRange, nil
}
