package main

import (
"fmt"

flow "github.com/joemocha/flow"
)

// NewDataProcessorNode creates a node that processes input data
func NewDataProcessorNode() *flow.Node {
	node := flow.NewNode()
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
return "processed", nil
})
	
	node.SetPrepFunc(func(shared *flow.SharedState) interface{} {
value := shared.GetInt("input")
processed := value * 2

shared.Set("processed_value", processed)
fmt.Printf("Processed: %d -> %d\n", value, processed)

return processed
})
	
	return node
}

// NewValidatorNode creates a node that validates processed data
func NewValidatorNode() *flow.Node {
	node := flow.NewNode()
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
return "validation_complete", nil
})
	
	node.SetPrepFunc(func(shared *flow.SharedState) interface{} {
value := shared.GetInt("processed_value")
return value
})
	
	node.SetPostFunc(func(shared *flow.SharedState, prepResult interface{}, execResult interface{}) string {
value := prepResult.(int)

if value > 10 {
			shared.Set("validation_result", "valid")
			fmt.Printf("Validation: %d is valid (> 10)\n", value)
			return "valid"
		} else {
			shared.Set("validation_result", "invalid")
			fmt.Printf("Validation: %d is invalid (<= 10)\n", value)
			return "invalid"
		}
	})
	
	return node
}

// NewOutputNode creates a node that outputs final results
func NewOutputNode() *flow.Node {
	node := flow.NewNode()
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
return "output_complete", nil
})
	
	node.SetPrepFunc(func(shared *flow.SharedState) interface{} {
validationResult := shared.Get("validation_result")
processedValue := shared.GetInt("processed_value")

if validationResult == "valid" {
result := fmt.Sprintf("SUCCESS: Processed value %d is valid", processedValue)
shared.Set("final_result", result)
fmt.Println(result)
} else {
result := fmt.Sprintf("REJECTED: Processed value %d is invalid", processedValue)
shared.Set("final_result", result)
fmt.Println(result)
}

return validationResult
})
	
	return node
}

func main() {
	state := flow.NewSharedState()
	state.Set("input", 7)

	// Build workflow: Process -> Validate -> Output
	processor := NewDataProcessorNode()
	validator := NewValidatorNode()
	validOutput := NewOutputNode()
	invalidOutput := NewOutputNode()

	// Chain nodes with conditional branching
	processor.Next(validator, "processed")
	validator.Next(validOutput, "valid")
	validator.Next(invalidOutput, "invalid")

	// Create and run flow
	workflow := flow.NewFlow().Start(processor)
	result := workflow.Run(state)

	fmt.Printf("Workflow result: %s\n", result)
	fmt.Printf("Final value: %v\n", state.Get("final_result"))
}
