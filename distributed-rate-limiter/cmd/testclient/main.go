package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultURL            = "http://localhost:8080/checkRateLimit"
	defaultTotalRequests  = 100000
	defaultConcurrency    = 400
	defaultUniqueClients  = 100
	defaultRequestTimeout = 5 * time.Second
)

var (
	targetURL     = defaultURL
	totalRequests = defaultTotalRequests
	concurrency   = defaultConcurrency
	uniqueClients = defaultUniqueClients
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

	transport = &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   defaultRequestTimeout,
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
		value = fmt.Sprintf("192.168.1.%d", rand.Intn(uniqueClients))
	} else {
		value = fmt.Sprintf("api-key-%d", rand.Intn(uniqueClients))
	}

	return RlRequest{
		Endpoint:    endpoint,
		ClientKey:   key,
		ClientValue: value,
	}
}

func worker(jobs <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range jobs {
		body := req

		resp, err := httpClient.Post(
			targetURL,
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			failed.Add(1)
			continue
		}

		recordStatus(resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			success.Add(1)
		} else {
			failed.Add(1)
		}
	}
}

func main() {
	targetURL = envOrDefault("TARGET_URL", defaultURL)
	totalRequests = envIntOrDefault("TOTAL_REQUESTS", defaultTotalRequests)
	concurrency = envIntOrDefault("CONCURRENCY", defaultConcurrency)
	uniqueClients = envIntOrDefault("UNIQUE_CLIENTS", defaultUniqueClients)
	if totalRequests <= 0 || concurrency <= 0 || uniqueClients <= 0 {
		log.Fatal("TOTAL_REQUESTS, CONCURRENCY, and UNIQUE_CLIENTS must be positive")
	}

	profilePath := envOrDefault("CPU_PROFILE", "cpu.prof")
	f, err := os.Create(profilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	jobs := make(chan []byte, concurrency)
	// pre-generate requests
	requests := make([][]byte, totalRequests)
	for i := 0; i < totalRequests; i++ {
		req, _ := json.Marshal(randomRequest())
		requests[i] = req
	}

	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for i := 0; i < totalRequests; i++ {
		jobs <- requests[i]
	}

	close(jobs)
	wg.Wait()

	elapsed := time.Since(start)

	fmt.Println("====================================")
	fmt.Printf("Requests      : %d\n", totalRequests)
	fmt.Printf("Concurrency   : %d\n", concurrency)
	fmt.Printf("Success       : %d\n", success.Load())
	fmt.Printf("Failed        : %d\n", failed.Load())
	fmt.Printf("Duration      : %v\n", elapsed)
	fmt.Printf("Req/sec       : %.2f\n", float64(totalRequests)/elapsed.Seconds())
	fmt.Printf("Avg latency   : %v\n", elapsed/time.Duration(totalRequests))

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
		log.Printf("%f percent requests failed\n", (float64(failed.Load())/float64(totalRequests))*100)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid integer for %s: %v", name, err)
	}
	return parsed
}
