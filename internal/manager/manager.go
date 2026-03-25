package manager

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	tasks map[string]*DownloadTask
	order []string
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

func (m *Manager) GetTasks() []*DownloadTask {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	tasks := make([]*DownloadTask, 0, len(m.order))

	for _, id := range m.order {
		if t, ok := m.tasks[id]; ok {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (m *Manager) AddTask(url string, output string, autoStart bool) string {
	id := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())

	status := StatusPaused
	if autoStart {
		status = StatusQueued
	}

	task := &DownloadTask{
		ID:     id,
		URL:    url,
		Output: output,
		Status: status,
		ctx:    ctx,
		cancel: cancel,
	}

	m.Mu.Lock()
	m.tasks[id] = task
	m.order = append(m.order, id)
	m.Mu.Unlock()

	if autoStart {
		select {
		case m.queue <- task:
		default:
			go func() { m.queue <- task }()
		}
	}

	m.SaveTasks()

	return id
}

func (m *Manager) StartTask(id string) {
	m.Mu.Lock()
	task, exists := m.tasks[id]
	m.Mu.Unlock()

	if !exists {
		return
	}

	task.Mu.Lock()
	task.Status = StatusRunning
	task.Mu.Unlock()

	select {
	case m.queue <- task:
	default:
		go func() { m.queue <- task }()
	}
}

func (m *Manager) Pause(id string) {
	m.Mu.Lock()
	task, exists := m.tasks[id]
	m.Mu.Unlock()

	if !exists {
		return
	}

	task.Mu.Lock()
	task.isPaused = true // ✅ mark intent
	task.Status = StatusPaused
	task.Mu.Unlock()

	task.cancel()
}

func (m *Manager) Cancel(id string) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	task, exists := m.tasks[id]
	if !exists {
		return
	}
	task.cancel()
	delete(m.tasks, id)
	// ✅ remove from order slice
	for i, v := range m.order {
		if v == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
}

func (m *Manager) Resume(id string) {
	m.Mu.Lock()
	task, exists := m.tasks[id]
	m.Mu.Unlock()

	if !exists {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	task.Mu.Lock()
	task.ctx = ctx
	task.cancel = cancel
	task.isPaused = false // ✅ reset flag
	task.Status = StatusRunning
	task.Mu.Unlock()

	select {
	case m.queue <- task:
	default:
		go func() { m.queue <- task }()
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
