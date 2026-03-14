package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Anarghya1610/godownloader/internal/utils"
	"github.com/Anarghya1610/godownloader/pkg/progress"
)

func Download(url string, output string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        256,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			DisableCompression:  true,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}

	// Get file size
	size, err := getFileSize(url, client)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fmt.Println("File size:", size, "bytes")

	// Initialize progress
	prog := progress.New(size)

	// Create output file
	file, err := os.Create(output)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := file.Truncate(size); err != nil {
		return fmt.Errorf("failed to pre-allocate file: %w", err)
	}

	// Start progress display
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				prog.Print()
			}
		}
	}()

	supportsRange, err := serverSupportsRange(ctx, client, url)
	if err != nil {
		close(stop)
		return err
	}

	if !supportsRange {
		fmt.Println("Range requests are not available for this URL; falling back to single-stream download")
		if err := DownloadSingle(ctx, client, url, file, prog); err != nil {
			close(stop)
			return err
		}
		close(stop)
		fmt.Println("Download completed successfully")
		return nil
	}

	numWorkers := utils.DecideWorkers(size)
	if override := os.Getenv("GODOWNLOADER_WORKERS"); override != "" {
		v, parseErr := strconv.Atoi(override)
		if parseErr == nil && v > 0 {
			numWorkers = v
		}
	}

	fmt.Println("Num_Workers:", numWorkers)

	chunkQueue := make(chan Chunk, numWorkers*2)
	errCh := make(chan error, numWorkers)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go Worker(ctx, cancel, client, url, file, prog, chunkQueue, errCh, &wg)
	}

	chunkSize := int64(16 * 1024 * 1024) // 16 MB

enqueueLoop:
	for start := int64(0); start < size; start += chunkSize {
		select {
		case <-ctx.Done():
			break enqueueLoop
		default:
		}

		end := start + chunkSize - 1

		if end >= size {
			end = size - 1
		}

		chunk := Chunk{
			Start: start,
			End:   end,
		}

		select {
		case <-ctx.Done():
			break enqueueLoop
		case chunkQueue <- chunk:
		}
	}

	close(chunkQueue)

	wg.Wait()
	close(stop)
	close(errCh)

	if err, ok := <-errCh; ok {
		return err
	}

	fmt.Println("Download completed successfully")

	return nil
}
