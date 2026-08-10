package service

import (
	"context"
	"testing"

	v1 "task-manager-api/api/task/v1"
	"task-manager-api/internal/biz"
	"task-manager-api/internal/data"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestService() *TaskService {
	d := &data.Data{}
	repo := data.NewTaskRepo(d)
	uc := biz.NewTaskUsecase(repo)
	return NewTaskService(uc)
}

func TestTaskServiceCRUD(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	// Create
	created, err := s.CreateTask(ctx, &v1.CreateTaskRequest{Title: "svc task", Description: "desc"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.GetId() == "" {
		t.Error("id should be assigned")
	}
	if created.GetStatus() != v1.TaskStatus_todo {
		t.Errorf("default status = %v, want todo", created.GetStatus())
	}

	// Get
	got, err := s.GetTask(ctx, &v1.GetTaskRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.GetTitle() != "svc task" {
		t.Errorf("title = %q", got.GetTitle())
	}

	// Get missing → gRPC NOT_FOUND
	_, err = s.GetTask(ctx, &v1.GetTaskRequest{Id: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("want NotFound, got %v", err)
	}

	// Update (partial)
	updated, err := s.UpdateTask(ctx, &v1.UpdateTaskRequest{
		Id:    created.GetId(),
		Title: "renamed",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.GetTitle() != "renamed" {
		t.Errorf("title = %q, want renamed", updated.GetTitle())
	}
	if updated.GetDescription() != "desc" {
		t.Errorf("description should be preserved, got %q", updated.GetDescription())
	}

	// List
	list, err := s.ListTasks(ctx, &v1.ListTasksRequest{})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.GetTasks()) != 1 {
		t.Errorf("want 1 task, got %d", len(list.GetTasks()))
	}

	// Delete
	if _, err := s.DeleteTask(ctx, &v1.DeleteTaskRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := s.GetTask(ctx, &v1.GetTaskRequest{Id: created.GetId()}); err == nil {
		t.Error("task should be gone")
	}

	// Delete missing
	_, err = s.DeleteTask(ctx, &v1.DeleteTaskRequest{Id: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("want NotFound on delete missing, got %v", err)
	}

	// Create invalid (empty title) → INVALID_ARGUMENT
	_, err = s.CreateTask(ctx, &v1.CreateTaskRequest{Title: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("want InvalidArgument, got %v", err)
	}
}
