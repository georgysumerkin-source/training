package service

import (
	"context"
	"sync"

	"github.com/georgysumerkin-source/training/internal/common"
	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/repository"
)

type TaskService struct {
	taskRepository repository.TaskRepository
	locker         sync.RWMutex
}

func NewTaskService(taskRepository repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, id, title, description string) (*domain.Task, error) {
	fn := func() (*domain.Task, error) {
		taskData, err := domain.NewTask(id, title, description)
		if err != nil {
			return nil, err
		}
		if err = s.taskRepository.Create(*taskData); err != nil {
			return nil, err
		}

		return taskData, nil
	}

	taskData, err := common.WithLockResult(ctx, &s.locker, fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	fn := func() (*domain.Task, error) {
		taskData, err := s.taskRepository.GetByID(id)
		if err != nil {
			return nil, err
		}

		return taskData, nil
	}
	taskData, err := common.WithLockResult(ctx, s.locker.RLocker(), fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) GetTaskList(ctx context.Context) ([]*domain.Task, error) {
	fn := func() ([]*domain.Task, error) {
		taskData, err := s.taskRepository.GetList()
		if err != nil {
			return nil, err
		}

		return taskData, nil
	}
	taskData, err := common.WithLockResult(ctx, s.locker.RLocker(), fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	fn := func() error {
		return s.taskRepository.Delete(id)
	}

	return common.WithLock(ctx, &s.locker, fn)
}

func (s *TaskService) ChangeTaskStatus(ctx context.Context, id string, status string) (*domain.Task, error) {
	fn := func() (*domain.Task, error) {
		taskData, err := s.taskRepository.GetByID(id)
		if err != nil {
			return nil, err
		}
		if err = taskData.SetStatus(status); err != nil {
			return nil, err
		}
		if err = s.taskRepository.Update(*taskData); err != nil {
			return nil, err
		}

		return taskData, nil
	}

	taskData, err := common.WithLockResult(ctx, &s.locker, fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}
