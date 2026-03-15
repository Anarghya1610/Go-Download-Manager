package downloader

import (
	"github.com/Anarghya1610/godownloader/internal/metadata"
)

type Chunk struct {
	Start      int64
	End        int64
	Downloaded int64
	State      *metadata.ChunkState
}
