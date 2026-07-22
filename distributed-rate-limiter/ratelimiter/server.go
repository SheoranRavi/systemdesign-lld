package ratelimiter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SheoranRavi/drl/config"
	"github.com/SheoranRavi/drl/model"
	"github.com/SheoranRavi/drl/util"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// tokenBucketScript refills and consumes one token atomically. RefillRate is
// expressed in tokens per second. Redis TIME is used so all rate-limiter
// instances use the same clock.
var tokenBucketScript = redis.NewScript(`
local now = redis.call('TIME')
local now_ms = now[1] * 1000 + math.floor(now[2] / 1000)

local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local tokens = tonumber(redis.call('HGET', KEYS[1], 'tokens'))
local last_refill_ms = tonumber(redis.call('HGET', KEYS[1], 'last_refill_ms'))

if tokens == nil or last_refill_ms == nil then
    tokens = max_tokens
    last_refill_ms = now_ms
else
    local elapsed_ms = math.max(0, now_ms - last_refill_ms)
    tokens = math.min(max_tokens, tokens + (elapsed_ms * refill_rate / 1000))
end

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call('HSET', KEYS[1], 'tokens', tokens, 'last_refill_ms', now_ms)
redis.call('PEXPIRE', KEYS[1], math.max(1000, math.ceil(max_tokens / refill_rate * 2000)))
return allowed
`)

type RateLimiter struct {
	// needs redis connection
	// needs config
	options   RlOptions
	rules     map[string]model.Rule
	ruleStore *config.Etcd
	rdb       *redis.Client
}

func NewRateLimiter(rlOptions RlOptions, ruleStore *config.Etcd) *RateLimiter {
	rl := &RateLimiter{options: rlOptions, ruleStore: ruleStore}
	rl.start()
	return rl
}

func (rl *RateLimiter) loadRules() {
	rl.rules = rl.ruleStore.GetAllRules(context.Background())
}

func (rl *RateLimiter) startWatching() {
	wChan := rl.ruleStore.GetWatchCh(context.Background())
	for wr := range wChan {
		for _, ev := range wr.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				var rule model.Rule
				json.Unmarshal(ev.Kv.Value, &rule)
				// update in-memory cache
				rl.rules[string(ev.Kv.Key)] = rule
			case clientv3.EventTypeDelete:
				// remove from cache
				delete(rl.rules, string(ev.Kv.Key))
			}
		}
	}
}

func (rl *RateLimiter) start() {
	rl.loadRules()
	go rl.startWatching()
	rl.rdb = redis.NewClient(&redis.Options{
		Addr: rl.options.RedisAddr,
	})
}

func (rl *RateLimiter) IsAllowed(req model.Request) model.Response {
	// look at the rule
	var rule model.Rule
	var ruleKey string
	switch req.Client.Key {
	case model.IpAddressKey:
		ruleKey = util.GetRuleKey(model.UnAuthedRule.String(), req.Endpoint)
		rule = rl.rules[ruleKey]
	case model.UserIdKey:
		ruleKey = util.GetRuleKey(model.UserIdRule.String(), req.Endpoint)
		rule = rl.rules[ruleKey]
	case model.ApiKey:
		ruleKey = util.GetRuleKey(model.ApiKeyRule.String(), req.Endpoint)
		rule = rl.rules[ruleKey]
	default:
		return model.Response{IsAllowed: false, StatusCode: http.StatusBadRequest, Message: "No matching rule for request key"}
	}

	if rule.MaxTokens <= 0 || rule.RefillRate <= 0 {
		return model.Response{IsAllowed: false, StatusCode: http.StatusInternalServerError, Message: "Invalid rate-limit rule"}
	}
	if rl.rdb == nil {
		return model.Response{IsAllowed: false, StatusCode: http.StatusServiceUnavailable, Message: "Rate limiter is not initialized"}
	}

	// Keep a separate bucket for every client, rule bucket, and endpoint.
	// The key is passed as KEYS[1]; limits are passed as ARGV[1] and ARGV[2].
	bucketKey := fmt.Sprintf("ratelimit:%d:%s:%s", req.Client.Key, ruleKey, req.Client.Value)
	allowed, err := tokenBucketScript.Run(
		context.Background(), rl.rdb, []string{bucketKey}, rule.MaxTokens, rule.RefillRate,
	).Int()
	if err != nil {
		return model.Response{IsAllowed: false, StatusCode: http.StatusServiceUnavailable, Message: "Rate limiter unavailable"}
	}
	if allowed == 1 {
		return model.Response{IsAllowed: true, StatusCode: http.StatusOK, Message: "Allowed"}
	}
	return model.Response{IsAllowed: false, StatusCode: http.StatusTooManyRequests, Message: "Rate limit exceeded"}
}

type RlOptions struct {
	RedisAddr string
}
