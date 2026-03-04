package main

import (
	"context"

	"godm/downloader"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx     context.Context
	manager *downloader.Manager
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. Initialises the download manager.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg := downloader.LoadConfig()
	a.manager = downloader.NewManager(ctx, cfg, func(eventName string, data ...interface{}) {
		runtime.EventsEmit(ctx, eventName, data...)
	})
}

// ProbeDownload fetches metadata for a URL without downloading the file.
func (a *App) ProbeDownload(url string) (downloader.DownloadInfo, error) {
	return a.manager.ProbeDownload(url)
}

// AddDownload enqueues a new download for the given URL and confirmed file name.
func (a *App) AddDownload(url, fileName string) (downloader.Download, error) {
	return a.manager.Add(url, fileName)
}

// PauseDownload pauses an active download.
func (a *App) PauseDownload(id string) error {
	return a.manager.Pause(id)
}

// ResumeDownload resumes a paused download.
func (a *App) ResumeDownload(id string) error {
	return a.manager.Resume(id)
}

// CancelDownload cancels and removes a download.
func (a *App) CancelDownload(id string) error {
	return a.manager.Cancel(id)
}

// RevealFile opens the downloaded file in the system file manager.
func (a *App) RevealFile(id string) error {
	return a.manager.RevealFile(id)
}

// GetDownloads returns all active/pending/paused/errored downloads.
func (a *App) GetDownloads() []downloader.Download {
	return a.manager.GetActive()
}

// GetCompletedDownloads returns all completed downloads.
func (a *App) GetCompletedDownloads() []downloader.Download {
	return a.manager.GetCompleted()
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() downloader.Config {
	return a.manager.GetConfig()
}

// UpdateConfig applies and persists new configuration.
func (a *App) UpdateConfig(cfg downloader.Config) (downloader.Config, error) {
	if err := a.manager.UpdateConfig(cfg); err != nil {
		return a.manager.GetConfig(), err
	}
	return a.manager.GetConfig(), nil
}
