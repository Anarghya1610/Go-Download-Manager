package manager

import (
	"context"
	"sync"

	"github.com/Anarghya1610/godownloader/internal/downloader"
	"github.com/Anarghya1610/godownloader/pkg/progress"
)

type TaskStatus string

const (
	StatusQueued    TaskStatus = "queued"
	StatusRunning   TaskStatus = "running"
	StatusPaused    TaskStatus = "paused"
	StatusCompleted TaskStatus = "completed"
	StatusError     TaskStatus = "error"
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
}

func (t *DownloadTask) Start() {

	go func() {

		t.Mu.Lock()
		t.Status = StatusRunning
		t.Mu.Unlock()

		err := downloader.Download(t.ctx, t.URL, t.Output)

		t.Mu.Lock()
		defer t.Mu.Unlock()

		if err != nil {
			t.Status = StatusError
			t.Error = err.Error()
		} else {
			t.Status = StatusCompleted
		}

	}()

}
