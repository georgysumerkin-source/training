package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/internal_errors"
	"github.com/georgysumerkin-source/training/internal/repository"
	"github.com/georgysumerkin-source/training/internal/service"
)

// Unconfigured methods panic so unexpected repository calls fail the test.
type repositoryStub struct {
	locker sync.RWMutex
	create func(domain.Task) error
	get    func(string) (*domain.Task, error)
	list   func() ([]*domain.Task, error)
	update func(domain.Task) error
	delete func(string) error
}

func (r *repositoryStub) Create(task domain.Task) error {
	return r.create(task)
}
func (r *repositoryStub) GetByID(id string) (*domain.Task, error) {
	return r.get(id)
}
func (r *repositoryStub) GetList() ([]*domain.Task, error) { return r.list() }
func (r *repositoryStub) Update(task domain.Task) error {
	return r.update(task)
}
func (r *repositoryStub) Delete(id string) error { return r.delete(id) }

func (r *repositoryStub) GetLocker() *sync.RWMutex { return &r.locker }

var _ repository.TaskRepository = (*repositoryStub)(nil)

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
	s := service.NewTaskService(&repositoryStub{})
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
	check := func(id string) {
		t.Helper()
		if id != "id" {
			t.Fatalf("repository id: %q", id)
		}
	}
	for _, operation := range []string{"create", "get", "list", "delete", "status get", "status update"} {
		t.Run(operation, func(t *testing.T) {
			calls := 0
			fail := func(id string) error { check(id); calls++; return failure }
			r := repositoryStub{}
			var task *domain.Task
			var err error
			switch operation {
			case "create":
				r.create = func(task domain.Task) error {
					if task.Title != "title" || task.Description != "description" || task.Status != domain.StatusNew {
						t.Fatalf("created task: %+v", task)
					}
					return fail(task.ID)
				}
				task, err = service.NewTaskService(&r).CreateTask(ctx, " id ", " title ", " description ")
			case "get", "status get":
				r.get = func(id string) (*domain.Task, error) { return nil, fail(id) }
				if operation == "get" {
					task, err = service.NewTaskService(&r).GetTask(ctx, "id")
				} else {
					task, err = service.NewTaskService(&r).ChangeTaskStatus(ctx, "id", domain.StatusRunning)
				}
			case "list":
				r.list = func() ([]*domain.Task, error) { return nil, fail("id") }
				list, listErr := service.NewTaskService(&r).GetTaskList(ctx)
				err = listErr
				if list != nil {
					t.Fatalf("list on error: %v", list)
				}
			case "delete":
				r.delete = fail
				err = service.NewTaskService(&r).DeleteTask(ctx, "id")
			case "status update":
				r.get = func(id string) (*domain.Task, error) {
					check(id)
					return &domain.Task{ID: id, Title: "title", Status: domain.StatusNew}, nil
				}
				r.update = func(task domain.Task) error {
					if task.Status != domain.StatusRunning || task.Title != "title" {
						t.Fatalf("updated task: %+v", task)
					}
					return fail(task.ID)
				}
				task, err = service.NewTaskService(&r).ChangeTaskStatus(ctx, "id", domain.StatusRunning)
			}
			if !r.locker.TryLock() {
				t.Fatal("repository lock not released after error")
			}
			r.locker.Unlock()
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
			r := repositoryStub{get: func(string) (*domain.Task, error) { return &task, nil }}
			got, err := service.NewTaskService(&r).ChangeTaskStatus(context.Background(), "id", status)
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

func TestTaskServiceContext(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(fmt.Sprintf("expired=%t", expired), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			want := context.Canceled
			if expired {
				cancel()
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				want = context.DeadlineExceeded
			} else {
				cancel()
			}
			defer cancel()
			// GetLocker is allowed, but no data operation may run with a canceled context.
			s := service.NewTaskService(&repositoryStub{})
			for _, operation := range []struct {
				name string
				run  func() error
			}{
				{"create", func() error {
					task, err := s.CreateTask(ctx, "id", "title", "")
					if task != nil {
						t.Error("create returned task")
					}
					return err
				}},
				{"get", func() error {
					task, err := s.GetTask(ctx, "id")
					if task != nil {
						t.Error("get returned task")
					}
					return err
				}},
				{"list", func() error {
					list, err := s.GetTaskList(ctx)
					if list != nil {
						t.Error("list returned tasks")
					}
					return err
				}},
				{"delete", func() error { return s.DeleteTask(ctx, "id") }},
				{"status", func() error {
					task, err := s.ChangeTaskStatus(ctx, "id", domain.StatusRunning)
					if task != nil {
						t.Error("status returned task")
					}
					return err
				}},
			} {
				if err := operation.run(); !errors.Is(err, want) {
					t.Errorf("%s: error = %v, want %v", operation.name, err, want)
				}
			}
		})
	}
}

func TestTaskServiceConcurrentAccess(t *testing.T) {
	r := repository.NewMemoryRepository()
	s := service.NewTaskService(r)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Go(func() {
			s := service.NewTaskService(r)
			id := fmt.Sprint(i)
			if _, err := s.CreateTask(ctx, id, "title", ""); err != nil {
				t.Error(err)
				return
			}
			changed, err := s.ChangeTaskStatus(ctx, id, domain.StatusRunning)
			if err != nil {
				t.Error(err)
				return
			}
			got, err := s.GetTask(ctx, id)
			if err != nil || got == nil || changed == nil || *got != *changed {
				t.Errorf("get: %+v, %v; want %+v", got, err, changed)
			}
			if _, err := s.GetTaskList(ctx); err != nil {
				t.Error(err)
			}
			if err := s.DeleteTask(ctx, id); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	list, err := s.GetTaskList(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("remaining tasks: %v, %v", list, err)
	}
}

func TestConcurrentStatusChange(t *testing.T) {
	r := repository.NewMemoryRepository()
	s := service.NewTaskService(r)
	ctx := context.Background()
	if _, err := s.CreateTask(ctx, "id", "title", ""); err != nil {
		t.Fatal(err)
	}
	const workers = 32
	start := make(chan struct{})
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Go(func() {
			s := service.NewTaskService(r)
			<-start
			_, err := s.ChangeTaskStatus(ctx, "id", domain.StatusRunning)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, internal_errors.ErrInvalidTransition) {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful transitions = %d, want 1", successes)
	}
	task, err := s.GetTask(ctx, "id")
	if err != nil || task == nil || task.Status != domain.StatusRunning {
		t.Fatalf("final task: %+v, %v", task, err)
	}
}

func TestTaskServiceUsesRepositoryLock(t *testing.T) {
	for _, operation := range []string{"create", "get", "list", "delete", "status"} {
		t.Run(operation, func(t *testing.T) {
			r := &repositoryStub{}
			calls := 0
			checkLock := func() {
				t.Helper()
				calls++
				if r.locker.TryLock() {
					r.locker.Unlock()
					t.Fatal("repository mutex is not held")
				}
				readLocked := r.locker.TryRLock()
				if readLocked {
					r.locker.RUnlock()
				}
				wantRead := operation == "get" || operation == "list"
				if readLocked != wantRead {
					t.Fatalf("read lock = %t, want %t", readLocked, wantRead)
				}
			}
			r.create = func(domain.Task) error { checkLock(); return nil }
			r.get = func(id string) (*domain.Task, error) {
				checkLock()
				return &domain.Task{ID: id, Title: "title", Status: domain.StatusNew}, nil
			}
			r.list = func() ([]*domain.Task, error) { checkLock(); return nil, nil }
			r.delete = func(string) error { checkLock(); return nil }
			r.update = func(domain.Task) error { checkLock(); return nil }
			s := service.NewTaskService(r)
			ctx := context.Background()
			var err error
			switch operation {
			case "create":
				_, err = s.CreateTask(ctx, "id", "title", "")
			case "get":
				_, err = s.GetTask(ctx, "id")
			case "list":
				_, err = s.GetTaskList(ctx)
			case "delete":
				err = s.DeleteTask(ctx, "id")
			case "status":
				_, err = s.ChangeTaskStatus(ctx, "id", domain.StatusRunning)
			}
			if err != nil {
				t.Fatal(err)
			}
			wantCalls := 1
			if operation == "status" {
				wantCalls = 2
			}
			if calls != wantCalls {
				t.Fatalf("calls = %d, want %d", calls, wantCalls)
			}
			if !r.locker.TryLock() {
				t.Fatal("repository mutex not released")
			}
			r.locker.Unlock()
		})
	}
}
