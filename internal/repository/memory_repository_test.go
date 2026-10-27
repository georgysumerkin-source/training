package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/repository"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	r := repository.NewMemoryRepository()
	list, err := r.GetList(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("empty list: %v, %v", list, err)
	}
	first := domain.Task{ID: "1", Title: "first", Status: domain.StatusNew}
	second := domain.Task{ID: "2", Title: "second", Status: domain.StatusNew}
	for _, task := range []domain.Task{first, second} {
		if err := r.Create(ctx, task); err != nil {
			t.Fatal(err)
		}
	}
	duplicate := first
	duplicate.Title = "duplicate"
	if err := r.Create(ctx, duplicate); !errors.Is(err, repository.ErrTaskAlreadyExists) {
		t.Fatal(err)
	}
	got, err := r.GetByID(ctx, first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("get: %+v, %v", got, err)
	}
	got.Title = "local mutation"
	list, err = r.GetList(ctx)
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
	got, err = r.GetByID(ctx, first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("reads must return copies: %+v, %v", got, err)
	}
	first.Title = "updated"
	first.Status = domain.StatusRunning
	if err := r.Update(ctx, first); err != nil {
		t.Fatal(err)
	}
	got, err = r.GetByID(ctx, first.ID)
	if err != nil || got == nil || *got != first {
		t.Fatalf("update not persisted: %+v, %v", got, err)
	}
	if err := r.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := r.GetByID(ctx, first.ID); got != nil || !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("deleted task: %v, %v", got, err)
	}
	if err := r.Delete(ctx, first.ID); !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatal(err)
	}
	if err := r.Update(ctx, first); !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatal(err)
	}
	list, err = r.GetList(ctx)
	if err != nil || len(list) != 1 || *list[0] != second {
		t.Fatalf("remaining tasks: %v, %v", list, err)
	}
}

func TestMemoryRepositoryContext(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(fmt.Sprint("expired=", expired), func(t *testing.T) {
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
			r := repository.NewMemoryRepository()
			original := domain.Task{ID: "existing", Title: "original"}
			if err := r.Create(context.Background(), original); err != nil {
				t.Fatal(err)
			}
			changed := original
			changed.Title = "changed"
			for _, err := range []error{r.Create(ctx, domain.Task{ID: "new"}), r.Update(ctx, changed), r.Delete(ctx, original.ID)} {
				if !errors.Is(err, want) {
					t.Fatalf("error = %v, want %v", err, want)
				}
			}
			if task, err := r.GetByID(ctx, original.ID); task != nil || !errors.Is(err, want) {
				t.Fatalf("get: %v, %v", task, err)
			}
			if list, err := r.GetList(ctx); list != nil || !errors.Is(err, want) {
				t.Fatalf("list: %v, %v", list, err)
			}
			list, err := r.GetList(context.Background())
			if err != nil || len(list) != 1 || *list[0] != original {
				t.Fatalf("canceled operations changed storage: %v, %v", list, err)
			}
		})
	}
}

func TestMemoryRepositoryConcurrentAccess(t *testing.T) {
	r := repository.NewMemoryRepository()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Go(func() {
			task := domain.Task{ID: fmt.Sprint(i), Title: "title", Status: domain.StatusNew}
			if err := r.Create(ctx, task); err != nil {
				t.Error(err)
				return
			}
			task.Status = domain.StatusRunning
			if err := r.Update(ctx, task); err != nil {
				t.Error(err)
			}
			if got, err := r.GetByID(ctx, task.ID); err != nil || got == nil || *got != task {
				t.Errorf("get: %v, %v", got, err)
			}
			if _, err := r.GetList(ctx); err != nil {
				t.Error(err)
			}
			if err := r.Delete(ctx, task.ID); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	list, err := r.GetList(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("remaining tasks: %v, %v", list, err)
	}
}
