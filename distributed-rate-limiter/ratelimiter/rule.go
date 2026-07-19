package ratelimiter

type Rule struct {
	RefillRate int
	MaxTokens  int
	Endpoint   string
	RuleBucket RuleBucket
}

type RuleBucket int

const (
	UnAuthedRule RuleBucket = iota
	AuthorizedRule
	PremiumRule
	IpAddressRule
)
