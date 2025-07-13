package main

import (
	"fmt"
	goflow "github.com/sam/goflow/flow"
	"math/rand"
	"time"
)

// RetryNode demonstrates user-built retry pattern (not framework built-in)
type RetryNode struct {
	baseNode   *goflow.BaseNode
	MaxRetries int
	RetryDelay time.Duration
}

func NewRetryNode(maxRetries int, delay time.Duration) *RetryNode {
	return &RetryNode{
		baseNode:   goflow.NewBaseNode(),
		MaxRetries: maxRetries,
		RetryDelay: delay,
	}
}

// Delegate BaseNode methods
func (n *RetryNode) SetParams(params map[string]interface{}) { n.baseNode.SetParams(params) }
func (n *RetryNode) GetParam(key string) interface{}         { return n.baseNode.GetParam(key) }
func (n *RetryNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}
func (n *RetryNode) GetSuccessors() map[string]goflow.Node       { return n.baseNode.GetSuccessors() }
func (n *RetryNode) Prep(shared *goflow.SharedState) interface{} { return n.baseNode.Prep(shared) }
func (n *RetryNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}

// User-implemented retry logic
func (n *RetryNode) Exec(prepResult interface{}) (string, error) {
	// Simulate API call that might fail
	if rand.Float32() < 0.7 {
		return "", fmt.Errorf("API temporarily unavailable")
	}
	return "api_success", nil
}

func (n *RetryNode) Run(shared *goflow.SharedState) string {
	prepResult := n.Prep(shared)

	// Retry logic implemented as user pattern
	for attempt := 0; attempt < n.MaxRetries; attempt++ {
		result, err := n.Exec(prepResult)
		if err == nil {
			return n.Post(shared, prepResult, result)
		}

		fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
		if attempt < n.MaxRetries-1 {
			time.Sleep(n.RetryDelay)
		}
	}

	// All retries failed
	return n.Post(shared, prepResult, "retry_failed")
}

func main() {
	state := goflow.NewSharedState()

	// User builds retry pattern, not framework built-in
	retryNode := NewRetryNode(3, time.Millisecond*100)

	result := retryNode.Run(state)
	fmt.Printf("Final result: %s\n", result)
}
