package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Anarghya1610/godownloader/pkg/progress"
)

func DownloadChunkWithRetry(client *http.Client, url string, chunk Chunk, file *os.File, prog *progress.Progress) error {

	var err error

	for i := 0; i < 3; i++ {

		err = DownloadChunk(client, url, chunk, file, prog)

		if err == nil {
			return nil
		}
		fmt.Println("Retrying chunk:", chunk, "error:", err)

		time.Sleep(1 * time.Second)
	}

	return err
}

func DownloadChunk(client *http.Client, url string, chunk Chunk, file *os.File, prog *progress.Progress) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", chunk.Start, chunk.End)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//fmt.Println("Protocol:", resp.Proto)

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("Server did not return partial content")
	}

	buffer := make([]byte, 512*1024)

	offset := chunk.Start

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			written, err := file.WriteAt(buffer[:n], offset)

			if err != nil {
				return err
			}

			if written != n {
				return fmt.Errorf("Written bytes mismatch: expected %d, got %d", n, written)
			}

			offset += int64(written)
			prog.Add(int64(n))
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}
