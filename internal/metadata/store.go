package metadata

import (
	"encoding/json"
	"os"
)

func Save(path string, state *DownloadState) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(state)
}

func Load(path string) (*DownloadState, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var state DownloadState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&state)

	return &state, err
}
