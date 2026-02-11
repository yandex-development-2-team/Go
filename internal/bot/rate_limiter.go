package bot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/time/rate"
)

var botRateLimitHitsTotal uint64

type RateLimiter struct {
	apiLimiter  *rate.Limiter
	msgLimiters map[int64]*rate.Limiter
	mu          sync.Mutex

	memcache *memcache.Client

	now func() time.Time
}

func NewRateLimiter(memcacheAddr string) *RateLimiter {
	return &RateLimiter{
		apiLimiter:  rate.NewLimiter(rate.Limit(10), 10), // 10 req/sec
		msgLimiters: make(map[int64]*rate.Limiter),
		memcache:    memcache.New(memcacheAddr),
		now:         time.Now,
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

func (rl *RateLimiter) WaitIfNeeded(ctx context.Context, chatID *int64) error {
	if err := rl.apiLimiter.Wait(ctx); err != nil {
		return err
	}

	if chatID != nil {
		l := rl.getMsgLimiter(*chatID)
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}

	return rl.incrementWithLimit(ctx, chatID)
}

func (rl *RateLimiter) incrementWithLimit(ctx context.Context, chatID *int64) error {
	const limit = uint64(30) // messages per chat
	key := fmt.Sprintf("rate:%d:%d", 0, rl.now().Unix())
	if chatID != nil {
		key = fmt.Sprintf("rate:%d:%d", *chatID, rl.now().Unix())
	}

	for {
		item := &memcache.Item{
			Key:        key,
			Value:      []byte("1"),
			Expiration: 1,
		}

		err := rl.memcache.Add(item)
		if err == memcache.ErrNotStored {
			newVal, err := rl.memcache.Increment(key, 1)
			if err != nil {
				return err
			}

			if newVal > limit {
				atomic.AddUint64(&botRateLimitHitsTotal, 1)

				now := rl.now()
				nextWindow := now.Truncate(time.Second).Add(time.Second)
				sleep := time.Until(nextWindow)
				log.Printf("[WARN] rate limit delay=%v chatID=%v", sleep, chatID)

				timer := time.NewTimer(sleep)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}

				continue
			}
		} else if err != nil {
			return err
		}

		break
	}

	return nil
}
