package downloader

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Anarghya1610/godownloader/pkg/progress"
)

func getFileSize(url string, client *http.Client) (int64, error) {
	resp, err := client.Head(url)
	if err == nil {
		defer resp.Body.Close()

		if resp.ContentLength > 0 {
			return resp.ContentLength, nil
		}
	}

	// fallback to range request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Range", "bytes=0-0")

	resp, err = client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	cr := resp.Header.Get("Content-Range")

	var start, end, size int64
	_, err = fmt.Sscanf(cr, "bytes %d-%d/%d", &start, &end, &size)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Content-Range header")
	}

	return size, nil
}

func Download(url string, output string) error {
	var client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			DisableCompression:  true,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   false, // disable HTTP/2
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
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

	file.Truncate(size)

	// Start progress display
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				prog.Print()
				time.Sleep(time.Second)
			}
		}
	}()

	chunkQueue := make(chan Chunk, 100)

	numWorkers := 4

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go Worker(client, url, file, prog, chunkQueue, &wg)
	}

	chunkSize := int64(4 * 1024 * 1024) // 4 MB

	for start := int64(0); start < size; start += chunkSize {

		end := start + chunkSize - 1

		if end >= size {
			end = size - 1
		}

		chunkQueue <- Chunk{
			Start: start,
			End:   end,
		}
	}
	close(chunkQueue)

	wg.Wait()
	close(stop)

	fmt.Println("Download completed successfully")

	return nil
}
