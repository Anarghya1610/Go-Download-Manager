package main

import (
	"fmt"
	"os"

	"github.com/Anarghya1610/gdm/internal/manager"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  gdm <url> <output>")
		return
	}

	url := os.Args[1]
	output := os.Args[2]

	mgr := manager.NewManager(3)

	id1 := mgr.AddTask(url, output, true)
	fmt.Println("Download started with ID:", id1)

	select {} // keep program alive
}
