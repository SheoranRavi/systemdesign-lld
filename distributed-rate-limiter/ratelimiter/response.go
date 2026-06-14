package ratelimiter

type Response struct {
	StatusCode int
	Message    string
	IsAllowed  bool
}
