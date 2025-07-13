package goflow

// Node interface defines the minimal contract (like PocketFlow)
type Node interface {
	SetParams(params map[string]interface{})
	GetParam(key string) interface{}
	Next(node Node, action string) Node
	Prep(shared *SharedState) interface{}
	Exec(prepResult interface{}) (string, error)
	Post(shared *SharedState, prepResult interface{}, execResult string) string
	Run(shared *SharedState) string
	GetSuccessors() map[string]Node
}

// BaseNode implements core node functionality (like PocketFlow's BaseNode)
type BaseNode struct {
	params     map[string]interface{}
	successors map[string]Node
}

func NewBaseNode() *BaseNode {
	return &BaseNode{
		params:     make(map[string]interface{}),
		successors: make(map[string]Node),
	}
}

func (n *BaseNode) SetParams(params map[string]interface{}) {
	n.params = params
}

func (n *BaseNode) GetParam(key string) interface{} {
	return n.params[key]
}

func (n *BaseNode) Next(node Node, action string) Node {
	if action == "" {
		action = "default"
	}
	n.successors[action] = node
	return node
}

func (n *BaseNode) GetSuccessors() map[string]Node {
	return n.successors
}

func (n *BaseNode) Prep(shared *SharedState) interface{} {
	return nil
}

func (n *BaseNode) Exec(prepResult interface{}) (string, error) {
	return "default", nil
}

func (n *BaseNode) Post(shared *SharedState, prepResult interface{}, execResult string) string {
	return execResult
}

func (n *BaseNode) Run(shared *SharedState) string {
	// Default implementation of the prep -> exec -> post lifecycle
	prepResult := n.Prep(shared)
	execResult, err := n.Exec(prepResult)
	if err != nil {
		panic(err) // Match Python behavior
	}
	return n.Post(shared, prepResult, execResult)
}
