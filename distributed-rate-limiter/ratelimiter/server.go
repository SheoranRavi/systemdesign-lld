package ratelimiter

import (
	"net/http"

	"github.com/SheoranRavi/drl/config"
	"github.com/SheoranRavi/drl/model"
	"github.com/SheoranRavi/drl/util"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	// needs redis connection
	// needs config
	config    RLConfig
	options   RlOptions
	rules     map[string]model.Rule
	ruleStore *config.Etcd
	rdb       *redis.Client
}

func (rl *RateLimiter) loadRules() {
	rl.rules = rl.ruleStore.GetAllRules()
}

func (rl *RateLimiter) start() {
	rl.loadRules()
	rl.rdb = redis.NewClient(&redis.Options{
		Addr:     rl.options.RedisAddr,
		Password: rl.options.RedisPass,
		DB:       rl.options.RedisDb,
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
	// call Redis lua script
	// response
	return model.Response{}
}

func NewRateLimiter(rlOptions RlOptions) *RateLimiter {
	rl := &RateLimiter{options: rlOptions}
	rl.start()
	return rl
}

type RlOptions struct {
	RedisAddr string
	RedisPass string
	RedisDb   int
}
