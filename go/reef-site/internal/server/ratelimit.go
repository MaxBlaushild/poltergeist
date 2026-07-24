package server

import (
	"sync"

	"golang.org/x/time/rate"
)

// previewRateLimiter is R-2.6's "Preview requests are rate-limited per
// session" — a small per-process token bucket per session ID. This is
// deliberately not distributed (no Redis-backed limiter): reef-site runs as
// part of the single composed core process today (see INVENTORY.md), so an
// in-memory limiter is exactly as effective as a distributed one until that
// changes, and adds no new infrastructure dependency.
type previewRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func newPreviewRateLimiter() *previewRateLimiter {
	return &previewRateLimiter{limiters: make(map[string]*rate.Limiter)}
}

// allow permits up to 3 preview requests/second per session, bursting to 5 —
// generous enough for live slider-dragging (debounced client-side to 300ms
// per R-2.6, i.e. ~3.3/s at most from a well-behaved client) while still
// bounding a runaway or abusive client.
func (l *previewRateLimiter) allow(sessionID string) bool {
	l.mu.Lock()
	limiter, ok := l.limiters[sessionID]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(3), 5)
		l.limiters[sessionID] = limiter
	}
	l.mu.Unlock()
	return limiter.Allow()
}
