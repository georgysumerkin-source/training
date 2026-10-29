package repository_test

import (
	"errors"
	"testing"

	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/repository"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	r := repository.NewMemoryRepository()
	list, err := r.GetList()
	if err != nil || len(list) != 0 {
		t.Fatalf("empty list: %v, %v", list, err)
	}
	first := domain.Task{ID: "1", Title: "first", Status: domain.StatusNew}
	second := domain.Task{ID: "2", Title: "second", Status: domain.StatusNew}
	for _, task := range []domain.Task{first, second} {
		if err := r.Create(task); err != nil {
			t.Fatal(err)
		}
	}
	duplicate := first
	duplicate.Title = "duplicate"
	if err := r.Create(duplicate); !errors.Is(err, repository.ErrTaskAlreadyExists) {
		t.Fatal(err)
	}
	got, err := r.GetByID(first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("get: %+v, %v", got, err)
	}
	got.Title = "local mutation"
	list, err = r.GetList()
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v, %v", list, err)
	}
	seen := map[string]domain.Task{}
	for _, task := range list {
		seen[task.ID] = *task
		task.Title = "local list mutation"
	}
	if seen[first.ID] != first || seen[second.ID] != second {
		t.Fatalf("list contents: %+v", seen)
	}
	got, err = r.GetByID(first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("reads must return copies: %+v, %v", got, err)
	}
	first.Title = "updated"
	first.Status = domain.StatusRunning
	if err := r.Update(first); err != nil {
		t.Fatal(err)
	}
	got, err = r.GetByID(first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("update not persisted: %+v, %v", got, err)
	}
	if err := r.Delete(first.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := r.GetByID(first.ID); got != nil || !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("deleted task: %v, %v", got, err)
	}
	if err := r.Delete(first.ID); !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatal(err)
	}
	if err := r.Update(first); !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatal(err)
	}
	list, err = r.GetList()
	if err != nil || len(list) != 1 || *list[0] != second {
		t.Fatalf("remaining tasks: %v, %v", list, err)
	}
}

func TestMemoryRepositoryGetLocker(t *testing.T) {
	r := repository.NewMemoryRepository()
	locker := r.GetLocker()
	if locker == nil {
		t.Fatal("nil mutex")
	}
	if r.GetLocker() != locker {
		t.Fatal("GetLocker must return the same mutex")
	}
	other := repository.NewMemoryRepository()
	if other.GetLocker() == locker {
		t.Fatal("independent repositories share a mutex")
	}
	locker.Lock()
	defer locker.Unlock()
	if r.GetLocker().TryRLock() {
		r.GetLocker().RUnlock()
		t.Fatal("GetLocker returned an unlocked mutex")
	}
	if !other.GetLocker().TryLock() {
		t.Fatal("independent repository is blocked")
	}
	other.GetLocker().Unlock()
}
