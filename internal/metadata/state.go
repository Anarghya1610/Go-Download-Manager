package metadata

import "sync"

type ChunkState struct {
	Start      int64 `json:"start"`
	End        int64 `json:"end"`
	Downloaded int64 `json:"downloaded"`

	Parent *DownloadState `json:"-"`
}

type DownloadState struct {
	URL       string       `json:"url"`
	Output    string       `json:"output"`
	Size      int64        `json:"size"`
	Chunks    []ChunkState `json:"chunks"`
	StartedAt int64        `json:"started_at"`

	Mu sync.Mutex `json:"-"`
}
