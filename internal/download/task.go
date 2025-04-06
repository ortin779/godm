package download

import (
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
)

type Task struct {
	Url         string    `json:"url"`
	FilePath    string    `json:"file_path"`
	Parts       []*Part   `json:"parts"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	Status      Status    `json:"status"`
	ContentSize int64     `json:"content_size"`
	Error       string    `json:"error"`
}

func NewTask(url, filePath string, noOfParts int) (*Task, error) {
	metadata, err := getMetadata(url)
	if err != nil {
		return nil, err
	}

	return &Task{
		Url:         url,
		FilePath:    filePath,
		CreatedAt:   time.Now(),
		Status:      Pending,
		Parts:       generateParts(url, metadata.ContentSize, noOfParts),
		ContentSize: metadata.ContentSize,
	}, nil
}

func getMetadata(url string) (Metadata, error) {
	metadata := Metadata{}
	resp, err := http.DefaultClient.Head(url)
	if err != nil {
		return metadata, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return metadata, errors.New("received non-OK status code")
	}

	acceptRanges := resp.Header.Get("accept-ranges")

	metadata.ContentSize = resp.ContentLength
	metadata.IsRangeSupported = acceptRanges != ""

	return metadata, nil
}

func generateParts(url string, totalSize int64, noOfParts int) []*Part {
	parts := make([]*Part, noOfParts)
	partSize := totalSize / int64(noOfParts)

	for i := 0; i < noOfParts; i++ {
		start := i * int(partSize)
		end := (i + 1) * int(partSize)
		if end > int(totalSize) {
			end = int(totalSize)
		}
		parts[i] = NewPart(url, start, end)
	}

	return parts
}

func (t *Task) Start() error {
	if t.Status == InProgress {
		return errors.New("the task is already running")
	}

	wg := sync.WaitGroup{}
	wg.Add(len(t.Parts))
	errs := make([]error, 0)

	defer t.cleanUp()

	t.Status = InProgress
	for _, part := range t.Parts {
		go func() {
			defer wg.Done()

			err := part.Download()
			if err != nil {
				errs = append(errs, err)
			}

		}()
	}

	wg.Wait()
	if len(errs) > 0 {
		t.Status = Failed
		return errs[len(errs)-1]
	}

	err := t.mergeParts()
	if err != nil {
		return err
	}

	t.Status = Completed
	return nil
}

func (t *Task) mergeParts() error {
	f, err := os.OpenFile(t.FilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, part := range t.Parts {
		bytes, err := part.Data()
		if err != nil {
			return err
		}
		_, err = f.WriteAt(bytes, int64(part.Start))
		if err != nil {
			return nil
		}
	}

	return nil
}

func (t *Task) cleanUp() {
	for _, part := range t.Parts {
		part.Remove()
	}
}
