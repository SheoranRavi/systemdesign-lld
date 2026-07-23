package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	URL            = "http://localhost:8080/checkRateLimit"
	TotalRequests  = 10000
	Concurrency    = 100
	UniqueClients  = 100
	RequestTimeout = 5 * time.Second
)

type RlRequest struct {
	Endpoint    string `json:"endpoint"`
	ClientKey   string `json:"clientKey"`
	ClientValue string `json:"clientValue"`
}

var (
	endpoints = []string{
		"/login",
		"/signup",
		"/getgeo",
	}

	endpointClientKeys = map[string]string{
		"/login":  "IpAddressKey",
		"/signup": "IpAddressKey",
		"/getgeo": "ApiKey",
	}

	httpClient = &http.Client{
		Timeout: RequestTimeout,
	}

	success atomic.Int64
	failed  atomic.Int64

	statusMu    sync.Mutex
	statusCount = make(map[int]int64)
)

func recordStatus(statusCode int) {
	statusMu.Lock()
	statusCount[statusCode]++
	statusMu.Unlock()
}

func randomRequest() RlRequest {
	endpoint := endpoints[rand.Intn(len(endpoints))]
	key := endpointClientKeys[endpoint]

	var value string
	if key == "IpAddressKey" {
		value = fmt.Sprintf("192.168.1.%d", rand.Intn(UniqueClients))
	} else {
		value = fmt.Sprintf("api-key-%d", rand.Intn(UniqueClients))
	}

	return RlRequest{
		Endpoint:    endpoint,
		ClientKey:   key,
		ClientValue: value,
	}
}

func worker(jobs <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for range jobs {
		req := randomRequest()

		body, err := json.Marshal(req)
		if err != nil {
			failed.Add(1)
			continue
		}

		resp, err := httpClient.Post(
			URL,
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			failed.Add(1)
			continue
		}

		recordStatus(resp.StatusCode)

		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			success.Add(1)
		} else {
			failed.Add(1)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	jobs := make(chan struct{}, Concurrency)

	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for i := 0; i < TotalRequests; i++ {
		jobs <- struct{}{}
	}

	close(jobs)
	wg.Wait()

	elapsed := time.Since(start)

	fmt.Println("====================================")
	fmt.Printf("Requests      : %d\n", TotalRequests)
	fmt.Printf("Concurrency   : %d\n", Concurrency)
	fmt.Printf("Success       : %d\n", success.Load())
	fmt.Printf("Failed        : %d\n", failed.Load())
	fmt.Printf("Duration      : %v\n", elapsed)
	fmt.Printf("Req/sec       : %.2f\n", float64(TotalRequests)/elapsed.Seconds())
	fmt.Printf("Avg latency   : %v\n", elapsed/time.Duration(TotalRequests))

	statusMu.Lock()
	statusCodes := make([]int, 0, len(statusCount))
	for statusCode := range statusCount {
		statusCodes = append(statusCodes, statusCode)
	}
	sort.Ints(statusCodes)
	for _, statusCode := range statusCodes {
		fmt.Printf("HTTP %d       : %d\n", statusCode, statusCount[statusCode])
	}
	statusMu.Unlock()

	fmt.Println("====================================")

	if failed.Load() > 0 {
		log.Printf("%d requests failed\n", failed.Load())
		log.Printf("%f percent requests failed\n", (float64(failed.Load())/float64(TotalRequests))*100)
	}
}
