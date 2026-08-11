package server

import (
	"math"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
)

// tokenBucket 是一个固定速率令牌桶，实现 kratos ratelimit.Limiter 接口。
// - 按时间流逝每秒补充 rate 个令牌（上限 burst，允许突发）
// - 桶内无令牌时 Allow 返回 error → ratelimit 中间件返回 429
// 相比默认的 BBR（自适应限流），固定令牌桶的速率确定、压测可复现，
// 也更贴合作业要求的"令牌桶限流"。
type tokenBucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64 // 每秒补充令牌数（即允许的 QPS）
	burst  float64 // 桶容量（最大突发请求数）
}

// newTokenBucket 创建一个固定速率令牌桶，rps 为每秒请求数，burst 为突发容量。
func newTokenBucket(rps float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens: float64(burst),
		last:   time.Now(),
		rate:   rps,
		burst:  float64(burst),
	}
}

// Allow 尝试获取一个令牌；无令牌可用时返回错误（限流拒绝）。
func (b *tokenBucket) Allow() (ratelimit.DoneFunc, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	b.tokens = math.Min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
	b.last = now

	if b.tokens < 1 {
		return nil, ratelimit.ErrLimitExceed
	}
	b.tokens--
	return func(ratelimit.DoneInfo) {}, nil
}
