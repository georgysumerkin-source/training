package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/georgysumerkin-source/training/internal/domain"
	"github.com/georgysumerkin-source/training/internal/internal_errors"
)

func TestNewTask(t *testing.T) {
	got, err := domain.NewTask(" \tid\n", " Заголовок ", " описание\n")
	want := domain.Task{ID: "id", Title: "Заголовок", Description: "описание", Status: domain.StatusNew}
	if err != nil || got == nil || *got != want {
		t.Fatalf("NewTask = %+v, %v; want %+v", got, err, want)
	}
}

func TestTaskValidationBoundaries(t *testing.T) {
	for _, field := range []struct {
		name     string
		min, max int
	}{
		{"ID", 1, 64}, {"Title", 3, 100}, {"Description", 0, 500},
	} {
		for _, length := range []int{0, field.min, field.min - 1, field.max, field.max + 1} {
			if length < 0 {
				continue
			}
			t.Run(fmt.Sprintf("%s/length=%d", field.name, length), func(t *testing.T) {
				task := domain.Task{ID: "id", Title: "title", Status: domain.StatusNew}
				value := strings.Repeat("я", length)
				switch field.name {
				case "ID":
					task.ID = value
				case "Title":
					task.Title = value
				case "Description":
					task.Description = value
				}
				invalid := length < field.min || length > field.max
				err := task.Validate()
				created, createErr := domain.NewTask(task.ID, task.Title, task.Description)
				if invalid {
					for _, e := range []error{err, createErr} {
						if !errors.Is(e, internal_errors.ErrInvalidTask) || !strings.Contains(e.Error(), field.name) {
							t.Fatalf("error = %v, want invalid %s", e, field.name)
						}
					}
					if created != nil {
						t.Fatal("invalid task returned")
					}
				} else if err != nil || createErr != nil || created == nil {
					t.Fatalf("valid task rejected: %v, %v", err, createErr)
				}
			})
		}
	}
	for _, status := range []domain.Status{domain.StatusNew, domain.StatusRunning, domain.StatusStopped, "unknown", ""} {
		task := domain.Task{ID: "id", Title: "title", Status: status}
		err := task.Validate()
		valid := status == domain.StatusNew || status == domain.StatusRunning || status == domain.StatusStopped
		if valid && err != nil {
			t.Fatal(err)
		}
		if !valid && !errors.Is(err, internal_errors.ErrInvalidTask) {
			t.Fatalf("status %q: %v", status, err)
		}
	}
}

func TestSetStatus(t *testing.T) {
	for _, from := range []domain.Status{domain.StatusNew, domain.StatusRunning, domain.StatusStopped} {
		for _, to := range []string{domain.StatusNew, domain.StatusRunning, domain.StatusStopped, "unknown", ""} {
			t.Run(string(from)+"/"+to, func(t *testing.T) {
				task := domain.Task{ID: "id", Title: "title", Description: "description", Status: from}
				before := task
				err := task.SetStatus(to)
				allowed := from == domain.StatusNew && to == domain.StatusRunning || from == domain.StatusRunning && to == domain.StatusStopped
				if allowed {
					before.Status = domain.Status(to)
					if err != nil || task != before {
						t.Fatalf("transition: %+v, %v", task, err)
					}
				} else {
					want := internal_errors.ErrInvalidTransition
					if to == "unknown" || to == "" {
						want = internal_errors.ErrInvalidTask
					}
					if !errors.Is(err, want) || task != before {
						t.Fatalf("rejected transition mutated task or wrong error: %+v, %v", task, err)
					}
				}
			})
		}
	}
}
