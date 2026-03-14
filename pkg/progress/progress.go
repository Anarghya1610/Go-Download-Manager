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

	lastBytes int64
	lastTime  time.Time
}

func New(total int64) *Progress {
	now := time.Now()
	return &Progress{
		Total:     total,
		startTime: now,
		lastTime:  now,
	}
}

func (p *Progress) Add(n int64) {
	atomic.AddInt64(&p.Downloaded, n)
}

func (p *Progress) Print() {
	downloaded := atomic.LoadInt64(&p.Downloaded)

	percent := float64(downloaded) / float64(p.Total) * 100

	now := time.Now()
	elapsed := now.Sub(p.lastTime).Seconds()

	bytes := downloaded - p.lastBytes

	speed := float64(bytes) / elapsed / (1024 * 1024)

	p.lastBytes = downloaded
	p.lastTime = now

	fmt.Printf("\rDownloading: %.2f%% | Speed: %.2f MB/s", percent, speed)
}
