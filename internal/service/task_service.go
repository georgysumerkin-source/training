package service

import (
	"context"

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
	taskData, err := domain.NewTask(id, title, description)
	if err != nil {
		return nil, err
	}

	if err = s.taskRepository.Create(ctx, *taskData); err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	taskData, err := s.taskRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) GetTaskList(ctx context.Context) ([]*domain.Task, error) {
	taskData, err := s.taskRepository.GetList(ctx)
	if err != nil {
		return nil, err
	}

	return taskData, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	return s.taskRepository.Delete(ctx, id)
}

func (s *TaskService) ChangeTaskStatus(ctx context.Context, id string, status string) (*domain.Task, error) {
	taskData, err := s.taskRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = taskData.SetStatus(status); err != nil {
		return nil, err
	}

	if err = s.taskRepository.Update(ctx, *taskData); err != nil {
		return nil, err
	}

	return taskData, nil
}
