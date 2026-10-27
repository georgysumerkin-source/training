package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/internal_errors"
	"github.com/georgysumerkin-source/training/internal/repository"
	"github.com/georgysumerkin-source/training/internal/service"
)

// Unconfigured methods panic so unexpected repository calls fail the test.
type repositoryStub struct {
	create func(context.Context, domain.Task) error
	get    func(context.Context, string) (*domain.Task, error)
	list   func(context.Context) ([]*domain.Task, error)
	update func(context.Context, domain.Task) error
	delete func(context.Context, string) error
}

func (r repositoryStub) Create(ctx context.Context, task domain.Task) error {
	return r.create(ctx, task)
}
func (r repositoryStub) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return r.get(ctx, id)
}
func (r repositoryStub) GetList(ctx context.Context) ([]*domain.Task, error) { return r.list(ctx) }
func (r repositoryStub) Update(ctx context.Context, task domain.Task) error {
	return r.update(ctx, task)
}
func (r repositoryStub) Delete(ctx context.Context, id string) error { return r.delete(ctx, id) }

func TestTaskServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	s := service.NewTaskService(repository.NewMemoryRepository())
	created, err := s.CreateTask(ctx, " id ", " title ", " description ")
	want := domain.Task{ID: "id", Title: "title", Description: "description", Status: domain.StatusNew}
	if err != nil || created == nil || *created != want {
		t.Fatalf("create: %+v, %v", created, err)
	}
	created.Title = "local mutation"
	got, err := s.GetTask(ctx, "id")
	if err != nil || got == nil || *got != want {
		t.Fatalf("get: %+v, %v", got, err)
	}
	if task, err := s.CreateTask(ctx, "id", "duplicate", ""); task != nil || !errors.Is(err, repository.ErrTaskAlreadyExists) {
		t.Fatalf("duplicate: %v, %v", task, err)
	}
	for _, status := range []string{domain.StatusRunning, domain.StatusStopped} {
		got, err = s.ChangeTaskStatus(ctx, "id", status)
		want.Status = domain.Status(status)
		if err != nil || got == nil || *got != want {
			t.Fatalf("change status: %+v, %v", got, err)
		}
		stored, err := s.GetTask(ctx, "id")
		if err != nil || stored == nil || *stored != want {
			t.Fatalf("persisted status: %+v, %v", stored, err)
		}
	}
	if task, err := s.ChangeTaskStatus(ctx, "id", domain.StatusRunning); task != nil || !errors.Is(err, internal_errors.ErrInvalidTransition) {
		t.Fatalf("invalid transition: %v, %v", task, err)
	}
	list, err := s.GetTaskList(ctx)
	if err != nil || len(list) != 1 || *list[0] != want {
		t.Fatalf("list: %v, %v", list, err)
	}
	if err := s.DeleteTask(ctx, "id"); err != nil {
		t.Fatal(err)
	}
	if task, err := s.GetTask(ctx, "id"); task != nil || !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("deleted task: %v, %v", task, err)
	}
	list, err = s.GetTaskList(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("empty list: %v, %v", list, err)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	s := service.NewTaskService(repositoryStub{})
	for _, tt := range []struct{ id, title string }{{"", "title"}, {"id", "  "}} {
		task, err := s.CreateTask(context.Background(), tt.id, tt.title, "")
		if task != nil || !errors.Is(err, internal_errors.ErrInvalidTask) {
			t.Fatalf("invalid task: %v, %v", task, err)
		}
	}
}

func TestTaskServiceRepositoryErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	failure := errors.New("repository failure")
	check := func(got context.Context, id string) {
		t.Helper()
		if got != ctx || id != "id" {
			t.Fatalf("repository args: %v, %q", got, id)
		}
	}
	for _, operation := range []string{"create", "get", "list", "delete", "status get", "status update"} {
		t.Run(operation, func(t *testing.T) {
			calls := 0
			fail := func(got context.Context, id string) error { check(got, id); calls++; return failure }
			r := repositoryStub{}
			var task *domain.Task
			var err error
			switch operation {
			case "create":
				r.create = func(ctx context.Context, task domain.Task) error {
					if task.Title != "title" || task.Description != "description" || task.Status != domain.StatusNew {
						t.Fatalf("created task: %+v", task)
					}
					return fail(ctx, task.ID)
				}
				task, err = service.NewTaskService(r).CreateTask(ctx, " id ", " title ", " description ")
			case "get", "status get":
				r.get = func(ctx context.Context, id string) (*domain.Task, error) { return nil, fail(ctx, id) }
				if operation == "get" {
					task, err = service.NewTaskService(r).GetTask(ctx, "id")
				} else {
					task, err = service.NewTaskService(r).ChangeTaskStatus(ctx, "id", domain.StatusRunning)
				}
			case "list":
				r.list = func(ctx context.Context) ([]*domain.Task, error) { return nil, fail(ctx, "id") }
				list, listErr := service.NewTaskService(r).GetTaskList(ctx)
				err = listErr
				if list != nil {
					t.Fatalf("list on error: %v", list)
				}
			case "delete":
				r.delete = fail
				err = service.NewTaskService(r).DeleteTask(ctx, "id")
			case "status update":
				r.get = func(ctx context.Context, id string) (*domain.Task, error) {
					check(ctx, id)
					return &domain.Task{ID: id, Title: "title", Status: domain.StatusNew}, nil
				}
				r.update = func(ctx context.Context, task domain.Task) error {
					if task.Status != domain.StatusRunning || task.Title != "title" {
						t.Fatalf("updated task: %+v", task)
					}
					return fail(ctx, task.ID)
				}
				task, err = service.NewTaskService(r).ChangeTaskStatus(ctx, "id", domain.StatusRunning)
			}
			if task != nil || !errors.Is(err, failure) || calls != 1 {
				t.Fatalf("result: %v, %v; repository calls %d", task, err, calls)
			}
		})
	}
}

func TestChangeTaskStatusRejectsWithoutUpdate(t *testing.T) {
	for _, status := range []string{domain.StatusNew, domain.StatusStopped, "unknown"} {
		t.Run(status, func(t *testing.T) {
			original := domain.Task{ID: "id", Title: "title", Status: domain.StatusNew}
			task := original
			r := repositoryStub{get: func(context.Context, string) (*domain.Task, error) { return &task, nil }}
			got, err := service.NewTaskService(r).ChangeTaskStatus(context.Background(), "id", status)
			want := internal_errors.ErrInvalidTransition
			if status == "unknown" {
				want = internal_errors.ErrInvalidTask
			}
			if got != nil || !errors.Is(err, want) || task != original {
				t.Fatalf("rejected change: %v, %v; task %+v", got, err, task)
			}
		})
	}
}
