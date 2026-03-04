# goDM — Go Download Manager

A fast, native desktop download manager built with [Wails v2](https://wails.io) (Go + WebKit). Supports multipart parallel downloading, pause/resume/cancel, real-time progress tracking, and persistent state across restarts.

## Features

### Downloading
- **2-step add dialog** — paste a URL, preview the detected filename and file size before starting
- **Multipart downloads** — automatically splits files into parallel parts (default: 4 parts) when the server supports `Accept-Ranges`, maximizing download speed
- **Single-part fallback** — seamlessly falls back to a single connection for servers that don't support range requests
- **Concurrent downloads** — configurable limit on how many downloads run simultaneously (default: 3)
- **Resume on restart** — active downloads are saved as paused and can be resumed after the app restarts

### Controls
- **Pause / Resume** — suspend and continue any active or pending download
- **Cancel** — stop a download and clean up all temporary part files
- **Show in Finder / Explorer** — reveal the completed file in the system file manager

### URL Handling
- **Redirect following** — automatically follows HTTP redirects to reach the real file
- **Google Drive** — normalizes `/file/d/ID/view` share links to direct download URLs
- **Bot-protection detection** — detects HTML responses (Cloudflare challenges, login pages, viewer pages) and shows a clear error instead of writing corrupted data to disk
- **Content-Disposition** — extracts the real filename from server headers (supports RFC 5987 encoding)

### UI
- **Real-time progress** — progress bar, percentage, downloaded/total size, and live download speed updated every 500 ms
- **All / Completed tabs** — separate views for in-progress and finished downloads
- **Search bar** — filter downloads by filename across both tabs
- **Status badges** — visual indicators for active, paused, pending, error, and completed states
- **Settings panel** — slide-over panel to adjust max concurrent downloads, parts per download, and the default download directory

### Persistence
- Download state is saved to `~/.config/godm/state.json`
- Configuration is saved to `~/.config/godm/godm_config.json`
- Downloads that were active when the app closed are restored as paused

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Desktop runtime | [Wails v2](https://wails.io) |
| Backend | Go |
| Frontend | React + TypeScript |
| Styling | Tailwind CSS v3 |
| UI components | Shadcn UI pattern (Radix UI primitives) |
| Component design | Atomic design (atoms → molecules → organisms) |

## Project Structure

```
.
├── app.go                  # Wails app — exposes Go methods to the frontend
├── main.go                 # Entry point
├── downloader/
│   ├── config.go           # Config struct, load/save to disk
│   ├── download.go         # Download struct, Status constants
│   ├── manager.go          # Manager: add, pause, resume, cancel, reveal
│   ├── multipart.go        # HTTP range download, part merging, progress tracking
│   ├── probe.go            # URL metadata probe, Google Drive normalization
│   └── store.go            # State persistence (save/load JSON)
└── frontend/
    └── src/
        ├── hooks/
        │   ├── useDownloads.ts     # State management + Wails event subscription
        │   └── useWailsEvents.ts   # Generic EventsOn wrapper
        ├── lib/
        │   ├── types.ts            # TypeScript types matching Go structs
        │   └── utils.ts            # cn(), formatBytes(), formatSpeed()
        └── components/
            ├── atoms/              # DownloadProgress, FileSize
            ├── molecules/          # DownloadItem, AddDownloadDialog, SearchBar
            └── organisms/          # DownloadList, DownloadTabs, SettingsPanel
```

## Development

```bash
# Run in dev mode with hot reload
wails dev

# Build a production binary
wails build

# Regenerate Wails JS bindings after Go struct changes
wails generate module
```

> **Node.js note**: if `npm` is not in your PATH, prefix commands with the full path, e.g.
> `export PATH="$HOME/.nvm/versions/node/v24.14.0/bin:$PATH"`

## Default Configuration

| Setting | Default |
|---------|---------|
| Max concurrent downloads | 3 |
| Max parts per download | 4 |
| Download directory | `~/Downloads` |

## Supported Platforms

| Platform | Reveal file command |
|----------|-------------------|
| macOS | `open -R <file>` |
| Windows | `explorer /select, <file>` |
| Linux | `xdg-open <directory>` |
