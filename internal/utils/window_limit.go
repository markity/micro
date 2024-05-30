package utils

import (
	"sync"
	"time"
)

type WindowLimit struct {
	mu         sync.Mutex
	window     []time.Time
	timeWindow time.Duration
	limit      int
}

func (limit *WindowLimit) Add() bool {
	limit.mu.Lock()
	defer limit.mu.Unlock()

	now := time.Now()
	if len(limit.window) < limit.limit {
		limit.window = append(limit.window, now)
		return true
	}

	if now.After(limit.window[0].Add(limit.timeWindow)) {
		limit.window = limit.window[1:]
		limit.window = append(limit.window, now)
		return true
	}

	return false
}

func NewWindowLimit(dur time.Duration, limit int) *WindowLimit {
	if limit <= 0 {
		panic("check it!!")
	}
	return &WindowLimit{
		window:     make([]time.Time, 0),
		limit:      limit,
		timeWindow: dur,
	}
}
