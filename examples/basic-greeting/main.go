package main

import (
"fmt"

flow "github.com/joemocha/flow"
)

func main() {
	state := flow.NewSharedState()

	// Create adaptive node with just parameters and business logic
	node := flow.NewNode()
	node.SetParams(map[string]interface{}{
"name": "World",
})
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
name := node.GetParam("name").(string)
fmt.Printf("Hello, %s!\n", name)
return "greeted", nil
})

	result := node.Run(state)
	fmt.Printf("Result: %s\n", result)
}
