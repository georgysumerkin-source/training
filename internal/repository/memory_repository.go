package repository

import (
	"context"
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

func (r *MemoryRepository) Create(ctx context.Context, task domain.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, exist := r.storage[task.ID]; exist {
		return ErrTaskAlreadyExists
	}
	r.storage[task.ID] = task

	return nil
}

func (r *MemoryRepository) Update(ctx context.Context, task domain.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, exist := r.storage[task.ID]; !exist {
		return ErrTaskNotFound
	}
	r.storage[task.ID] = task

	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, exist := r.storage[taskID]; !exist {
		return ErrTaskNotFound
	}
	delete(r.storage, taskID)

	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, taskID string) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.locker.RLock()
	defer r.locker.RUnlock()

	taskData, exist := r.storage[taskID]
	if !exist {
		return nil, ErrTaskNotFound
	}

	return &taskData, nil
}

func (r *MemoryRepository) GetList(ctx context.Context) ([]*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.locker.RLock()
	defer r.locker.RUnlock()

	tasks := make([]*domain.Task, 0, len(r.storage))
	for _, taskData := range r.storage {
		tasks = append(tasks, &taskData)
	}
	return tasks, nil
}
