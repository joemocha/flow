package main

import (
"fmt"

flow "github.com/joemocha/flow"
)

func main() {
	state := flow.NewSharedState()

	// Automatic batch processing when batch: true is set
	node := flow.NewNode()
	node.SetParams(map[string]interface{}{
"data":  []int{1, 2, 3, 4, 5},
"batch": true,
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
