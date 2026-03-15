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

func (p *Progress) SetResume(n int64, startedAt int64) {
	p.startTime = time.Unix(startedAt, 0)
	atomic.StoreInt64(&p.Downloaded, n)
}

func (p *Progress) Print() {
	downloaded := atomic.LoadInt64(&p.Downloaded)

	percent := 0.0
	if p.Total > 0 {
		percent = float64(downloaded) / float64(p.Total) * 100
	}

	now := time.Now()
	elapsed := now.Sub(p.lastTime).Seconds()
	totalTimeElapsed := now.Sub(p.startTime).Seconds()

	bytes := downloaded - p.lastBytes

	speed := 0.0
	if elapsed > 0 {
		speed = float64(bytes) / elapsed / (1024 * 1024)
	}

	averageSpeed := 0.0
	if totalTimeElapsed > 0 {
		averageSpeed = float64(downloaded) / totalTimeElapsed / (1024 * 1024)
	}

	p.lastBytes = downloaded
	p.lastTime = now

	fmt.Printf("\rDownloading: %.2f%% | Speed: %.2f MB/s | Average Speed: %.2f MB/s | Time elapsed: %.0fs", percent, speed, averageSpeed, totalTimeElapsed)
}
