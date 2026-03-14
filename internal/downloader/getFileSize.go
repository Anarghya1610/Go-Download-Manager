package downloader

import (
	"context"
	"fmt"
	"net/http"
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

func serverSupportsRange(ctx context.Context, client *http.Client, url string) (bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPartialContent {
		return true, nil
	}

	return false, nil
}
