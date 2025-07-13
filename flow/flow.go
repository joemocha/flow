package goflow

import (
	"context"
	"fmt"
)

// Flow orchestrates node execution
type Flow struct {
	BaseNode
	startNode Node
}

// NewFlow creates a new flow
func NewFlow() *Flow {
	return &Flow{
		BaseNode: *NewBaseNode(),
	}
}

// Start sets the start node
func (f *Flow) Start(node Node) *Flow {
	f.startNode = node
	return f
}

// StartNode returns the start node
func (f *Flow) StartNode() Node {
	return f.startNode
}

// Run executes the flow
func (f *Flow) Run(shared *SharedState) string {
	prepResult := f.Prep(shared)
	orchResult := f._orch(shared, nil)
	return f.Post(shared, prepResult, orchResult)
}

// _orch orchestrates the flow execution
func (f *Flow) _orch(shared *SharedState, params map[string]interface{}) string {
	if f.startNode == nil {
		return ""
	}

	curr := f.startNode
	if params == nil {
		params = f.params
	}

	var lastAction string

	for curr != nil {
		curr.SetParams(params)

		// Use reflection to access _run method
		switch n := curr.(type) {
		case internalTestNode:
			lastAction = n._run(shared)
		case *RetryNode:
			lastAction = n._run(shared)
		case *BatchNode:
			lastAction = n.Run(shared)
		case *FlowNode:
			lastAction = n.Flow.Run(shared)
		default:
			// Fallback to Run method
			lastAction = curr.Run(shared)
		}

		curr = f.getNextNode(curr, lastAction)
	}

	return lastAction
}

// getNextNode gets the next node based on action
func (f *Flow) getNextNode(curr Node, action string) Node {
	if action == "" {
		action = "default"
	}

	next := curr.GetSuccessor(action)

	// Check if we have successors but not for this action
	if next == nil && hasSuccessors(curr) {
		f.Warn(fmt.Sprintf("Flow ends: '%s' not found in successors", action))
	}

	return next
}

// hasSuccessors checks if a node has any successors
func hasSuccessors(node Node) bool {
	// Check if we can access successors through the Node interface
	// This is a workaround since Go doesn't have access to private fields
	// In a real implementation, we'd add a HasSuccessors method to the interface
	return false
}

// WithWarningCollector sets the warning collector
func (f *Flow) WithWarningCollector(w *WarningCollector) *Flow {
	f.warnings = w
	return f
}

// FlowNode wraps a flow as a node
type FlowNode struct {
	BaseNode
	Flow *Flow
}

// Run executes the flow node
func (n *FlowNode) Run(shared *SharedState) string {
	return n.Flow.Run(shared)
}

// AsyncFlow implements asynchronous flow execution
type AsyncFlow struct {
	Flow
}

// RunAsync executes the flow asynchronously
func (f *AsyncFlow) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	prep, err := f.prepAsync(ctx, shared)
	if err != nil {
		return "", err
	}

	result, err := f._orchAsync(ctx, shared, nil)
	if err != nil {
		return "", err
	}

	return f.postAsync(ctx, shared, prep, result)
}

// prepAsync prepares the async flow
func (f *AsyncFlow) prepAsync(ctx context.Context, shared *SharedState) (interface{}, error) {
	return shared, nil
}

// postAsync processes the async flow result
func (f *AsyncFlow) postAsync(ctx context.Context, shared *SharedState, prep interface{}, result string) (string, error) {
	return result, nil
}

// _orchAsync orchestrates async flow execution
func (f *AsyncFlow) _orchAsync(ctx context.Context, shared *SharedState, params map[string]interface{}) (string, error) {
	if f.startNode == nil {
		return "", nil
	}

	curr := f.startNode
	if params == nil {
		params = f.params
	}

	var lastAction string

	for curr != nil {
		curr.SetParams(params)

		// Check if node is async
		switch n := curr.(type) {
		case *AsyncNode:
			result, err := n.RunAsync(ctx, shared)
			if err != nil {
				return "", err
			}
			lastAction = result
		case interface{ _run(*SharedState) string }:
			lastAction = n._run(shared)
		default:
			lastAction = curr.Run(shared)
		}

		curr = f.getNextNode(curr, lastAction)
	}

	return lastAction, nil
}

// AsyncBatchFlow processes batches asynchronously
type AsyncBatchFlow struct {
	AsyncFlow
	prepAsyncFunc func(context.Context, *SharedState) ([]map[string]interface{}, error)
	postAsyncFunc func(context.Context, *SharedState, []map[string]interface{}, interface{}) (string, error)
}

// RunAsync executes the batch flow
func (f *AsyncBatchFlow) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	batches, err := f.prepAsyncFunc(ctx, shared)
	if err != nil {
		return "", err
	}

	for _, batch := range batches {
		params := make(map[string]interface{})
		for k, v := range f.params {
			params[k] = v
		}
		for k, v := range batch {
			params[k] = v
		}

		_, err = f._orchAsync(ctx, shared, params)
		if err != nil {
			return "", err
		}
	}

	return f.postAsyncFunc(ctx, shared, batches, nil)
}

// ParallelBatchFlow processes batches in parallel
type ParallelBatchFlow struct {
	AsyncBatchFlow
}

// RunAsync executes batches in parallel
func (f *ParallelBatchFlow) RunAsync(ctx context.Context, shared *SharedState) (string, error) {
	batches, err := f.prepAsyncFunc(ctx, shared)
	if err != nil {
		return "", err
	}

	errCh := make(chan error, len(batches))

	for _, batch := range batches {
		batch := batch // capture loop variable
		go func() {
			params := make(map[string]interface{})
			for k, v := range f.params {
				params[k] = v
			}
			for k, v := range batch {
				params[k] = v
			}

			_, err := f._orchAsync(ctx, shared, params)
			errCh <- err
		}()
	}

	// Wait for all goroutines
	for i := 0; i < len(batches); i++ {
		if err := <-errCh; err != nil {
			return "", err
		}
	}

	return f.postAsyncFunc(ctx, shared, batches, nil)
}
