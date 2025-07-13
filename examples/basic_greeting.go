package main

import (
	"fmt"
	goflow "github.com/sam/goflow/flow"
)

type GreetingNode struct {
	baseNode *goflow.BaseNode
}

func NewGreetingNode() *GreetingNode {
	return &GreetingNode{
		baseNode: goflow.NewBaseNode(),
	}
}

// Delegate BaseNode methods
func (n *GreetingNode) SetParams(params map[string]interface{}) {
	n.baseNode.SetParams(params)
}

func (n *GreetingNode) GetParam(key string) interface{} {
	return n.baseNode.GetParam(key)
}

func (n *GreetingNode) Next(node goflow.Node, action string) goflow.Node {
	return n.baseNode.Next(node, action)
}

func (n *GreetingNode) GetSuccessors() map[string]goflow.Node {
	return n.baseNode.GetSuccessors()
}

func (n *GreetingNode) Prep(shared *goflow.SharedState) interface{} {
	return n.baseNode.Prep(shared)
}

func (n *GreetingNode) Post(shared *goflow.SharedState, prepResult interface{}, execResult string) string {
	return n.baseNode.Post(shared, prepResult, execResult)
}

// Custom Exec method
func (n *GreetingNode) Exec(prepResult interface{}) (string, error) {
	name := n.GetParam("name").(string)
	fmt.Printf("Hello, %s!\n", name)
	return "greeted", nil
}

// Override Run to ensure our Exec method gets called
func (n *GreetingNode) Run(shared *goflow.SharedState) string {
	prepResult := n.Prep(shared)
	execResult, err := n.Exec(prepResult) // This will call GreetingNode.Exec()
	if err != nil {
		panic(err)
	}
	return n.Post(shared, prepResult, execResult)
}

func main() {
	state := goflow.NewSharedState()

	node := NewGreetingNode()
	node.SetParams(map[string]interface{}{
		"name": "World",
	})

	result := node.Run(state)
	fmt.Printf("Result: %s\n", result)
}
