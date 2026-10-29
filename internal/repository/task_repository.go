package repository

import (
	"sync"

	"github.com/georgysumerkin-source/training/internal/domain"
)

type TaskRepository interface {
	Create(domain.Task) error
	GetByID(string) (*domain.Task, error)
	GetList() ([]*domain.Task, error)
	Update(domain.Task) error
	Delete(string) error
	GetLocker() *sync.RWMutex
}
