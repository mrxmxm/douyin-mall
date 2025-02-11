package ratelimit

import (
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r float64, b int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(r), b),
	}
}

func (l *RateLimiter) Allow() bool {
	return l.limiter.Allow()
}
