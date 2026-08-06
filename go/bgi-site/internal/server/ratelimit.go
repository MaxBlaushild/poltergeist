package server

import (
	"sync"

	"golang.org/x/time/rate"
)

// previewRateLimiter mirrors go/reef-site's own — see that file's comment
// for the reasoning (in-memory is fine as long as bgi-site runs inside the
// single composed core process).
type previewRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func newPreviewRateLimiter() *previewRateLimiter {
	return &previewRateLimiter{limiters: make(map[string]*rate.Limiter)}
}

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
