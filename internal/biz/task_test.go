package biz

import (
	"context"
	"sync"
	"testing"
)

// memoryRepo 是测试用的内存实现，避免依赖 data 包造成循环引用。
type memoryRepo struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{tasks: make(map[string]*Task)}
}

func (r *memoryRepo) List(_ context.Context) ([]*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t)
	}
	return out, nil
}

func (r *memoryRepo) FindByID(_ context.Context, id string) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return t, nil
}

func (r *memoryRepo) Create(_ context.Context, t *Task) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[t.ID] = t
	return t, nil
}

func (r *memoryRepo) Update(_ context.Context, t *Task) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[t.ID]; !ok {
		return nil, ErrTaskNotFound
	}
	r.tasks[t.ID] = t
	return t, nil
}

func (r *memoryRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}

func newTestUsecase() *TaskUsecase {
	return NewTaskUsecase(newMemoryRepo())
}

func TestCreateTask(t *testing.T) {
	uc := newTestUsecase()

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{"valid", &Task{Title: "write report", Status: StatusTodo}, false},
		{"empty title", &Task{Title: "  ", Status: StatusTodo}, true},
		{"nil task", nil, true},
		{"invalid status", &Task{Title: "t", Status: Status(99)}, true},
		{"zero status", &Task{Title: "t", Status: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.Create(context.Background(), tt.task)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID == "" {
				t.Error("id should be assigned")
			}
			if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
				t.Error("timestamps should be assigned")
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	uc := newTestUsecase()
	created, _ := uc.Create(context.Background(), &Task{Title: "t1", Status: StatusTodo})

	t.Run("found", func(t *testing.T) {
		got, err := uc.Get(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "t1" {
			t.Errorf("title = %q, want %q", got.Title, "t1")
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := uc.Get(context.Background(), "no-such-id"); err == nil {
			t.Error("want ErrTaskNotFound, got nil")
		}
	})

	t.Run("empty id", func(t *testing.T) {
		if _, err := uc.Get(context.Background(), "  "); err == nil {
			t.Error("want error for empty id, got nil")
		}
	})
}

func TestListTasks(t *testing.T) {
	uc := newTestUsecase()

	t.Run("empty", func(t *testing.T) {
		tasks, err := uc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("want 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("multiple", func(t *testing.T) {
		_, _ = uc.Create(context.Background(), &Task{Title: "a", Status: StatusTodo})
		_, _ = uc.Create(context.Background(), &Task{Title: "b", Status: StatusInProgress})
		tasks, err := uc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 {
			t.Errorf("want 2 tasks, got %d", len(tasks))
		}
	})
}

func TestUpdateTask(t *testing.T) {
	uc := newTestUsecase()
	created, _ := uc.Create(context.Background(), &Task{Title: "old", Status: StatusTodo})

	t.Run("update ok", func(t *testing.T) {
		got, err := uc.Update(context.Background(), created.ID, &Task{
			Title: "new title",
			Status: StatusDone,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "new title" || got.Status != StatusDone {
			t.Errorf("got title=%q status=%d, want %q/%d", got.Title, got.Status, "new title", StatusDone)
		}
		if got.UpdatedAt.Before(created.UpdatedAt) {
			t.Error("updated_at should be refreshed")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := uc.Update(context.Background(), "missing", &Task{Title: "x", Status: StatusTodo})
		if err == nil {
			t.Error("want ErrTaskNotFound, got nil")
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := uc.Update(context.Background(), "", &Task{Title: "x", Status: StatusTodo})
		if err == nil {
			t.Error("want error for empty id, got nil")
		}
	})

	t.Run("invalid title", func(t *testing.T) {
		_, err := uc.Update(context.Background(), created.ID, &Task{Title: "", Status: StatusTodo})
		if err == nil {
			t.Error("want error for empty title, got nil")
		}
	})
}

func TestDeleteTask(t *testing.T) {
	uc := newTestUsecase()
	created, _ := uc.Create(context.Background(), &Task{Title: "to delete", Status: StatusTodo})

	t.Run("delete ok", func(t *testing.T) {
		if err := uc.Delete(context.Background(), created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := uc.Get(context.Background(), created.ID); err == nil {
			t.Error("task should be gone")
		}
	})

	t.Run("delete missing", func(t *testing.T) {
		if err := uc.Delete(context.Background(), "missing"); err == nil {
			t.Error("want ErrTaskNotFound, got nil")
		}
	})
}
