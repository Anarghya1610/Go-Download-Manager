package main

import (
	"fmt"
	"os"

	"github.com/Anarghya1610/godownloader/internal/downloader"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("usage: godownloader <url> <output>")
		return
	}

	url := os.Args[1]
	output := os.Args[2]

	err := downloader.Download(url, output)
	if err != nil {
		fmt.Println("error:", err)
	}
}
