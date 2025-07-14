package Flow

import (
	"fmt"
	"sync"
	"time"
)

// Node is the single adaptive node that changes behavior based on parameters
type Node struct {
	params     map[string]interface{}
	successors map[string]*Node

	// User-provided functions (optional)
	execFunc func(interface{}) (interface{}, error)
	prepFunc func(*SharedState) interface{}
	postFunc func(*SharedState, interface{}, interface{}) string
}

// NewNode creates a new adaptive node
func NewNode() *Node {
	return &Node{
		params:     make(map[string]interface{}),
		successors: make(map[string]*Node),
	}
}

// SetParams sets node parameters
func (n *Node) SetParams(params map[string]interface{}) {
	n.params = params
}

// GetParam gets a parameter value
func (n *Node) GetParam(key string) interface{} {
	return n.params[key]
}

// Next sets the next node for a given action
func (n *Node) Next(node *Node, action string) *Node {
	if action == "" {
		action = "default"
	}
	n.successors[action] = node
	return node
}

// GetSuccessors returns all successors
func (n *Node) GetSuccessors() map[string]*Node {
	return n.successors
}

// SetExecFunc sets the user's business logic function
func (n *Node) SetExecFunc(fn func(interface{}) (interface{}, error)) {
	n.execFunc = fn
}

// SetPrepFunc sets optional preparation function
func (n *Node) SetPrepFunc(fn func(*SharedState) interface{}) {
	n.prepFunc = fn
}

// SetPostFunc sets optional post-processing function
func (n *Node) SetPostFunc(fn func(*SharedState, interface{}, interface{}) string) {
	n.postFunc = fn
}

// Run executes the node with adaptive behavior based on parameters
func (n *Node) Run(shared *SharedState) string {
	// Check for batch processing first
	if n.getBoolParam("batch") {
		if data := n.GetParam("data"); data != nil {
			return n.runBatch(shared, data)
		}
		// If batch: true but no data, fall through to single execution
	}

	// Check for retry behavior
	if retryMax := n.getIntParam("retry_max"); retryMax > 0 {
		return n.runWithRetry(shared, retryMax)
	}

	// Default single execution
	return n.runSingle(shared)
}

// runSingle executes the basic prep -> exec -> post lifecycle
func (n *Node) runSingle(shared *SharedState) string {
	// Prep phase
	var prepResult interface{}
	if n.prepFunc != nil {
		prepResult = n.prepFunc(shared)
	}

	// Exec phase
	var execResult interface{} = "default"
	if n.execFunc != nil {
		result, err := n.execFunc(prepResult)
		if err != nil {
			panic(err) // Match Python behavior
		}
		execResult = result
	}

	// Post phase
	if n.postFunc != nil {
		return n.postFunc(shared, prepResult, execResult)
	}

	// Convert result to string
	if str, ok := execResult.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", execResult)
}

// runWithRetry wraps execution with retry logic when retry_max > 0
func (n *Node) runWithRetry(shared *SharedState, maxRetries int) string {
	retryDelay := n.getDurationParam("retry_delay")

	// Prep phase (once)
	var prepResult interface{}
	if n.prepFunc != nil {
		prepResult = n.prepFunc(shared)
	}

	// Retry loop around exec phase
	var execResult interface{} = "default"
	for attempt := 0; attempt < maxRetries; attempt++ {
		if n.execFunc != nil {
			result, err := n.execFunc(prepResult)
			if err == nil {
				execResult = result
				break
			}

			// Log retry attempt (could be made configurable)
			if attempt < maxRetries-1 && retryDelay > 0 {
				time.Sleep(retryDelay)
			}

			// Last attempt failed
			if attempt == maxRetries-1 {
				panic(err)
			}
		} else {
			execResult = "default"
			break
		}
	}

	// Post phase
	if n.postFunc != nil {
		return n.postFunc(shared, prepResult, execResult)
	}

	// Convert result to string
	if str, ok := execResult.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", execResult)
}

// runBatch processes data by calling exec once per item
func (n *Node) runBatch(shared *SharedState, data interface{}) string {
	// Check for parallel processing
	if n.getBoolParam("parallel") {
		return n.runBatchParallel(shared, data)
	}

	// Sequential batch processing
	return n.runBatchSequential(shared, data)
}

// runBatchSequential processes items one by one
func (n *Node) runBatchSequential(shared *SharedState, data interface{}) string {
	items := n.convertToSlice(data)
	results := make([]interface{}, 0, len(items))
	retryMax := n.getIntParam("retry_max")
	retryDelay := n.getDurationParam("retry_delay")

	for _, item := range items {
		if n.execFunc != nil {
			var result interface{}
			var err error

			// Apply retry logic if configured
			if retryMax > 0 {
				for attempt := 0; attempt < retryMax; attempt++ {
					result, err = n.execFunc(item)
					if err == nil {
						break
					}
					if attempt < retryMax-1 && retryDelay > 0 {
						time.Sleep(retryDelay)
					}
				}
			} else {
				result, err = n.execFunc(item)
			}

			if err != nil {
				panic(err)
			}
			results = append(results, result)
		}
	}

	// Store results in shared state
	shared.Set("batch_results", results)
	return "batch_complete"
}

// runBatchParallel processes items concurrently
func (n *Node) runBatchParallel(shared *SharedState, data interface{}) string {
	items := n.convertToSlice(data)
	parallelLimit := n.getIntParam("parallel_limit")
	if parallelLimit <= 0 {
		parallelLimit = len(items) // No limit
	}
	retryMax := n.getIntParam("retry_max")
	retryDelay := n.getDurationParam("retry_delay")

	results := make([]interface{}, len(items))
	sem := make(chan struct{}, parallelLimit)
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(index int, data interface{}) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if n.execFunc != nil {
				var result interface{}
				var err error

				// Apply retry logic if configured
				if retryMax > 0 {
					for attempt := 0; attempt < retryMax; attempt++ {
						result, err = n.execFunc(data)
						if err == nil {
							break
						}
						if attempt < retryMax-1 && retryDelay > 0 {
							time.Sleep(retryDelay)
						}
					}
				} else {
					result, err = n.execFunc(data)
				}

				if err != nil {
					panic(err)
				}
				results[index] = result
			}
		}(i, item)
	}

	wg.Wait()

	// Store results in shared state
	shared.Set("batch_results", results)
	return "batch_complete"
}

// Helper methods for parameter extraction
func (n *Node) getIntParam(key string) int {
	if val := n.GetParam(key); val != nil {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

func (n *Node) getBoolParam(key string) bool {
	if val := n.GetParam(key); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (n *Node) getDurationParam(key string) time.Duration {
	if val := n.GetParam(key); val != nil {
		if d, ok := val.(time.Duration); ok {
			return d
		}
	}
	return 0
}

// convertToSlice handles different slice types
func (n *Node) convertToSlice(data interface{}) []interface{} {
	switch v := data.(type) {
	case []interface{}:
		return v
	case []int:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	default:
		// Single item, wrap in slice
		return []interface{}{data}
	}
}
