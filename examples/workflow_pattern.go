package main

import (
	"fmt"
	goflow "github.com/sam/goflow/flow"
)

// DataProcessorNode processes input data
type DataProcessorNode struct {
	baseNode *goflow.BaseNode
}

func NewDataProcessorNode() *DataProcessorNode {
	return &DataProcessorNode{baseNode: goflow.NewBaseNode()}
}

func (n *DataProcessorNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *DataProcessorNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *DataProcessorNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}
func (n *DataProcessorNode) GetSuccessors() map[string]goflow.Node { return n.baseNode.GetSuccessors() }
func (n *DataProcessorNode) Prep(shared *goflow.SharedState) interface{} {
	return n.baseNode.Prep(shared)
}
func (n *DataProcessorNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *DataProcessorNode) Exec(prepResult interface{}) (string, error) {
	return "processed", nil
}

func (n *DataProcessorNode) Run(shared *goflow.SharedState) string {
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
	baseNode *goflow.BaseNode
}

func NewValidatorNode() *ValidatorNode {
	return &ValidatorNode{baseNode: goflow.NewBaseNode()}
}

func (n *ValidatorNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *ValidatorNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *ValidatorNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}
func (n *ValidatorNode) GetSuccessors() map[string]goflow.Node       { return n.baseNode.GetSuccessors() }
func (n *ValidatorNode) Prep(shared *goflow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *ValidatorNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *ValidatorNode) Exec(prepResult interface{}) (string, error) {
	return "validated", nil
}

func (n *ValidatorNode) Run(shared *goflow.SharedState) string {
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
	baseNode *goflow.BaseNode
}

func NewOutputNode() *OutputNode {
	return &OutputNode{baseNode: goflow.NewBaseNode()}
}

func (n *OutputNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *OutputNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *OutputNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}
func (n *OutputNode) GetSuccessors() map[string]goflow.Node       { return n.baseNode.GetSuccessors() }
func (n *OutputNode) Prep(shared *goflow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *OutputNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}
func (n *OutputNode) Exec(prepResult interface{}) (string, error) {
	return "output_complete", nil
}

func (n *OutputNode) Run(shared *goflow.SharedState) string {
	prepResult := n.Prep(shared)

	value := shared.GetInt("processed_value")
	fmt.Printf("Final output: %d\n", value)
	shared.Set("final_result", value)

	execResult, _ := n.Exec(prepResult)
	return n.Post(shared, prepResult, execResult)
}

func main() {
	state := goflow.NewSharedState()
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
	flow := goflow.NewFlow().Start(processor)
	result := flow.Run(state)

	fmt.Printf("Workflow result: %s\n", result)
	fmt.Printf("Final value: %v\n", state.Get("final_result"))
}
