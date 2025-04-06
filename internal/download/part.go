package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/ortin779/godm/internal/config"
)

type Part struct {
	Id     string `json:"id"`
	Url    string `json:"url"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Status Status `json:"status"`
}

func NewPart(url string, start, end int) *Part {
	return &Part{
		Id:     uuid.NewString(),
		Url:    url,
		Start:  start,
		End:    end,
		Status: Pending,
	}
}

func (p *Part) Download() error {
	if p.Status == InProgress {
		return nil
	}

	log.Printf("Started downloading %v", p.Id)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDirPath := filepath.Join(homeDir, config.BaseDir)
	partPath := filepath.Join(baseDirPath, p.Id+".part")
	f, err := os.OpenFile(partPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodGet, p.Url, nil)
	if err != nil {
		return err
	}
	rangeHeader := fmt.Sprintf("Range: bytes %d-%d", p.Start, p.End)
	req.Header.Add("Range", rangeHeader)

	p.Status = InProgress
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.Status = Failed
		return nil
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		p.Status = Failed
		return err
	}
	p.Status = Completed

	log.Printf("Download completed %v", p.Id)
	return nil
}

func (p *Part) Remove() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDirPath := filepath.Join(homeDir, config.BaseDir)
	partPath := filepath.Join(baseDirPath, p.Id+".part")
	err = os.Remove(partPath)
	if err != nil {
		return err
	}

	return nil
}

func (p *Part) Data() ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDirPath := filepath.Join(homeDir, config.BaseDir)
	partPath := filepath.Join(baseDirPath, p.Id+".part")
	f, err := os.OpenFile(partPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}
