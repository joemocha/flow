package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	flow "github.com/joemocha/flow"
)

func main() {
	state := flow.NewSharedState()

	// Composed behavior: batch + parallel + retry in single node!
	node := flow.NewNode()
	node.SetParams(map[string]interface{}{
		// Batch configuration
		"data": []string{
			"https://api1.example.com",
			"https://api2.example.com",
			"https://api3.example.com",
			"https://api4.example.com",
			"https://api5.example.com",
		},
		"batch": true,
		// Parallel configuration
		"parallel":       true,
		"parallel_limit": 2, // Max 2 concurrent requests
		// Retry configuration
		"retries":     3,
		"retry_delay": time.Millisecond * 200,
	})

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		// Pure business logic - all patterns applied automatically!
		url := item.(string)

		// Simulate API call that might fail
		// Generate cryptographically secure random number (0-99)
		n, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return "", fmt.Errorf("random generation failed: %v", err)
		}
		if n.Int64() < 60 { // 60% chance of failure (equivalent to < 0.6)
			return "", fmt.Errorf("failed to fetch %s", url)
		}

		// Simulate processing time
		time.Sleep(time.Millisecond * 100)
		return fmt.Sprintf("data from %s", url), nil
	})

	start := time.Now()
	result := node.Run(state)
	elapsed := time.Since(start)

	fmt.Printf("Composed result: %s\n", result)
	fmt.Printf("Execution time: %v\n", elapsed)

	// Results automatically collected
	results := state.Get("batch_results")
	fmt.Printf("Fetched data: %v\n", results)

	fmt.Println("\nBehaviors applied automatically:")
	fmt.Println("✓ Batch processing (5 URLs)")
	fmt.Println("✓ Parallel execution (max 2 concurrent)")
	fmt.Println("✓ Retry logic (up to 3 attempts per URL)")
	fmt.Println("✓ Results collection")
}
