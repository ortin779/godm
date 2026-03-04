package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func stateFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "godm", "state.json")
}

// saveState serialises all current download records to disk.
// It is safe to call from a goroutine.
func (m *Manager) saveState() {
	m.tasksMu.RLock()
	records := make([]Download, 0, len(m.tasks))
	for _, t := range m.tasks {
		records = append(records, t.snapshot())
	}
	m.tasksMu.RUnlock()

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return
	}
	p := stateFilePath()
	_ = os.MkdirAll(filepath.Dir(p), 0755)
	_ = os.WriteFile(p, data, 0644)
}

// loadState reads persisted download records and re-populates the task map.
// Downloads that were active/pending are restored as paused so the user can
// choose to resume them manually.
func (m *Manager) loadState() {
	data, err := os.ReadFile(stateFilePath())
	if err != nil {
		return // no saved state yet — that's fine
	}
	var records []Download
	if err := json.Unmarshal(data, &records); err != nil {
		return
	}
	for _, d := range records {
		// Active/pending downloads lost their goroutines on shutdown → paused.
		if d.Status == StatusActive || d.Status == StatusPending {
			d.Status = StatusPaused
		}
		// Cancelled downloads are excluded from the restored list.
		if d.Status == StatusCancelled {
			continue
		}
		m.tasks[d.ID] = &downloadTask{download: d}
	}
}
