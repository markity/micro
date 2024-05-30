package utils

import "sync/atomic"

type NumLimit struct {
	max int64
	cur int64
}

func NewNumLimit(max int64) *NumLimit {
	return &NumLimit{
		max: max,
		cur: 0,
	}
}

func (limit *NumLimit) Add() bool {
	new := atomic.AddInt64(&limit.cur, 1)
	if new > limit.max {
		atomic.AddInt64(&limit.cur, -1)
		return false
	}
	return true
}

func (limit *NumLimit) Sub() {
	atomic.AddInt64(&limit.cur, -1)
}
