package ratelimiter

import "time"

type Request struct {
	Client    Client
	Timestamp time.Time
	Endpoint  string
}
