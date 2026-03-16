package downloader

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Anarghya1610/godownloader/internal/metadata"
	"github.com/Anarghya1610/godownloader/internal/utils"
	"github.com/Anarghya1610/godownloader/pkg/progress"
)

func Download(ctx context.Context, url string, output string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        256,
			MaxConnsPerHost:     64,
			MaxIdleConnsPerHost: 64,
			DisableCompression:  true,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   false,
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
		},
	}

	metaPath := output + ".meta"

	var state *metadata.DownloadState
	var size int64
	var err error

	if _, err := os.Stat(metaPath); err == nil {

		fmt.Println("Resuming download...")

		state, err = metadata.Load(metaPath)
		if err != nil {
			return err
		}

		size = state.Size

	} else {

		size, err = getFileSize(url, client)
		if err != nil {
			return fmt.Errorf("failed to get file size: %w", err)
		}
	}

	fmt.Println("File size:", size, "bytes")

	// Initialize progress
	prog := progress.New(size)
	if state != nil {
		var downloaded int64
		for _, c := range state.Chunks {
			downloaded += c.Downloaded
		}
		prog.SetResume(downloaded)
	}

	if state == nil {
		var chunks []metadata.ChunkState

		chunkSize := int64(16 * 1024 * 1024)

		for start := int64(0); start < size; start += chunkSize {
			end := start + chunkSize - 1
			if end >= size {
				end = size - 1
			}
			chunks = append(chunks, metadata.ChunkState{
				Start:      start,
				End:        end,
				Downloaded: 0,
			})
		}

		state = &metadata.DownloadState{
			URL:       url,
			Output:    output,
			Size:      size,
			Chunks:    chunks,
			StartedAt: time.Now().Unix(),
		}
	}

	// Re-link parent pointers after creating/loading metadata state.
	for i := range state.Chunks {
		state.Chunks[i].Parent = state
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				state.Mu.Lock()
				err := metadata.Save(metaPath, state)
				state.Mu.Unlock()
				if err != nil {
					fmt.Println("metadata save error:", err)
				}
			}
		}
	}()

	// Create output file
	file, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := file.Truncate(size); err != nil {
		return fmt.Errorf("Failed to pre-allocate file: %w", err)
	}

	// Start progress display
	stop := make(chan struct{})

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

	chunkQueue := make(chan Chunk, numWorkers*4)
	errCh := make(chan error, numWorkers)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go Worker(ctx, cancel, client, url, file, prog, chunkQueue, errCh, &wg)
	}

enqueueLoop:
	for i := range state.Chunks {
		c := state.Chunks[i]
		chunkSize := c.End - c.Start + 1
		if c.Downloaded < chunkSize {
			chunk := Chunk{
				Start: c.Start,
				End:   c.End,
				State: &state.Chunks[i],
			}
			select {
			case <-ctx.Done():
				break enqueueLoop
			case chunkQueue <- chunk:
			}
		}
	}

	close(chunkQueue)

	wg.Wait()
	close(stop)
	close(errCh)

	if err, ok := <-errCh; ok {
		cancel()
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() != size {
		return fmt.Errorf("File corrupted: expected size %d got %d", size, stat.Size())
	}
	fmt.Println("Download completed successfully")

	cancel()
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		fmt.Println("metadata cleanup error:", err)
	}

	return nil
}
