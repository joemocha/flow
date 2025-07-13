package main

import (
	"fmt"

	goflow "github.com/joemocha/flow/flow"
)

func main() {
	state := goflow.NewSharedState()

	// Create adaptive node with just parameters and business logic
	node := goflow.NewNode()
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
