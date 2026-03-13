package progress

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	Total      int64
	Downloaded int64
	startTime  time.Time
}

func New(total int64) *Progress {
	return &Progress{
		Total:      total,
		Downloaded: 0,
		startTime:  time.Now(),
	}
}

func (p *Progress) Add(n int64) {
	atomic.AddInt64(&p.Downloaded, n)
}

func (p *Progress) Print() {
	downloaded := atomic.LoadInt64(&p.Downloaded)

	percent := float64(downloaded) / float64(p.Total) * 100

	elapsed := time.Since(p.startTime).Seconds()

	speed := float64(downloaded) / elapsed / (1024 * 1024)

	fmt.Printf("\rDownloading: %.2f%% | Speed: %.2f MB/s", percent, speed)
}
