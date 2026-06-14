package ratelimiter

import (
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	// needs redis connection
	// needs config
	config RLConfig
	rdp    *redis.Client
}

func (rl *RateLimiter) start() {
	rdb := redis.NewClient(&redis.Options{})
}

func (rl *RateLimiter) IsAllowed(req Request) Response {

}

func Start() *RateLimiter {
	rl := &RateLimiter{}
	rl.start()
	return rl
}
