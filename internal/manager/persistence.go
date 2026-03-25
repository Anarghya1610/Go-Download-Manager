package manager

import (
	"encoding/json"
	"os"
)

const taskFile = "tasks.json"

type PersistedTask struct {
	URL    string
	Output string
}

func (m *Manager) SaveTasks() error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	var list []PersistedTask

	for _, t := range m.tasks {
		list = append(list, PersistedTask{
			URL:    t.URL,
			Output: t.Output,
		})
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(taskFile, data, 0644)
}

func (m *Manager) LoadTasks() error {

	data, err := os.ReadFile(taskFile)
	if err != nil {
		return nil // no file = first run
	}

	var list []PersistedTask
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	for _, t := range list {
		m.AddTask(t.URL, t.Output)
	}

	return nil
}
