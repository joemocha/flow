// Package Flow provides a revolutionary workflow orchestration library for Go.
//
// Flow uses a single adaptive node that automatically changes behavior based on parameters,
// eliminating boilerplate while enabling unprecedented composability for building AI agents,
// complex workflows, and data processing pipelines.
//
// Key Features:
//   - Single Adaptive Node: One node type that automatically adapts behavior based on parameters
//   - Parameter-Driven: Configure behavior through parameters, not inheritance
//   - Auto-Parallel: Add `parallel: true` to any batch operation for instant concurrency
//   - Auto-Retry: Set `retries > 0` to automatically enable retry logic with exponential backoff
//   - Auto-Batch: Set `batch: true` with `data` to automatically process collections
//   - Composable: Mix retry + batch + parallel in a single node declaration
//   - Thread-Safe: SharedState management for safe concurrent data sharing
//
// Basic Usage:
//
//	state := flow.NewSharedState()
//	node := flow.NewNode()
//	node.SetParams(map[string]interface{}{
//		"name": "World",
//	})
//	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
//		name := node.GetParam("name").(string)
//		fmt.Printf("Hello, %s!\n", name)
//		return "greeted", nil
//	})
//	result := node.Run(state)
package Flow

const (
	// DefaultAction represents the default action when no specific action is provided
	DefaultAction = "default"
)

// Flow orchestrates the execution of connected nodes in a workflow.
// It provides sequential traversal and action-based routing between nodes.
//
// A Flow embeds a Node and maintains a reference to the starting node.
// It executes nodes in sequence, following the connections defined by Next() calls
// based on the action strings returned by each node's execution.
type Flow struct {
	*Node
	startNode *Node
}

// NewFlow creates a new Flow instance.
// The returned Flow can be used to orchestrate node execution by setting
// a start node and calling Run().
//
// Example:
//
//	flow := NewFlow().Start(firstNode)
//	result := flow.Run(sharedState)
func NewFlow() *Flow {
	return &Flow{
		Node: NewNode(),
	}
}

// Start sets the starting node for this flow and returns the Flow for method chaining.
// The starting node will be the first node executed when Run() is called.
//
// Parameters:
//   - node: The Node to start execution from
//
// Returns:
//   - *Flow: The same Flow instance for method chaining
func (f *Flow) Start(node *Node) *Flow {
	f.startNode = node
	return f
}

// StartNode returns the current starting node of this flow.
// Returns nil if no starting node has been set.
func (f *Flow) StartNode() *Node {
	return f.startNode
}

// Run executes the flow starting from the start node (like PocketFlow's _orch)
func (f *Flow) Run(shared *SharedState) string {
	curr := f.startNode
	params := f.params
	var lastAction string

	for curr != nil {
		// Set params on current node
		if params != nil {
			curr.SetParams(params)
		}

		// Execute current node using Run method
		lastAction = curr.Run(shared)

		// Get next node based on the action
		curr = f.getNextNode(curr, lastAction)
	}

	return lastAction
}

// getNextNode gets the next node based on action (like PocketFlow's get_next_node)
func (f *Flow) getNextNode(curr *Node, action string) *Node {
	if action == "" {
		action = DefaultAction
	}

	successors := curr.GetSuccessors()
	if next, exists := successors[action]; exists {
		return next
	}

	// Try default if specific action not found
	if action != DefaultAction {
		if defaultNext, exists := successors[DefaultAction]; exists {
			return defaultNext
		}
	}

	return nil
}
