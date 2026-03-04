package downloader

import "time"

// Status represents the current state of a download.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusError     Status = "error"
)

// Download is the public-facing snapshot of a download's state.
type Download struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	FileName   string    `json:"fileName"`
	TotalSize  int64     `json:"totalSize"`
	Downloaded int64     `json:"downloaded"`
	Speed      int64     `json:"speed"` // bytes per second
	Status     Status    `json:"status"`
	Parts      int       `json:"parts"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Percent returns completion percentage (0-100).
func (d Download) Percent() float64 {
	if d.TotalSize <= 0 {
		return 0
	}
	return float64(d.Downloaded) / float64(d.TotalSize) * 100
}
