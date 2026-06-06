// Package ratelimit provides fixed-window per-key request limits.
package ratelimit

import (
	"sync"
	"time"
)

type counter struct {
	count int
	start time.Time
}

// Limiter enforces a fixed-window request cap per key.
type Limiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	keys   map[string]*counter
}

// New returns a limiter that allows limit events per key within window.
func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:  limit,
		window: window,
		keys:   make(map[string]*counter),
	}
}

// Allow reports whether key is within its rate limit.
func (l *Limiter) Allow(key string) bool {
	if l.limit <= 0 {
		return true
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	c, ok := l.keys[key]
	if !ok || now.Sub(c.start) >= l.window {
		l.keys[key] = &counter{count: 1, start: now}
		return true
	}
	if c.count >= l.limit {
		return false
	}
	c.count++
	return true
}

// Cleanup removes expired counters from memory.
func (l *Limiter) Cleanup() {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, c := range l.keys {
		if now.Sub(c.start) >= l.window {
			delete(l.keys, key)
		}
	}
}
