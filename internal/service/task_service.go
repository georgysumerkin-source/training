package service

import (
	"context"

	"github.com/georgysumerkin-source/training/internal/common"
	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/repository"
)

type TaskService struct {
	taskRepository repository.TaskRepository
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
	locker := s.taskRepository.GetLocker()
	taskData, err := common.WithLockResult(ctx, locker, fn)
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
	locker := s.taskRepository.GetLocker()
	taskData, err := common.WithLockResult(ctx, locker.RLocker(), fn)
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
	locker := s.taskRepository.GetLocker()
	taskData, err := common.WithLockResult(ctx, locker.RLocker(), fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	fn := func() error {
		return s.taskRepository.Delete(id)
	}
	locker := s.taskRepository.GetLocker()

	return common.WithLock(ctx, locker, fn)
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

	locker := s.taskRepository.GetLocker()
	taskData, err := common.WithLockResult(ctx, locker, fn)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}
