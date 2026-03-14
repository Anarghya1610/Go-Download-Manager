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
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   false, // disable HTTP/2
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
		},
	}

	size, err := getFileSize(url, client)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fmt.Println("File size:", size, "bytes")

	prog := progress.New(size)

	file, err := os.Create(output)
	if err != nil {
		return err
	}

	defer file.Close()

	//file.Truncate(size)

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

	numWorkers := 8
	chunkSize := size / int64(numWorkers)

	var chunks []Chunk

	for i := 0; i < numWorkers; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1

		if i == numWorkers-1 {
			end = size - 1
		}

		chunks = append(chunks, Chunk{Start: start, End: end})
	}

	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go func(c Chunk) {
			defer wg.Done()
			err := DownloadChunkWithRetry(client, url, c, file, prog)
			if err != nil {
				fmt.Println("Error downloading chunk:", err)
			}
		}(chunk)
	}

	wg.Wait()
	close(stop)

	fmt.Println("Download completed successfully")

	return nil
}
