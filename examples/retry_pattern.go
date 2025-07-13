package main

import (
	"fmt"
	"math/rand"
	"time"

	goflow "github.com/joemocha/flow/flow"
)

func main() {
	state := goflow.NewSharedState()

	// Automatic retry behavior when retry_max > 0
	node := goflow.NewNode()
	node.SetParams(map[string]interface{}{
		"retry_max":   3,
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
