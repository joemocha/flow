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

	// Automatic retry behavior when retries > 0
	node := flow.NewNode()
	node.SetParams(map[string]interface{}{
		"retries":     3,
		"retry_delay": time.Millisecond * 100,
	})
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		// Just business logic - retry is automatic!
		// Generate cryptographically secure random number (0-99)
		n, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return "", fmt.Errorf("random generation failed: %v", err)
		}
		if n.Int64() < 70 { // 70% chance of failure (equivalent to < 0.7)
			return "", fmt.Errorf("API temporarily unavailable")
		}
		return "api_success", nil
	})

	result := node.Run(state)
	fmt.Printf("Final result: %s\n", result)
}
