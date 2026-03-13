package downloader

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Anarghya1610/godownloader/pkg/progress"
)

func Download(url string, output string) error {
	resp, err := http.Head(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength
	fmt.Println("File size:", size, "bytes")

	prog := progress.New(size)

	file, err := os.Create(output)
	if err != nil {
		return err
	}

	defer file.Close()

	file.Truncate(size)

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
			err := DownloadChunk(url, c, file, prog)
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
