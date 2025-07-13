package goflow

import (
	"context"
	"fmt"
	"time"
)

// Node interface defines the contract for all nodes
type Node interface {
	SetParams(params map[string]interface{})
	GetParam(key string) interface{}
	Next(node Node, actions ...string) Node
	Then(node Node) Node
	GetSuccessor(action string) Node
	Prep(shared *SharedState) interface{}
	Exec(prepResult interface{}) (string, error)
	Post(shared *SharedState, prepResult interface{}, execResult string) string
	Run(shared *SharedState) string
	WithWarningCollector(w *WarningCollector) Node
}

// BaseNode implements the core node functionality
type BaseNode struct {
	params     map[string]interface{}
	successors map[string]Node
	warnings   *WarningCollector
}

// NewBaseNode creates a new BaseNode
func NewBaseNode() *BaseNode {
	return &BaseNode{
		params:     make(map[string]interface{}),
		successors: make(map[string]Node),
	}
}

// SetParams sets the node parameters
func (n *BaseNode) SetParams(params map[string]interface{}) {
	n.params = params
}

// GetParam retrieves a parameter value
func (n *BaseNode) GetParam(key string) interface{} {
	return n.params[key]
}

// Next adds a successor node for the given action
func (n *BaseNode) Next(node Node, actions ...string) Node {
	action := "default"
	if len(actions) > 0 {
		action = actions[0]
	}

	if _, exists := n.successors[action]; exists {
		n.Warn(fmt.Sprintf("Overwriting successor for action '%s'", action))
	}

	n.successors[action] = node
	return node
}

// Then is an alias for Next with default action
func (n *BaseNode) Then(node Node) Node {
	return n.Next(node)
}

// GetSuccessor retrieves the successor for the given action
func (n *BaseNode) GetSuccessor(action string) Node {
	return n.successors[action]
}

// Prep prepares the node for execution
func (n *BaseNode) Prep(shared *SharedState) interface{} {
	return shared
}

// Exec executes the node logic
func (n *BaseNode) Exec(prepResult interface{}) (string, error) {
	return "default", nil
}

// Post processes the execution result
func (n *BaseNode) Post(shared *SharedState, prepResult interface{}, execResult string) string {
	return execResult
}

// Run executes the complete node lifecycle
func (n *BaseNode) Run(shared *SharedState) string {
	if len(n.successors) > 0 {
		n.Warn("Node won't run successors. Use Flow.")
	}
	return n._run(shared)
}

// _run internal run implementation
func (n *BaseNode) _run(shared *SharedState) string {
	prepResult := n.Prep(shared)
	execResult, err := n._exec(prepResult)
	if err != nil {
		panic(err) // Match Python behavior
	}
	return n.Post(shared, prepResult, execResult)
}

// _exec internal exec implementation
func (n *BaseNode) _exec(prepResult interface{}) (string, error) {
	return n.Exec(prepResult)
}

// WithWarningCollector sets the warning collector
func (n *BaseNode) WithWarningCollector(w *WarningCollector) Node {
	n.warnings = w
	return n
}

// Warn adds a warning
func (n *BaseNode) Warn(message string) {
	if n.warnings != nil {
		n.warnings.Add(message)
	}
}

// RetryNode implements retry logic
type RetryNode struct {
	BaseNode
	MaxRetries   int
	RetryDelay   time.Duration
	execFunc     func(interface{}) (string, error)
	fallbackFunc func(interface{}, error) (string, error)
	currentRetry int
}

// CurrentRetry returns the current retry attempt
func (n *RetryNode) CurrentRetry() int {
	return n.currentRetry
}

// _exec implements retry logic
func (n *RetryNode) _exec(prepResult interface{}) (string, error) {
	var lastErr error

	for n.currentRetry = 0; n.currentRetry < n.MaxRetries; n.currentRetry++ {
		result, err := n.execFunc(prepResult)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if n.currentRetry == n.MaxRetries-1 {
			if n.fallbackFunc != nil {
				return n.fallbackFunc(prepResult, err)
			}
			return "", err
		}

		if n.RetryDelay > 0 {
			time.Sleep(n.RetryDelay)
		}
	}

	return "", lastErr
}

// BatchNode processes items in batches
type BatchNode struct {
	BaseNode
	prepFunc    func(*SharedState) interface{}
	execFunc    func(interface{}) (interface{}, error)
	postFunc    func(*SharedState, interface{}, []interface{}) string
	stopOnError bool
}

// Run executes the batch node
func (n *BatchNode) Run(shared *SharedState) string {
	var prep interface{} = shared
	if n.prepFunc != nil {
		prep = n.prepFunc(shared)
	}

	var items []interface{}
	switch v := prep.(type) {
	case []int:
		items = make([]interface{}, len(v))
		for i, item := range v {
			items[i] = item
		}
	case []string:
		items = make([]interface{}, len(v))
		for i, item := range v {
			items[i] = item
		}
	case []interface{}:
		items = v
	default:
		// Handle empty or nil
		items = []interface{}{}
	}

	results := make([]interface{}, len(items))
	for i, item := range items {
		result, err := n.execFunc(item)
		if err != nil {
			if n.stopOnError {
				panic(err)
			}
			results[i] = nil
		} else {
			results[i] = result
		}
	}

	if n.postFunc != nil {
		return n.postFunc(shared, prep, results)
	}
	return "done"
}

// ChunkBatchNode processes items in chunks
type ChunkBatchNode struct {
	BatchNode
	ChunkSize int
	prepFunc  func(*SharedState) interface{}
	execFunc  func(interface{}) (interface{}, error)
	postFunc  func(*SharedState, interface{}, []interface{}) string
}

// CreateChunks creates chunks from items
func (n *ChunkBatchNode) CreateChunks(items []int) [][]int {
	if n.ChunkSize <= 0 {
		n.ChunkSize = 1
	}

	var chunks [][]int
	for i := 0; i < len(items); i += n.ChunkSize {
		end := i + n.ChunkSize
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}

	return chunks
}

// Run executes the chunk batch node
func (n *ChunkBatchNode) Run(shared *SharedState) string {
	// Use the prepFunc and execFunc from ChunkBatchNode if available
	if n.prepFunc != nil {
		n.BatchNode.prepFunc = n.prepFunc
	}
	if n.execFunc != nil {
		n.BatchNode.execFunc = n.execFunc
	}
	if n.postFunc != nil {
		n.BatchNode.postFunc = n.postFunc
	}
	return n.BatchNode.Run(shared)
}

// AsyncNode implements async node functionality
type AsyncNode struct {
	BaseNode
	prepAsyncFunc func(context.Context, *SharedState) (interface{}, error)
	execAsyncFunc func(context.Context, interface{}) (string, error)
	postAsyncFunc func(context.Context, *SharedState, interface{}, string) (string, error)
}

// RunAsync executes the async node
func (n *AsyncNode) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	if len(n.successors) > 0 {
		n.Warn("Node won't run successors. Use AsyncFlow.")
	}

	var prep interface{}
	var err error

	if n.prepAsyncFunc != nil {
		prep, err = n.prepAsyncFunc(ctx, shared)
		if err != nil {
			return "", err
		}
	} else {
		prep = shared
	}

	var exec string
	if n.execAsyncFunc != nil {
		exec, err = n.execAsyncFunc(ctx, prep)
		if err != nil {
			return "", err
		}
	}

	if n.postAsyncFunc != nil {
		return n.postAsyncFunc(ctx, shared, prep, exec)
	}

	return exec, nil
}

// Run panics for async nodes
func (n *AsyncNode) Run(shared *SharedState) string {
	panic("Use RunAsync")
}

// AsyncRetryNode implements async retry logic
type AsyncRetryNode struct {
	AsyncNode
	MaxRetries        int
	RetryDelay        time.Duration
	execAsyncFunc     func(context.Context, interface{}) (string, error)
	fallbackAsyncFunc func(context.Context, interface{}, error) (string, error)
}

// RunAsync executes with retry logic
func (n *AsyncRetryNode) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	prep := shared

	for i := 0; i < n.MaxRetries; i++ {
		result, err := n.execAsyncFunc(ctx, prep)
		if err == nil {
			return result, nil
		}

		if i == n.MaxRetries-1 {
			if n.fallbackAsyncFunc != nil {
				return n.fallbackAsyncFunc(ctx, prep, err)
			}
			return "", err
		}

		if n.RetryDelay > 0 {
			select {
			case <-time.After(n.RetryDelay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("max retries exceeded")
}

// AsyncBatchNode processes batches asynchronously
type AsyncBatchNode struct {
	BatchNode
}

// ParallelBatchNode processes batches in parallel
type ParallelBatchNode struct {
	AsyncBatchNode
	MaxConcurrency  int
	prepFunc        func(*SharedState) interface{}
	execAsyncFunc   func(context.Context, interface{}) (interface{}, error)
	postFunc        func(*SharedState, interface{}, []interface{}) string
	continueOnError bool
}

// RunAsync executes batch items in parallel
func (n *ParallelBatchNode) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	prep := n.prepFunc(shared)

	var items []interface{}
	switch v := prep.(type) {
	case []int:
		items = make([]interface{}, len(v))
		for i, item := range v {
			items[i] = item
		}
	default:
		panic("unsupported type")
	}

	results := make([]interface{}, len(items))
	errCh := make(chan error, len(items))

	for i, item := range items {
		i, item := i, item // capture loop variables
		go func() {
			result, err := n.execAsyncFunc(ctx, item)
			if err != nil && !n.continueOnError {
				errCh <- err
				return
			}
			results[i] = result
			errCh <- nil
		}()
	}

	// Wait for all goroutines
	for i := 0; i < len(items); i++ {
		if err := <-errCh; err != nil && !n.continueOnError {
			return "", err
		}
	}

	return n.postFunc(shared, prep, results), nil
}
