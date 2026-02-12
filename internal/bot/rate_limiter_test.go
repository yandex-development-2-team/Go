package bot

import (
	"context"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/time/rate"
)

func flushMemcache(t *testing.T) {
	mc := memcache.New("127.0.0.1:11211")
	if err := mc.FlushAll(); err != nil {
		t.Fatalf("failed to flush memcache: %v", err)
	}
}

func uniqueChatID() int64 {
	return time.Now().UnixNano() // уникальный ID для теста
}

func TestRateLimiter_GlobalLimit_Distributed(t *testing.T) {
	flushMemcache(t)

	fixed := time.Now()
	rl := &RateLimiter{
		apiLimiter:  rate.NewLimiter(rate.Inf, 0),
		msgLimiters: make(map[int64]*rate.Limiter),
		memcache:    memcache.New("127.0.0.1:11211"),
		now:         func() time.Time { return fixed },
	}

	chatID := uniqueChatID()

	// полностью забиваем лимит
	for i := 0; i < 30; i++ {
		if err := rl.WaitIfNeeded(context.Background(), &chatID); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.WaitIfNeeded(ctx, &chatID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected context deadline exceeded")
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("did not actually wait")
	}
}

func TestRateLimiter_ChatLimit_Distributed(t *testing.T) {
	flushMemcache(t)

	fixed := time.Now()
	rl := &RateLimiter{
		apiLimiter:  rate.NewLimiter(rate.Inf, 0),
		msgLimiters: make(map[int64]*rate.Limiter),
		memcache:    memcache.New("127.0.0.1:11211"),
		now:         func() time.Time { return fixed },
	}

	chatID := uniqueChatID()

	for i := 0; i < 30; i++ {
		if err := rl.WaitIfNeeded(context.Background(), &chatID); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.WaitIfNeeded(ctx, &chatID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected context deadline exceeded")
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("did not actually wait")
	}
}

func TestRateLimiter_ContextTimeout_Distributed(t *testing.T) {
	flushMemcache(t)

	rl := &RateLimiter{
		apiLimiter:  rate.NewLimiter(rate.Inf, 0),
		msgLimiters: make(map[int64]*rate.Limiter),
		memcache:    memcache.New("127.0.0.1:11211"),
		now:         time.Now,
	}

	chatID := uniqueChatID()

	for i := 0; i < 30; i++ {
		if err := rl.WaitIfNeeded(context.Background(), &chatID); err != nil {
			t.Fatal(err)
		}
	}

	// контекст таймаут 50ms
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.WaitIfNeeded(ctx, &chatID)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected context deadline exceeded")
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("did not actually wait")
	}
}
