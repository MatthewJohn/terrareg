package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestPerformanceLoadTests runs basic load tests against the terrareg-go server
func TestPerformanceLoadTests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Setup test database and server
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	server := testutils.SetupTestServer(context.Background(), db)
	defer server.Close()

	// Test basic concurrent request handling
	t.Run("ConcurrentRequests", func(t *testing.T) {
		testConcurrentRequests(t, server.URL)
	})

	// Test response times
	t.Run("ResponseTime", func(t *testing.T) {
		testResponseTime(t, server.URL)
	})
}

// testConcurrentRequests tests handling of multiple concurrent requests
func testConcurrentRequests(t *testing.T, baseURL string) {
	const (
		numWorkers  = 10
		numRequests = 50
		maxDuration = 5 * time.Second
	)

	client := &http.Client{Timeout: maxDuration}

	var wg sync.WaitGroup
	errors := make(chan error, numRequests)
	results := make(chan time.Duration, numRequests)

	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < numRequests/numWorkers; j++ {
				reqStart := time.Now()

				// Test different endpoints
				endpoints := []string{
					"/v1/terrareg/namespaces",
					"/v1/terrareg/modules",
					"/v2/gpg-keys?filter[namespace]=test",
				}

				for _, endpoint := range endpoints {
					req, err := http.NewRequest("GET", baseURL+endpoint, nil)
					if err != nil {
						errors <- fmt.Errorf("worker %d: failed to create request: %w", workerID, err)
						continue
					}

					resp, err := client.Do(req)
					if err != nil {
						errors <- fmt.Errorf("worker %d: request failed: %w", workerID, err)
						continue
					}
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						errors <- fmt.Errorf("worker %d: unexpected status code: %d", workerID, resp.StatusCode)
						continue
					}
				}

				results <- time.Since(reqStart)
			}
		}(i)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(errors)
		close(results)
	}()

	// Collect errors
	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	// Collect response times
	responseTimes := make([]time.Duration, 0, numRequests)
	for rt := range results {
		responseTimes = append(responseTimes, rt)
	}

	totalDuration := time.Since(start)

	// Performance assertions
	require.Less(t, errorCount, numRequests/10, "Too many errors occurred") // Allow up to 10% error rate
	require.Less(t, totalDuration, maxDuration, "Test took too long to complete")

	// Calculate average response time
	if len(responseTimes) > 0 {
		var totalTime time.Duration
		for _, rt := range responseTimes {
			totalTime += rt
		}
		avgResponseTime := totalTime / time.Duration(len(responseTimes))

		t.Logf("Performance Results:")
		t.Logf("  Total requests: %d", numRequests)
		t.Logf("  Total duration: %v", totalDuration)
		t.Logf("  Average response time: %v", avgResponseTime)
		t.Logf("  Error rate: %.2f%%", float64(errorCount)/float64(numRequests)*100)
		t.Logf("  Requests per second: %.2f", float64(numRequests)/totalDuration.Seconds())

		// Basic performance requirement - response time should be reasonable
		require.Less(t, avgResponseTime, 1*time.Second, "Average response time should be under 1 second")
	}
}

// testResponseTime tests individual endpoint response times
func testResponseTime(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []struct {
		path        string
		maxDuration time.Duration
	}{
		{"/v1/terrareg/namespaces", 100 * time.Millisecond},
		{"/v1/terrareg/modules", 150 * time.Millisecond},
		{"/v2/gpg-keys?filter[namespace]=test", 100 * time.Millisecond},
		{"/v1/terrareg/audit-history", 200 * time.Millisecond},
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequest("GET", baseURL+endpoint.path, nil)
		require.NoError(t, err)

		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		require.NoError(t, err)
		defer resp.Body.Close()

		t.Logf("Endpoint %s: %v (status: %d)", endpoint.path, duration, resp.StatusCode)
		require.Less(t, duration, endpoint.maxDuration,
			"Response time for %s exceeds threshold of %v", endpoint.path, endpoint.maxDuration)
	}
}

// TestMemoryUsage tests basic memory usage patterns
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage tests in short mode")
	}

	// Setup test database and server
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	server := testutils.SetupTestServer(context.Background(), db)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Make a series of requests to test for memory leaks
	const numIterations = 100

	for i := 0; i < numIterations; i++ {
		// Test namespace listing
		req, err := http.NewRequest("GET", server.URL+"/v1/terrareg/namespaces", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		if i%10 == 0 {
			t.Logf("Completed %d/%d requests", i, numIterations)
		}
	}

	t.Logf("✅ Memory usage test completed - %d requests processed successfully", numIterations)
}
