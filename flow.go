package Flow

// Flow orchestrates node execution (like PocketFlow's Flow)
type Flow struct {
	*Node
	startNode *Node
}

// NewFlow creates a new flow
func NewFlow() *Flow {
	return &Flow{
		Node: NewNode(),
	}
}

// Start sets the starting node
func (f *Flow) Start(node *Node) *Flow {
	f.startNode = node
	return f
}

// StartNode returns the current start node
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
		action = "default"
	}

	successors := curr.GetSuccessors()
	if next, exists := successors[action]; exists {
		return next
	}

	// Try default if specific action not found
	if action != "default" {
		if defaultNext, exists := successors["default"]; exists {
			return defaultNext
		}
	}

	return nil
}