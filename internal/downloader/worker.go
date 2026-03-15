package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Anarghya1610/godownloader/pkg/progress"
)

var ErrNoPartialContent = errors.New("server did not return partial content")

type RateLimitError struct {
	StatusCode int
	Range      string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited: status=%d, range=%s, retry-after=%s", e.StatusCode, e.Range, e.RetryAfter)
}

func Worker(ctx context.Context, cancel context.CancelFunc, client *http.Client, url string, file *os.File, prog *progress.Progress, chunkQueue <-chan Chunk, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	buffer := make([]byte, 1024*1024)

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-chunkQueue:
			if !ok {
				return
			}

			err := DownloadChunkWithRetry(ctx, client, url, chunk, file, prog, buffer)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("chunk %d-%d failed: %w", chunk.Start, chunk.End, err):
				default:
				}
				cancel()
				return
			}
		}
	}
}

func DownloadChunkWithRetry(ctx context.Context, client *http.Client, url string, chunk Chunk, file *os.File, prog *progress.Progress, buffer []byte) error {
	var err error

	for i := 0; i < 8; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = DownloadChunk(ctx, client, url, chunk, file, prog, buffer)
		if err == nil {
			return nil
		}

		var rateLimitErr *RateLimitError
		if errors.As(err, &rateLimitErr) {
			wait := rateLimitErr.RetryAfter
			wait += time.Duration(rand.Intn(1000)) * time.Millisecond
			if wait <= 0 {
				wait = time.Duration(1<<i) * time.Second
				if wait > 30*time.Second {
					wait = 30 * time.Second
				}
			}

			fmt.Printf("\nRate limited for chunk %d-%d, waiting %s before retry...", chunk.Start, chunk.End, wait)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		if errors.Is(err, ErrNoPartialContent) {
			return err
		}

		fmt.Println("Retrying chunk:", chunk, "error:", err)
		time.Sleep(time.Duration(1<<i) * 500 * time.Millisecond)
	}

	return err
}

func DownloadChunk(ctx context.Context, client *http.Client, url string, chunk Chunk, file *os.File, prog *progress.Progress, buffer []byte) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resumeStart := chunk.Start + chunk.State.Downloaded
	if resumeStart > chunk.End {
		return nil
	}
	rangeHeader := fmt.Sprintf("bytes=%d-%d", resumeStart, chunk.End)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//fmt.Println("protocol used: ", resp.Proto)

	if resp.StatusCode != http.StatusPartialContent {
		if resp.StatusCode == http.StatusTooManyRequests {
			return &RateLimitError{
				StatusCode: resp.StatusCode,
				Range:      rangeHeader,
				RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
			}
		}

		return fmt.Errorf("%w: status=%d, range=%s, content-range=%q", ErrNoPartialContent, resp.StatusCode, rangeHeader, resp.Header.Get("Content-Range"))
	}

	contentRange := resp.Header.Get("Content-Range")
	expectedPrefix := fmt.Sprintf("bytes %d-%d/", resumeStart, chunk.End)

	if !strings.HasPrefix(contentRange, expectedPrefix) {
		return fmt.Errorf("invalid content-range: got %s expected prefix %s", contentRange, expectedPrefix)
	}

	offset := resumeStart

	for {
		n, readErr := resp.Body.Read(buffer)

		remaining := chunk.End - offset + 1
		if int64(n) > remaining {
			n = int(remaining)
		}

		if n > 0 {
			written, writeErr := file.WriteAt(buffer[:n], offset)
			if writeErr != nil {
				return writeErr
			}

			if written != n {
				return fmt.Errorf("Written bytes mismatch: expected %d, got %d", n, written)
			}

			offset += int64(written)
			if chunk.State != nil {
				if chunk.State.Parent != nil {
					chunk.State.Parent.Mu.Lock()
					chunk.State.Downloaded += int64(written)
					chunk.State.Parent.Mu.Unlock()
				} else {
					chunk.State.Downloaded += int64(written)
				}
			}
			prog.Add(int64(written))
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return readErr
		}
	}

	expected := chunk.End - resumeStart + 1
	if chunk.State.Downloaded != expected {
		return fmt.Errorf("Chunk incomplete: expected %d got %d", expected, chunk.State.Downloaded)
	}

	return nil
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	seconds, err := time.ParseDuration(value + "s")
	if err == nil {
		return seconds
	}

	when, err := time.Parse(time.RFC1123, value)
	if err != nil {
		return 0
	}

	d := time.Until(when)
	if d < 0 {
		return 0
	}

	return d
}

func DownloadSingle(ctx context.Context, client *http.Client, url string, file *os.File, prog *progress.Progress) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("single stream download failed: status=%d", resp.StatusCode)
	}

	buffer := make([]byte, 1024*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			if written != n {
				return fmt.Errorf("written bytes mismatch: expected %d, got %d", n, written)
			}
			prog.Add(int64(written))
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
