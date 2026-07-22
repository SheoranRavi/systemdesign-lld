package distributedratelimiter

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SheoranRavi/drl/config"
	"github.com/SheoranRavi/drl/model"
	"github.com/SheoranRavi/drl/ratelimiter"
	"github.com/SheoranRavi/drl/test"
)

func main() {
	etcd := config.GetEtcd(config.EtcdOptions{Address: "localhost:2379"})
	ctx := context.Background()
	test.PutTestRules(ctx, etcd)
	rlOptions := ratelimiter.RlOptions{RedisAddr: "localhost:6379"}
	rateLimiter := ratelimiter.NewRateLimiter(rlOptions, etcd)

	rlh := NewRateLimitHandler(rateLimiter)
	http.HandleFunc("/checkRateLimit", rlh.IsAllowedHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type RateLimitHandler struct {
	rateLimiter *ratelimiter.RateLimiter
}

func NewRateLimitHandler(rateLimiter *ratelimiter.RateLimiter) *RateLimitHandler {
	return &RateLimitHandler{rateLimiter: rateLimiter}
}

func (rlh *RateLimitHandler) IsAllowedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rlReq RlRequest
	if err := json.NewDecoder(r.Body).Decode(&rlReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if rlReq.Endpoint == "" || rlReq.ClientKey == "" || rlReq.ClientValue == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	req := model.Request{
		Timestamp: time.Now(),
		Endpoint:  rlReq.Endpoint,
		Client: model.Client{
			Key:   model.KeyType(rlReq.ClientKey), // apikey, useridkey, ipaddresskey
			Value: rlReq.ClientValue,
		},
	}
	res := rlh.rateLimiter.IsAllowed(req)
	w.WriteHeader(res.StatusCode)
	resByte, _ := json.Marshal(res)
	w.Write(resByte)
}

type RlRequest struct {
	Endpoint    string `json:"endpoint"`
	ClientKey   string `json:"clientKey"`
	ClientValue string `json:"clientValue"`
}
