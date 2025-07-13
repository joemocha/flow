package goflow

// internalTestNode is used internally by flow orchestration
type internalTestNode interface {
	_run(shared *SharedState) string
}

// Ensure testNode can be used by flow orchestration
func (n *testNode) _run(shared *SharedState) string {
	// Generate warnings if any
	if n.warnings != nil {
		for _, w := range n.warnings {
			n.Warn(w)
		}
	}

	prepResult := n.Prep(shared)

	var execResult string
	if n.execFunc != nil {
		result, err := n.execFunc(shared)
		if err != nil {
			panic(err)
		}
		execResult = result
	} else {
		execResult = "default"
	}

	return n.Post(shared, prepResult, execResult)
}

// Additional methods to make test nodes work with BaseNode
func (n *instrumentedNode) Run(shared *SharedState) string {
	if len(n.successors) > 0 {
		n.Warn("Node won't run successors. Use Flow.")
	}
	return n._run(shared)
}

func (n *instrumentedNode) _run(shared *SharedState) string {
	prepResult := n.Prep(shared)
	execResult, err := n.Exec(prepResult)
	if err != nil {
		panic(err)
	}
	return n.Post(shared, prepResult, execResult)
}

func (n *paramNode) Run(shared *SharedState) string {
	return n.BaseNode.Run(shared)
}

func (n *dataFlowNode) Run(shared *SharedState) string {
	return n.BaseNode.Run(shared)
}

func (n *warningNode) Run(shared *SharedState) string {
	return n.BaseNode.Run(shared)
}
