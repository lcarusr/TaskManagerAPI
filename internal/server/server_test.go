package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-manager-api/internal/biz"
	"task-manager-api/internal/conf"
	"task-manager-api/internal/data"
	"task-manager-api/internal/service"

	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

func newTestHTTPServer() *khttp.Server {
	cfg := &conf.Server{
		Http: &conf.Server_HTTP{Addr: "0.0.0.0:8080"},
		Grpc: &conf.Server_GRPC{Addr: "0.0.0.0:9000"},
	}
	d := &data.Data{}
	repo := data.NewTaskRepo(d)
	uc := biz.NewTaskUsecase(repo)
	svc := service.NewTaskService(uc)
	return NewHTTPServer(cfg, svc)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestHTTPServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /health = %d, want 200", w.Code)
	}
}

func TestTaskRoutes(t *testing.T) {
	srv := newTestHTTPServer()

	// POST /tasks → 201
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"route test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /tasks = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	// GET /tasks → 200
	req = httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /tasks = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "route test") {
		t.Errorf("GET /tasks body missing task: %s", w.Body.String())
	}

	// GET /tasks/{id} → 404 for missing
	req = httptest.NewRequest(http.MethodGet, "/tasks/no-such-id", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GET /tasks/missing = %d, want 404", w.Code)
	}
}
