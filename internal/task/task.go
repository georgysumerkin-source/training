package task

import (
	"fmt"

	"github.com/georgysumerkin-source/training/internal/common"
	"go.openly.dev/pointy"
)

type status string

type Task struct {
	Id          string
	Title       string
	Description string
	Status      status
}

func NewTask(
	Id string,
	title string,
	description string,
	taskStatus status,
) (*Task, error) {
	task := &Task{
		Id:          Id,
		Title:       title,
		Description: description,
		Status:      taskStatus,
	}
	if err := task.Validate(); err != nil {
		return nil, err
	}

	return task, nil
}

func (t *Task) Validate() error {
	if err := common.ValidateString(t.Id, pointy.Int(IdMinLen), pointy.Int(IdMaxLen)); err != nil {
		return err
	}

	if err := common.ValidateString(t.Title, pointy.Int(TitleMinLen), pointy.Int(TitleMaxLen)); err != nil {
		return err
	}

	if err := common.ValidateString(t.Description, pointy.Int(IdMinLen), pointy.Int(IdMaxLen)); err != nil {
		return err
	}

	if _, ok := statusMap[string(t.Status)]; !ok {
		return fmt.Errorf("invalid status: %s", t.Status)
	}

	return nil
}

func (t *Task) SetStatus(newStatus string) (*Task, error) {
	if err := t.validateStatus(status(newStatus)); err != nil {
		return nil, err
	}

	t.Status = status(newStatus)

	return t, nil
}

func (t *Task) validateStatus(curStatus status) error {
	switch curStatus {
	case StatusRunning:
		if t.Status == StatusStopped {
			return fmt.Errorf("task %s is stopped", t.Id)
		}
	case StatusStopped:
		if t.Status == StatusNew {
			return fmt.Errorf("task %s is new", t.Id)
		}
	}

	return nil
}
