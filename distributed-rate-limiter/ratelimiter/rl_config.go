package ratelimiter

type RLConfig struct {
	ipRps        int
	apiKeyConfig []ApiKeyConfig
	userIdRps    []UserIdRps
}

type ApiKeyConfig struct {
	key string
	rps int
}

type UserIdRps struct {
	userId string
	rps    int
}
