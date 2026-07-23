package test

import (
	"context"
	"encoding/json"

	"github.com/SheoranRavi/drl/config"
	"github.com/SheoranRavi/drl/model"
	"github.com/SheoranRavi/drl/util"
)

func PutTestRules(ctx context.Context, etcd *config.Etcd) {
	// need rules for each bucket
	rule := model.Rule{
		RuleBucket: model.UnAuthedRule,
		Endpoint:   "/login",
		MaxTokens:  10000,
		RefillRate: 5000,
	}

	ruleString, _ := json.Marshal(rule)
	etcd.Put(ctx, util.GetRuleKey(model.UnAuthedRule.String(), "/login"), string(ruleString))

	rule = model.Rule{
		RuleBucket: model.UnAuthedRule,
		Endpoint:   "/signup",
		MaxTokens:  2000,
		RefillRate: 1000,
	}
	ruleString, _ = json.Marshal(rule)
	etcd.Put(ctx, util.GetRuleKey(model.UnAuthedRule.String(), "/signup"), string(ruleString))

	rule = model.Rule{
		RuleBucket: model.ApiKeyRule,
		Endpoint:   "/getgeo",
		MaxTokens:  10000,
		RefillRate: 1000,
	}
	ruleString, _ = json.Marshal(rule)
	etcd.Put(ctx, util.GetRuleKey(model.ApiKeyRule.String(), "/getgeo"), string(ruleString))
}
