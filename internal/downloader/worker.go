package downloader

import(
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadChunk(url string, chunk Chunk, file *os.File) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	rangeHeader := fmt.Sprintf("bytes="%d-%d"", chunk.Start, chunk.End)
	req.Header.Set("Range", rangeHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer := make([]byte, 32*1024)

	offset := chunk.Start

	for{
		n, err := resp.Body.Read(buffer)

		if(n > 0){
			file.WriteAt(buffer[:n], offset)
			offset += int64(n)
		}

		if err != nil {
			return err
		}

		if(err == io.EOF){
			break
		}
	}

	return nil
}