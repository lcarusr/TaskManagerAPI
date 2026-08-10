package service

import (
	"context"
	"net/http"

	v1 "task-manager-api/api/task/v1"
	"task-manager-api/internal/biz"

	"github.com/go-kratos/kratos/v3/transport"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskService is the task service.
type TaskService struct {
	v1.UnimplementedTaskServiceServer

	uc *biz.TaskUsecase
}

// NewTaskService new a task service.
func NewTaskService(uc *biz.TaskUsecase) *TaskService {
	return &TaskService{uc: uc}
}

// ListTasks returns all tasks.
func (s *TaskService) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksReply, error) {
	tasks, err := s.uc.List(ctx)
	if err != nil {
		return nil, err
	}
	reply := &v1.ListTasksReply{Tasks: make([]*v1.Task, 0, len(tasks))}
	for _, t := range tasks {
		reply.Tasks = append(reply.Tasks, toProto(t))
	}
	return reply, nil
}

// GetTask returns a single task by id.
func (s *TaskService) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.Task, error) {
	t, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toProto(t), nil
}

// CreateTask creates a new task.
func (s *TaskService) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.Task, error) {
	t, err := s.uc.Create(ctx, &biz.Task{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Status:      fromProtoStatus(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}
	// 作业要求：POST /tasks → 201
	setHTTPStatus(ctx, http.StatusCreated)
	return toProto(t), nil
}

// UpdateTask updates an existing task.
// 采用部分更新语义：仅覆盖请求中出现的字段。
func (s *TaskService) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.Task, error) {
	current, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if req.GetTitle() != "" {
		current.Title = req.GetTitle()
	}
	if req.GetDescription() != "" {
		current.Description = req.GetDescription()
	}
	if req.GetStatus() != v1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		current.Status = fromProtoStatus(req.GetStatus())
	}
	t, err := s.uc.Update(ctx, req.GetId(), current)
	if err != nil {
		return nil, err
	}
	return toProto(t), nil
}

// DeleteTask deletes a task by id.
func (s *TaskService) DeleteTask(ctx context.Context, req *v1.DeleteTaskRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	// 作业要求：DELETE /tasks/{id} → 204
	setHTTPStatus(ctx, http.StatusNoContent)
	return &emptypb.Empty{}, nil
}

// setHTTPStatus 在 service 层设置自定义 HTTP 状态码（201/204 等）。
// kratos 的 DefaultResponseEncoder 不显式写状态码，响应写入时使用已设置的值。
func setHTTPStatus(ctx context.Context, code int) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if ht, ok := tr.(khttp.ResponseTransporter); ok {
			ht.Response().WriteHeader(code)
		}
	}
}

func fromProtoStatus(s v1.TaskStatus) biz.Status {
	switch s {
	case v1.TaskStatus_todo:
		return biz.StatusTodo
	case v1.TaskStatus_in_progress:
		return biz.StatusInProgress
	case v1.TaskStatus_done:
		return biz.StatusDone
	default:
		// TASK_STATUS_UNSPECIFIED → 默认 todo
		return biz.StatusTodo
	}
}

func toProto(t *biz.Task) *v1.Task {
	return &v1.Task{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      toProtoStatus(t.Status),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func toProtoStatus(s biz.Status) v1.TaskStatus {
	switch s {
	case biz.StatusTodo:
		return v1.TaskStatus_todo
	case biz.StatusInProgress:
		return v1.TaskStatus_in_progress
	case biz.StatusDone:
		return v1.TaskStatus_done
	default:
		return v1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}
