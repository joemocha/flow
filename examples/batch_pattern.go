package main

import (
	"fmt"
	goflow "github.com/sam/goflow/flow"
)

func main() {
	state := goflow.NewSharedState()

	// Automatic batch processing when batch_data is present
	node := goflow.NewNode()
	node.SetParams(map[string]interface{}{
		"batch_data": []int{1, 2, 3, 4, 5},
	})
	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		// Called once per item automatically!
		num := item.(int)
		return fmt.Sprintf("processed-%d", num*2), nil
	})

	result := node.Run(state)
	fmt.Printf("Batch result: %s\n", result)

	// Results are automatically stored in shared state
	results := state.Get("batch_results")
	fmt.Printf("Processed items: %v\n", results)
}