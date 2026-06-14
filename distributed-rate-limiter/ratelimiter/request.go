package ratelimiter

type Request struct {
	UserId    string
	ApiKey    string
	IpAddress string
}
