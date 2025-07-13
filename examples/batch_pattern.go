package main

import (
	"fmt"
	goflow "github.com/sam/goflow/flow"
)

// BatchNode demonstrates user-built batch processing pattern
type BatchNode struct {
	baseNode    *goflow.BaseNode
	ProcessFunc func(interface{}) (interface{}, error)
}

func NewBatchNode(processFunc func(interface{}) (interface{}, error)) *BatchNode {
	return &BatchNode{
		baseNode:    goflow.NewBaseNode(),
		ProcessFunc: processFunc,
	}
}

// Delegate BaseNode methods
func (n *BatchNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *BatchNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *BatchNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}
func (n *BatchNode) GetSuccessors() map[string]goflow.Node       { return n.baseNode.GetSuccessors() }
func (n *BatchNode) Prep(shared *goflow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *BatchNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}

// User-implemented batch processing
func (n *BatchNode) Exec(prepResult interface{}) (string, error) {
	// This would normally be overridden in concrete implementations
	return "batch_processed", nil
}

func (n *BatchNode) Run(shared *goflow.SharedState) string {
	prepResult := n.Prep(shared)

	// Get items from shared state
	items := shared.Get("items")
	if items == nil {
		return n.Post(shared, prepResult, "no_items")
	}

	// Process each item using user-provided function
	var results []interface{}
	switch itemList := items.(type) {
	case []interface{}:
		for _, item := range itemList {
			if n.ProcessFunc != nil {
				processed, err := n.ProcessFunc(item)
				if err != nil {
					fmt.Printf("Error processing item %v: %v\n", item, err)
					continue
				}
				results = append(results, processed)
			}
		}
	case []int:
		for _, item := range itemList {
			if n.ProcessFunc != nil {
				processed, err := n.ProcessFunc(item)
				if err != nil {
					fmt.Printf("Error processing item %v: %v\n", item, err)
					continue
				}
				results = append(results, processed)
			}
		}
	}

	// Store results back in shared state
	shared.Set("results", results)

	return n.Post(shared, prepResult, "batch_complete")
}

func main() {
	state := goflow.NewSharedState()

	// Add items to process
	state.Set("items", []int{1, 2, 3, 4, 5})

	// User builds batch pattern with custom processing function
	batchNode := NewBatchNode(func(item interface{}) (interface{}, error) {
		if num, ok := item.(int); ok {
			return fmt.Sprintf("processed-%d", num*2), nil
		}
		return nil, fmt.Errorf("invalid item type")
	})

	result := batchNode.Run(state)
	fmt.Printf("Batch result: %s\n", result)

	// Show processed results
	results := state.Get("results")
	fmt.Printf("Processed items: %v\n", results)
}
