package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type StatsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type StatsResponse struct {
	Stats []struct {
		Timestamp string `json:"ts"`
		Value     int    `json:"v"`
	} `json:"stats"`
}

type TestResult struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalDuration      time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	AvgLatency         time.Duration
	RPS                float64
}

type RequestResult struct {
	Duration time.Duration
	Error    error
}

func main() {
	baseURL := flag.String("url", "http://localhost:3000", "Base URL of the service")
	numRequests := flag.Int("requests", 1000, "Number of total requests to send")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent requests")
	bannerID := flag.Int("banner", 5, "Banner ID to test")
	flag.Parse()

	var wg sync.WaitGroup

	clickResults := make(chan RequestResult, *numRequests)
	statsResults := make(chan RequestResult, *numRequests)

	sem := make(chan struct{}, *concurrency)

	startTime := time.Now()

	// Launch goroutines for click requests
	for i := 0; i < *numRequests; i++ {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			start := time.Now()
			url := fmt.Sprintf("%s/counter/%d", *baseURL, *bannerID)
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				log.Printf("error making click request: %v", err)
				clickResults <- RequestResult{Error: err}
				return
			}
			resp.Body.Close()
			clickResults <- RequestResult{Duration: time.Since(start)}
		}()
	}

	// Launch goroutines for stats requests
	for i := 0; i < *numRequests; i++ {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			start := time.Now()
			url := fmt.Sprintf("%s/stats/%d", *baseURL, *bannerID)

			reqBody := StatsRequest{
				From: time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05"),
				To:   time.Now().Format("2006-01-02T15:04:05"),
			}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				log.Printf("error making stats request: %v", err)
				statsResults <- RequestResult{Error: err}
				return
			}
			resp.Body.Close()
			statsResults <- RequestResult{Duration: time.Since(start)}
		}()
	}

	wg.Wait()
	close(clickResults)
	close(statsResults)

	clickTestResult := calculateResults(clickResults, time.Since(startTime))
	statsTestResult := calculateResults(statsResults, time.Since(startTime))

	fmt.Printf("\nClick Endpoint Results:\n")
	printResults(clickTestResult)
	fmt.Printf("\nStats Endpoint Results:\n")
	printResults(statsTestResult)
	fmt.Printf("\nTotal Test Duration: %v\n", time.Since(startTime))
}

func calculateResults(results chan RequestResult, totalDuration time.Duration) TestResult {
	var result TestResult
	var totalLatency time.Duration

	result.MinLatency = time.Hour // initialize with a large value

	for res := range results {
		result.TotalRequests++
		if res.Error != nil {
			result.FailedRequests++
			continue
		}

		result.SuccessfulRequests++
		totalLatency += res.Duration

		if res.Duration < result.MinLatency {
			result.MinLatency = res.Duration
		}
		if res.Duration > result.MaxLatency {
			result.MaxLatency = res.Duration
		}
	}

	if result.SuccessfulRequests > 0 {
		result.AvgLatency = totalLatency / time.Duration(result.SuccessfulRequests)
		result.RPS = float64(result.SuccessfulRequests) / totalDuration.Seconds()
	}

	return result
}

func printResults(result TestResult) {
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", result.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", result.FailedRequests)
	fmt.Printf("RPS: %.2f\n", result.RPS)
	fmt.Printf("Min Latency: %v\n", result.MinLatency)
	fmt.Printf("Max Latency: %v\n", result.MaxLatency)
	fmt.Printf("Average Latency: %v\n", result.AvgLatency)
}
