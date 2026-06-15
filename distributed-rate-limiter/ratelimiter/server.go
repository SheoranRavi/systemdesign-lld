package ratelimiter

import (
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	// needs redis connection
	// needs config
	config  RLConfig
	options RlOptions
	rdb     *redis.Client
}

func (rl *RateLimiter) start() {
	rl.rdb = redis.NewClient(&redis.Options{
		Addr:     rl.options.RedisAddr,
		Password: rl.options.RedisPass,
		DB:       rl.options.RedisDb,
	})
}

func (rl *RateLimiter) IsAllowed(req Request) Response {

}

func Start(rlOptions RlOptions) *RateLimiter {
	rl := &RateLimiter{options: rlOptions}
	rl.start()
	return rl
}

type RlOptions struct {
	RedisAddr string
	RedisPass string
	RedisDb   int
}
