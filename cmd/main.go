package main

import (
	"fmt"

	"github.com/Anarghya1610/godownloader/internal/manager"
)

func main() {

	// if len(os.Args) < 3 {
	// 	fmt.Println("Usage:")
	// 	fmt.Println("  godownloader <url> <output>")
	// 	return
	// }

	// url := os.Args[1]
	// output := os.Args[2]

	// create download manager (3 concurrent downloads)
	mgr := manager.NewManager(3)

	id1 := mgr.AddTask("https://trashbytes.net/dl/Vzz7Fa5YoyaqmZTkpjkZPKZvriPurOacosvGVQ59SQUEB38q067MqyAblen_Wl0n1HTiLPwXgxuwXXF63nMZx7JwbiRBD6Ljd-6alHLjQQD0zZ8kvjQsmTYY3Vce53f-e_E3K0AHcudmxoZ9qw?v=1773579372-pwOjmqCIjzX6Pf9RrC7mj0YApZMkKQ37n4t%2FEVJOgsI%3D", "file1.rar")
	id2 := mgr.AddTask("https://trashbytes.net/dl/NNLDYBOKcg7Yv_g7U_TL677VozmMqGUVqYS7R5Yjvgs5rMhdezJZLPc1so9DbEyG4KKfjwuDpV_fiOXRL8N5bvzTptWtDdkAjjBjCS4hXp5GNK62e9QB61ujKtn4L3AEGFM?v=1773577964-KMx9%2FqXXxtlEkiBAfRVgz19kiaM%2F6xvtBzYsmTuXEBA%3D", "file2.rar")

	fmt.Println("Download started with ID:", id1)
	fmt.Println("Download started with ID:", id2)

	select {} // keep program alive
}
