package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	v1 "task-manager-api/api/task/v1"
	"task-manager-api/internal/biz"
	"task-manager-api/internal/data"
	"task-manager-api/internal/service"

	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

func TestTokenBucketBurst(t *testing.T) {
	b := newTokenBucket(100, 3) // burst = 3
	for i := 0; i < 3; i++ {
		if _, err := b.Allow(); err != nil {
			t.Fatalf("allow #%d should pass within burst, got err: %v", i+1, err)
		}
	}
	if _, err := b.Allow(); err == nil {
		t.Fatal("4th allow should be rejected (burst exhausted)")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	b := newTokenBucket(1000, 1) // 每秒补 1000 个，桶容量 1
	if _, err := b.Allow(); err != nil {
		t.Fatal("first allow should pass")
	}
	if _, err := b.Allow(); err == nil {
		t.Fatal("second allow should be rejected immediately")
	}
	// 5ms 后按 1000/s 补充约 5 个令牌
	time.Sleep(5 * time.Millisecond)
	if _, err := b.Allow(); err != nil {
		t.Fatal("allow should pass after refill")
	}
}

func TestTokenBucketConcurrent(t *testing.T) {
	b := newTokenBucket(100000, 1000) // 大速率大桶，并发下不应 panic/race
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = b.Allow()
			}
		}()
	}
	wg.Wait()
}

// newRateLimitedHTTPServer 构造带指定限流器的 HTTP server（供中间件测试）。
func newRateLimitedHTTPServer(limiter ratelimit.Limiter) *khttp.Server {
	d := &data.Data{}
	repo := data.NewTaskRepo(d)
	uc := biz.NewTaskUsecase(repo)
	svc := service.NewTaskService(uc)

	var opts = []khttp.ServerOption{
		khttp.Middleware(
			ratelimit.Server(ratelimit.WithLimiter(limiter)),
			recovery.Recovery(),
		),
	}
	srv := khttp.NewServer(opts...)
	v1.RegisterTaskServiceHTTPServer(srv, svc)
	return srv
}

func TestRateLimitMiddleware(t *testing.T) {
	srv := newRateLimitedHTTPServer(newTokenBucket(5, 5)) // 5 QPS，burst 5

	var mu sync.Mutex
	counts := map[int]int{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			mu.Lock()
			counts[w.Code]++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if counts[http.StatusTooManyRequests] == 0 {
		t.Fatalf("expected some 429 responses, got %v", counts)
	}
	if counts[http.StatusOK] == 0 {
		t.Fatalf("expected some 200 responses, got %v", counts)
	}
}
