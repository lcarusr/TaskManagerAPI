package server

import (
	"net/http"

	v1 "task-manager-api/api/task/v1"
	"task-manager-api/internal/conf"
	"task-manager-api/internal/service"

	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, task *service.TaskService) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			// 令牌桶限流：固定 100 QPS（突发 100），所有端点统一拦截，超限返回 429
			ratelimit.Server(ratelimit.WithLimiter(newTokenBucket(100, 100))),
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, khttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := khttp.NewServer(opts...)
	// 健康检查端点（GET /health → 200）
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	v1.RegisterTaskServiceHTTPServer(srv, task)
	return srv
}
