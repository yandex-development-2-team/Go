package bot

import (
	"context"
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var botRateLimitsHitsTotal uint64

type RateLimiter struct {
	apiLimiter  *rate.Limiter
	msgLimiters map[int64]*rate.Limiter
	mu          sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		apiLimiter:  rate.NewLimiter(rate.Limit(10), 10),
		msgLimiters: make(map[int64]*rate.Limiter),
	}
}

func (r *RateLimiter) getMsgLimiter(chatID int64) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, ok := r.msgLimiters[chatID]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(30), 30)
		r.msgLimiters[chatID] = limiter
	}
	return limiter
}

func (r *RateLimiter) WaitIfNeeded(ctx context.Context, chatID *int64) error {
	start := time.Now()
	if err := r.apiLimiter.Wait(ctx); err != nil {
		return err
	}
	if chatID != nil {
		limiter := r.getMsgLimiter(*chatID)
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
	}
	elapsed := time.Since(start)
	if elapsed > 0 {
		botRateLimitsHitsTotal++
		log.Printf(
			"[WARN] telegram rate limit hit, delayed for %s (chatID=%v)",
			elapsed,
			chatID,
		)
	}
	return nil
}
