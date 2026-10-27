package repository

import (
	"context"

	"github.com/georgysumerkin-source/training/internal/domain"
)

type TaskRepository interface {
	Create(context.Context, domain.Task) error
	GetByID(context.Context, string) (*domain.Task, error)
	GetList(context.Context) ([]*domain.Task, error)
	Update(context.Context, domain.Task) error
	Delete(context.Context, string) error
}
