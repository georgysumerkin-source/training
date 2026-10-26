package task

import (
	"errors"
	"fmt"
	"strings"

	"github.com/georgysumerkin-source/training/internal/common"
	"github.com/georgysumerkin-source/training/internal/internal_errors"
	"go.openly.dev/pointy"
)

type Status string

type Task struct {
	ID          string
	Title       string
	Description string
	Status      Status
}

func NewTask(
	ID string,
	title string,
	description string,
) (*Task, error) {
	task := &Task{
		ID:          strings.TrimSpace(ID),
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      StatusNew,
	}
	if err := task.Validate(); err != nil {
		return nil, err
	}

	return task, nil
}

func (t *Task) Validate() error {
	if err := common.ValidateString(t.ID, pointy.Int(IdMinLen), pointy.Int(IdMaxLen)); err != nil {
		return doValidateError(err, "ID")
	}

	if err := common.ValidateString(t.Title, pointy.Int(TitleMinLen), pointy.Int(TitleMaxLen)); err != nil {
		return doValidateError(err, "Title")
	}

	if err := common.ValidateString(t.Description, pointy.Int(DescriptionMinLen), pointy.Int(DescriptionMaxLen)); err != nil {
		return doValidateError(err, "Description")
	}

	if _, ok := statusMap[string(t.Status)]; !ok {
		return doValidateError(internal_errors.ErrInvalidTask, "Status")
	}

	return nil
}

func (t *Task) SetStatus(newStatus string) (*Task, error) {
	if err := t.validateStatus(Status(newStatus)); err != nil {
		return nil, doValidateError(err, string(t.Status), newStatus)
	}

	t.Status = Status(newStatus)

	return t, nil
}

func (t *Task) validateStatus(newStatus Status) error {
	if _, ok := statusMap[string(newStatus)]; !ok {
		return doValidateError(internal_errors.ErrInvalidTask, "Status")
	}

	switch newStatus {
	case StatusNew:
		return internal_errors.ErrInvalidTransition
	case StatusRunning:
		if t.Status == StatusStopped || t.Status == StatusRunning {
			return internal_errors.ErrInvalidTransition
		}
	case StatusStopped:
		if t.Status == StatusNew || t.Status == StatusStopped {
			return internal_errors.ErrInvalidTransition
		}
	}

	return nil
}

func doValidateError(err error, fields ...string) error {
	switch {
	case errors.Is(err, internal_errors.ErrInvalidTask):
		if len(fields) > 0 {
			return fmt.Errorf("invalid task params: %s: %w", fields[0], err)
		}

		return fmt.Errorf("invalid task params: %w", err)

	case errors.Is(err, internal_errors.ErrInvalidTransition):
		if len(fields) > 1 {
			return fmt.Errorf("invalid status transition: %s: %w", fields[1], err)
		}

		return fmt.Errorf("invalid task transition: %w", err)

	default:
		return err
	}
}
