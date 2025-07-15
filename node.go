package Flow

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"
)

const (
	// BatchCompleteAction represents the action returned when batch processing is complete
	BatchCompleteAction = "batch_complete"
)

// secureRandFloat64 generates a cryptographically secure random float64 between 0 and 1
func secureRandFloat64() float64 {
	// Generate a random number between 0 and 2^53-1 (max safe integer for float64)
	maxVal := big.NewInt(1 << 53)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return 0.05 // Fixed small jitter as fallback
	}
	return float64(n.Int64()) / float64(maxVal.Int64())
}

// Node is the core adaptive node that automatically changes behavior based on parameters.
// This single node type eliminates the need for multiple specialized node types by
// detecting patterns in its parameters and adapting its execution accordingly.
//
// Supported Adaptive Behaviors:
//   - Batch Processing: Set "batch": true with "data" to process collections
//   - Parallel Execution: Set "parallel": true to enable concurrent processing
//   - Retry Logic: Set "retries" > 0 to enable automatic retry with exponential backoff
//   - Composability: All patterns can be combined in a single node
//
// Parameter Detection Priority:
//  1. Batch Processing: "batch": true → process each item in "data"
//  2. Retry Logic: "retries" > 0 → wrap execution with exponential backoff retry
//  3. Single Execution: Default behavior
//
// The Node maintains a map of parameters, successor nodes for workflow chaining,
// and optional user-provided functions for custom prep, exec, and post processing.
type Node struct {
	params     map[string]interface{}
	successors map[string]*Node

	// User-provided functions (optional)
	execFunc func(interface{}) (interface{}, error)
	prepFunc func(*SharedState) interface{}
	postFunc func(*SharedState, interface{}, interface{}) string
}

// NewNode creates a new adaptive Node with empty parameters and successors.
// The returned Node can be configured with parameters and functions to define
// its behavior and then executed with Run().
//
// Example:
//
//	node := NewNode()
//	node.SetParams(map[string]interface{}{"retries": 3})
//	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
//		// Your business logic here
//		return "success", nil
//	})
//	result := node.Run(sharedState)
func NewNode() *Node {
	return &Node{
		params:     make(map[string]interface{}),
		successors: make(map[string]*Node),
	}
}

// SetParams configures the node's parameters that control its adaptive behavior.
// Parameters determine which execution patterns the node will use:
//   - "batch": true - enables batch processing of "data" parameter
//   - "parallel": true - enables parallel execution (requires "batch": true)
//   - "parallel_limit": int - limits concurrent goroutines (default: 10)
//   - "retries": int - enables retry logic with exponential backoff
//   - "retry_delay": time.Duration - base delay for retry backoff
//   - "data": []interface{} - data to process in batch mode
//
// Example:
//
//	node.SetParams(map[string]interface{}{
//		"data": []int{1, 2, 3, 4, 5},
//		"batch": true,
//		"parallel": true,
//		"retries": 3,
//	})
func (n *Node) SetParams(params map[string]interface{}) {
	n.params = params
}

// GetParam retrieves a parameter value by key.
// Returns nil if the parameter doesn't exist.
//
// Example:
//
//	retries := node.GetParam("retries")
//	if retries != nil {
//		retriesInt := retries.(int)
//	}
func (n *Node) GetParam(key string) interface{} {
	return n.params[key]
}

// Next establishes a connection to another node for workflow chaining.
// The connection is triggered when this node's execution returns the specified action string.
// If action is empty, "default" is used.
//
// Parameters:
//   - node: The target Node to connect to
//   - action: The action string that triggers this connection
//
// Returns:
//   - *Node: The target node (for method chaining)
//
// Example:
//
//	processor.Next(validator, "processed")
//	validator.Next(success, "valid")
//	validator.Next(failure, "invalid")
func (n *Node) Next(node *Node, action string) *Node {
	if action == "" {
		action = DefaultAction
	}
	n.successors[action] = node
	return node
}

// GetSuccessors returns a map of all successor nodes keyed by their action strings.
// This is primarily used internally by Flow for traversal.
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
	if retries := n.getIntParam("retries"); retries > 0 {
		return n.runWithRetry(shared, retries)
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
	var execResult interface{} = DefaultAction
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

// runWithRetry wraps execution with retry logic when retries > 0
func (n *Node) runWithRetry(shared *SharedState, maxRetries int) string {
	retryDelay := n.getDurationParam("retry_delay")

	// Prep phase (once)
	var prepResult interface{}
	if n.prepFunc != nil {
		prepResult = n.prepFunc(shared)
	}

	// Retry loop around exec phase
	var execResult interface{} = DefaultAction
	for attempt := 0; attempt < maxRetries; attempt++ {
		if n.execFunc != nil {
			result, err := n.execFunc(prepResult)
			if err == nil {
				execResult = result
				break
			}

			// Calculate exponential backoff with jitter for next attempt
			if attempt < maxRetries-1 && retryDelay > 0 {
				// Exponential backoff: retry_delay * (2^attempt) + jitter
				backoffDelay := time.Duration(float64(retryDelay) * math.Pow(2, float64(attempt)))
				// Add jitter (up to 10% of the backoff delay)
				jitter := time.Duration(secureRandFloat64() * float64(backoffDelay) * 0.1)
				totalDelay := backoffDelay + jitter
				time.Sleep(totalDelay)
			}

			// Last attempt failed
			if attempt == maxRetries-1 {
				panic(err)
			}
		} else {
			execResult = DefaultAction
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
	retries := n.getIntParam("retries")
	retryDelay := n.getDurationParam("retry_delay")

	for _, item := range items {
		if n.execFunc == nil {
			continue
		}

		var result interface{}
		var err error

		// Apply retry logic if configured
		if retries > 0 {
			for attempt := 0; attempt < retries; attempt++ {
				result, err = n.execFunc(item)
				if err == nil {
					break
				}
				if attempt < retries-1 && retryDelay > 0 {
					// Exponential backoff: retry_delay * (2^attempt) + jitter
					backoffDelay := time.Duration(float64(retryDelay) * math.Pow(2, float64(attempt)))
					// Add jitter (up to 10% of the backoff delay)
					jitter := time.Duration(secureRandFloat64() * float64(backoffDelay) * 0.1)
					totalDelay := backoffDelay + jitter
					time.Sleep(totalDelay)
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

	// Store results in shared state
	shared.Set("batch_results", results)
	return BatchCompleteAction
}

// runBatchParallel processes items concurrently
func (n *Node) runBatchParallel(shared *SharedState, data interface{}) string {
	items := n.convertToSlice(data)
	parallelLimit := n.getIntParam("parallel_limit")
	if parallelLimit <= 0 {
		parallelLimit = len(items) // No limit
	}
	retries := n.getIntParam("retries")
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
				if retries > 0 {
					for attempt := 0; attempt < retries; attempt++ {
						result, err = n.execFunc(data)
						if err == nil {
							break
						}
						if attempt < retries-1 && retryDelay > 0 {
							// Exponential backoff: retry_delay * (2^attempt) + jitter
							backoffDelay := time.Duration(float64(retryDelay) * math.Pow(2, float64(attempt)))
							// Add jitter (up to 10% of the backoff delay)
							jitter := time.Duration(secureRandFloat64() * float64(backoffDelay) * 0.1)
							totalDelay := backoffDelay + jitter
							time.Sleep(totalDelay)
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
	return BatchCompleteAction
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
