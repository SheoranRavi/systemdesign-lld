package ratelimiter

type Client struct {
	Key   string
	Value string
}

type KeyType int

const (
	ApiKey KeyType = iota
	UserIdKey
	IpAddressKey
)
