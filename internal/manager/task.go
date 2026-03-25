package manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Anarghya1610/gdm/internal/downloader"
	"github.com/Anarghya1610/gdm/pkg/progress"
)

type TaskStatus string

const (
	StatusQueued    TaskStatus = "queued"
	StatusRunning   TaskStatus = "running"
	StatusPaused    TaskStatus = "paused"
	StatusCompleted TaskStatus = "completed"
	StatusError     TaskStatus = "error"
	StatusCancelled TaskStatus = "cancelled"
)

type DownloadTask struct {
	ID     string
	URL    string
	Output string
	Status TaskStatus
	Error  string

	ctx    context.Context
	cancel context.CancelFunc

	progress *progress.Progress

	Mu sync.Mutex

	isPaused bool
}

func (t *DownloadTask) Start() {

	t.Mu.Lock()
	t.Status = StatusRunning
	t.Mu.Unlock()

	err := downloader.Download(t.ctx, t.URL, t.Output, func(p *progress.Progress) {
		t.Mu.Lock()
		t.progress = p
		t.Mu.Unlock()
	})

	t.Mu.Lock()
	defer t.Mu.Unlock()

	if err != nil {
		if t.ctx.Err() == context.Canceled {
			// DO NOT treat as error EVER
			return
		}

		// real error
		t.Status = StatusError
		t.Error = err.Error()
		return
	}

	t.Status = StatusCompleted
}

func (t *DownloadTask) GetProgress() float64 {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if t.progress == nil || t.progress.Total == 0 {
		return 0
	}
	downloaded := atomic.LoadInt64(&t.progress.Downloaded)
	return float64(downloaded) / float64(t.progress.Total)
}

func (t *DownloadTask) GetStatus() string {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	return string(t.Status)
}

func (t *DownloadTask) GetSpeed() string {
	t.Mu.Lock()
	p := t.progress
	t.Mu.Unlock()

	if p == nil {
		return ""
	}

	return fmt.Sprintf("%.2f MB/s", p.GetSpeedMB())
}
