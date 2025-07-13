package main

import (
	"fmt"

	Flow "github.com/joemocha/flow"
)

// DataProcessorNode processes input data
type DataProcessorNode struct {
	baseNode *Flow.BaseNode
}

func NewDataProcessorNode() *DataProcessorNode {
	return &DataProcessorNode{baseNode: Flow.NewBaseNode()}
}

func (n *DataProcessorNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *DataProcessorNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *DataProcessorNode) Next(node Flow.Node, action string) Flow.Node {
	return n.baseNode.Next(node, action)
}
func (n *DataProcessorNode) GetSuccessors() map[string]Flow.Node { return n.baseNode.GetSuccessors() }
func (n *DataProcessorNode) Prep(shared *Flow.SharedState) interface{} {
	return n.baseNode.Prep(shared)
}
func (n *DataProcessorNode) Post(shared *Flow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *DataProcessorNode) Exec(prepResult interface{}) (string, error) {
	return "processed", nil
}

func (n *DataProcessorNode) Run(shared *Flow.SharedState) string {
	prepResult := n.Prep(shared)

	// Get input value
	value := shared.GetInt("input")
	processed := value * 2

	shared.Set("processed_value", processed)
	fmt.Printf("Processed: %d -> %d\n", value, processed)

	execResult, _ := n.Exec(prepResult)
	return n.Post(shared, prepResult, execResult)
}

// ValidatorNode checks if processed data meets criteria
type ValidatorNode struct {
	baseNode *Flow.BaseNode
}

func NewValidatorNode() *ValidatorNode {
	return &ValidatorNode{baseNode: Flow.NewBaseNode()}
}

func (n *ValidatorNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *ValidatorNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *ValidatorNode) Next(node Flow.Node, action string) Flow.Node {
	return n.baseNode.Next(node, action)
}
func (n *ValidatorNode) GetSuccessors() map[string]Flow.Node       { return n.baseNode.GetSuccessors() }
func (n *ValidatorNode) Prep(shared *Flow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *ValidatorNode) Post(shared *Flow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *ValidatorNode) Exec(prepResult interface{}) (string, error) {
	return "validated", nil
}

func (n *ValidatorNode) Run(shared *Flow.SharedState) string {
	prepResult := n.Prep(shared)

	value := shared.GetInt("processed_value")

	if value > 10 {
		fmt.Printf("Validation: %d is valid (> 10)\n", value)
		return n.Post(shared, prepResult, "valid")
	} else {
		fmt.Printf("Validation: %d is invalid (<= 10)\n", value)
		return n.Post(shared, prepResult, "invalid")
	}
}

// OutputNode handles final output
type OutputNode struct {
	baseNode *Flow.BaseNode
}

func NewOutputNode() *OutputNode {
	return &OutputNode{baseNode: Flow.NewBaseNode()}
}

func (n *OutputNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *OutputNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *OutputNode) Next(node Flow.Node, action string) Flow.Node {
	return n.baseNode.Next(node, action)
}
func (n *OutputNode) GetSuccessors() map[string]Flow.Node       { return n.baseNode.GetSuccessors() }
func (n *OutputNode) Prep(shared *Flow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *OutputNode) Post(shared *Flow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *OutputNode) Exec(prepResult interface{}) (string, error) {
	return "output_complete", nil
}

func (n *OutputNode) Run(shared *Flow.SharedState) string {
	prepResult := n.Prep(shared)

	value := shared.GetInt("processed_value")
	fmt.Printf("Final output: %d\n", value)
	shared.Set("final_result", value)

	execResult, _ := n.Exec(prepResult)
	return n.Post(shared, prepResult, execResult)
}

func main() {
	state := Flow.NewSharedState()
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
	flow := Flow.NewFlow().Start(processor)
	result := flow.Run(state)

	fmt.Printf("Workflow result: %s\n", result)
	fmt.Printf("Final value: %v\n", state.Get("final_result"))
}
