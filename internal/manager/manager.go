package manager

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	tasks map[string]*DownloadTask
	queue chan *DownloadTask

	maxConcurrent int
	Mu            sync.Mutex
}

func NewManager(maxConcurrent int) *Manager {

	m := &Manager{
		tasks:         make(map[string]*DownloadTask),
		queue:         make(chan *DownloadTask, 100),
		maxConcurrent: maxConcurrent,
	}

	for i := 0; i < maxConcurrent; i++ {
		go m.worker()
	}

	return m
}

func (m *Manager) worker() {
	for task := range m.queue {
		task.Start()
	}
}

func (m *Manager) AddTask(url string, output string) string {
	id := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())

	task := &DownloadTask{
		ID:     id,
		URL:    url,
		Output: output,
		Status: StatusQueued,
		ctx:    ctx,
		cancel: cancel,
	}
	m.Mu.Lock()
	m.tasks[id] = task
	m.Mu.Unlock()

	select {
	case m.queue <- task:
	default:
		go func() { m.queue <- task }()
	}

	return id
}

func (m *Manager) Pause(id string) {
	m.Mu.Lock()
	task, exists := m.tasks[id]
	m.Mu.Unlock()

	if !exists {
		return
	}

	task.Mu.Lock()
	defer task.Mu.Unlock()

	if task.Status == StatusRunning {
		task.cancel()
		task.Status = StatusPaused
	}
}

func (m *Manager) Resume(id string) {
	m.Mu.Lock()
	task, exists := m.tasks[id]
	m.Mu.Unlock()

	if !exists {
		return
	}

	task.Mu.Lock()
	defer task.Mu.Unlock()

	if task.Status == StatusPaused {
		ctx, cancel := context.WithCancel(context.Background())
		task.ctx = ctx
		task.cancel = cancel
		task.Status = StatusQueued

		select {
		case m.queue <- task:
		default:
			go func() { m.queue <- task }()
		}
	}
}

func (m *Manager) List() []*DownloadTask {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	list := make([]*DownloadTask, 0, len(m.tasks))

	for _, t := range m.tasks {
		list = append(list, t)
	}
	return list
}
