package repository

import (
	"sync"

	"github.com/georgysumerkin-source/training/internal/domain"
)

type MemoryRepository struct {
	storage map[string]domain.Task
	locker  sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		storage: make(map[string]domain.Task),
		locker:  sync.RWMutex{},
	}
}

//var _ TaskRepository = (*MemoryRepository)(nil)

func (r *MemoryRepository) Create(task domain.Task) error {
	if _, exist := r.storage[task.ID]; exist {
		return ErrTaskAlreadyExists
	}
	r.storage[task.ID] = task

	return nil
}

func (r *MemoryRepository) Update(task domain.Task) error {
	if _, exist := r.storage[task.ID]; !exist {
		return ErrTaskNotFound
	}
	r.storage[task.ID] = task

	return nil
}

func (r *MemoryRepository) Delete(taskID string) error {
	if _, exist := r.storage[taskID]; !exist {
		return ErrTaskNotFound
	}
	delete(r.storage, taskID)

	return nil
}

func (r *MemoryRepository) GetByID(taskID string) (*domain.Task, error) {
	taskData, exist := r.storage[taskID]
	if !exist {
		return nil, ErrTaskNotFound
	}

	return &taskData, nil
}

func (r *MemoryRepository) GetList() ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0, len(r.storage))
	for _, taskData := range r.storage {
		tasks = append(tasks, &taskData)
	}

	return tasks, nil
}
func (r *MemoryRepository) GetLocker() *sync.RWMutex {
	return &r.locker
}
