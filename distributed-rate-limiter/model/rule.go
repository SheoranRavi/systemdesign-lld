package model

import "fmt"

type Rule struct {
	RefillRate int
	MaxTokens  int
	Endpoint   string
	RuleBucket RuleBucket
}

type RuleBucket int

const (
	UnAuthedRule RuleBucket = iota
	ApiKeyRule              // to limit individual api keys
	UserIdRule              // to limit by user Id across all clients (ios, web, etc)
	PremiumRule
)

func (r RuleBucket) String() string {
	switch r {
	case UnAuthedRule:
		return "UnAuthedRule"
	case ApiKeyRule:
		return "ApiKeyRule"
	case UserIdRule:
		return "UserIdRule"
	case PremiumRule:
		return "PremiumRule"
	default:
		return fmt.Sprintf("RuleBucket(%d)", r)
	}
}

type KeyType string

const (
	ApiKey       KeyType = "ApiKey"
	UserIdKey    KeyType = "UserIdKey"
	IpAddressKey KeyType = "IpAddressKey"
)
