package main

import (
	"fmt"
	"math/rand"
	"time"

	Flow "github.com/joemocha/flow"
)

func main() {
	state := Flow.NewSharedState()

	// Automatic retry behavior when retries > 0
	node := Flow.NewNode()
	node.SetParams(map[string]interface{}{
		"retries":     3,
		"retry_delay": time.Millisecond * 100,
	})
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		// Just business logic - retry is automatic!
		if rand.Float32() < 0.7 {
			return "", fmt.Errorf("API temporarily unavailable")
		}
		return "api_success", nil
	})

	result := node.Run(state)
	fmt.Printf("Final result: %s\n", result)
}
