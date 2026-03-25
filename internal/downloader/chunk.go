package downloader

import (
	"github.com/Anarghya1610/gdm/internal/metadata"
)

type Chunk struct {
	Start int64
	End   int64
	State *metadata.ChunkState
}
