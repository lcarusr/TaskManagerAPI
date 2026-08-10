package biz

import (
	"context"
	"strings"
	"time"

	v1 "task-manager-api/api/task/v1"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/google/uuid"
)

var (
	// ErrTaskNotFound is returned when a task does not exist.
	ErrTaskNotFound = errors.NotFound(v1.ErrorReason_TASK_NOT_FOUND.String(), "task not found")
	// ErrTaskInvalidArgument is returned when a task request is invalid.
	ErrTaskInvalidArgument = errors.BadRequest(v1.ErrorReason_TASK_INVALID_ARGUMENT.String(), "invalid task argument")
)

// Status is the status of a task.
type Status int

const (
	// StatusTodo is the initial status of a task.
	StatusTodo Status = iota + 1
	// StatusInProgress is the status of an in-progress task.
	StatusInProgress
	// StatusDone is the status of a completed task.
	StatusDone
)

// Task is the task domain model.
type Task struct {
	ID          string
	Title       string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskRepo is the task repository interface.
// The implementation is pluggable: in-memory, PostgreSQL, Redis, etc.
type TaskRepo interface {
	List(context.Context) ([]*Task, error)
	FindByID(context.Context, string) (*Task, error)
	Create(context.Context, *Task) (*Task, error)
	Update(context.Context, *Task) (*Task, error)
	Delete(context.Context, string) error
}

// TaskUsecase is the task usecase.
type TaskUsecase struct {
	repo TaskRepo
}

// NewTaskUsecase new a task usecase.
func NewTaskUsecase(repo TaskRepo) *TaskUsecase {
	return &TaskUsecase{repo: repo}
}

// List returns all tasks.
func (uc *TaskUsecase) List(ctx context.Context) ([]*Task, error) {
	return uc.repo.List(ctx)
}

// Get returns a task by ID.
func (uc *TaskUsecase) Get(ctx context.Context, id string) (*Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTaskInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

// Create creates a new task with a server-assigned UUID and timestamps.
func (uc *TaskUsecase) Create(ctx context.Context, t *Task) (*Task, error) {
	if err := validateTask(t); err != nil {
		return nil, err
	}
	now := time.Now()
	t.ID = uuid.NewString()
	t.CreatedAt = now
	t.UpdatedAt = now
	return uc.repo.Create(ctx, t)
}

// Update updates an existing task.
func (uc *TaskUsecase) Update(ctx context.Context, id string, t *Task) (*Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTaskInvalidArgument
	}
	if err := validateTask(t); err != nil {
		return nil, err
	}
	t.ID = id
	t.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, t)
}

// Delete deletes a task by ID.
func (uc *TaskUsecase) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrTaskInvalidArgument
	}
	return uc.repo.Delete(ctx, id)
}

func validateTask(t *Task) error {
	if t == nil || strings.TrimSpace(t.Title) == "" {
		return ErrTaskInvalidArgument
	}
	if t.Status < StatusTodo || t.Status > StatusDone {
		return ErrTaskInvalidArgument
	}
	return nil
}
