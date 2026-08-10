package data

import (
	"context"
	"sync"

	"task-manager-api/internal/biz"
)

type taskRepo struct {
	data *Data

	mu    sync.RWMutex
	tasks map[string]*biz.Task
}

// NewTaskRepo creates a new in-memory TaskRepo.
func NewTaskRepo(data *Data) biz.TaskRepo {
	return &taskRepo{
		data:  data,
		tasks: make(map[string]*biz.Task),
	}
}

func (r *taskRepo) List(_ context.Context) ([]*biz.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*biz.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, cloneTask(t))
	}
	return tasks, nil
}

func (r *taskRepo) FindByID(_ context.Context, id string) (*biz.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tasks[id]
	if !ok {
		return nil, biz.ErrTaskNotFound
	}
	return cloneTask(t), nil
}

func (r *taskRepo) Create(_ context.Context, t *biz.Task) (*biz.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[t.ID] = cloneTask(t)
	return cloneTask(t), nil
}

func (r *taskRepo) Update(_ context.Context, t *biz.Task) (*biz.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[t.ID]; !ok {
		return nil, biz.ErrTaskNotFound
	}
	r.tasks[t.ID] = cloneTask(t)
	return cloneTask(t), nil
}

func (r *taskRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return biz.ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}

func cloneTask(t *biz.Task) *biz.Task {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}
