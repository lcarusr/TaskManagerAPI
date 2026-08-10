package data

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"task-manager-api/internal/biz"
)

func newTestData() *Data {
	return &Data{}
}

func TestTaskRepoCRUD(t *testing.T) {
	repo := NewTaskRepo(newTestData())
	ctx := context.Background()

	// Create
	task := &biz.Task{
		ID:        "uuid-1",
		Title:     "first",
		Status:    biz.StatusTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// FindByID
	got, err := repo.FindByID(ctx, "uuid-1")
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if got.Title != "first" {
		t.Errorf("title = %q, want first", got.Title)
	}

	// Find missing
	if _, err := repo.FindByID(ctx, "nope"); err == nil {
		t.Error("want ErrTaskNotFound")
	}

	// Update
	task.Title = "updated"
	task.Status = biz.StatusDone
	if _, err := repo.Update(ctx, task); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ = repo.FindByID(ctx, "uuid-1")
	if got.Title != "updated" || got.Status != biz.StatusDone {
		t.Errorf("after update: title=%q status=%d", got.Title, got.Status)
	}

	// Update missing
	_, err = repo.Update(ctx, &biz.Task{ID: "nope", Title: "x", Status: biz.StatusTodo})
	if err == nil {
		t.Error("want ErrTaskNotFound on update missing")
	}

	// List
	_, _ = repo.Create(ctx, &biz.Task{ID: "uuid-2", Title: "second", Status: biz.StatusInProgress})
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("want 2 tasks, got %d", len(list))
	}

	// Delete
	if err := repo.Delete(ctx, "uuid-1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := repo.Delete(ctx, "uuid-1"); err == nil {
		t.Error("want ErrTaskNotFound on double delete")
	}
}

func TestTaskRepoConcurrent(t *testing.T) {
	repo := NewTaskRepo(newTestData())
	ctx := context.Background()

	var wg sync.WaitGroup
	const n = 50
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("id-%d", i)
			_, _ = repo.Create(ctx, &biz.Task{ID: id, Title: fmt.Sprintf("task-%d", i), Status: biz.StatusTodo})
		}(i)
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = repo.List(ctx)
			_, _ = repo.FindByID(ctx, fmt.Sprintf("id-%d", i))
		}(i)
	}
	wg.Wait()

	list, _ := repo.List(ctx)
	if len(list) != n {
		t.Errorf("want %d tasks after concurrent writes, got %d", n, len(list))
	}
}
