package goflow

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestNodeExecution tests basic node execution lifecycle
func TestNodeExecution(t *testing.T) {
	shared := NewSharedState()

	prepCalled := false
	execCalled := false
	postCalled := false

	node := &instrumentedNode{
		onPrep: func(s *SharedState) interface{} {
			prepCalled = true
			s.Set("prep", "completed")
			return "prep_result"
		},
		onExec: func(prep interface{}) (string, error) {
			execCalled = true
			assert.Equal(t, "prep_result", prep)
			return "exec_result", nil
		},
		onPost: func(s *SharedState, prep interface{}, exec string) string {
			postCalled = true
			assert.Equal(t, "prep_result", prep)
			assert.Equal(t, "exec_result", exec)
			s.Set("post", "completed")
			return "final_result"
		},
	}

	result := node.Run(shared)

	assert.True(t, prepCalled, "Prep should be called")
	assert.True(t, execCalled, "Exec should be called")
	assert.True(t, postCalled, "Post should be called")
	assert.Equal(t, "final_result", result)
	assert.Equal(t, "completed", shared.Get("prep"))
	assert.Equal(t, "completed", shared.Get("post"))
}

// TestNodeParams tests parameter passing to nodes
func TestNodeParams(t *testing.T) {
	node := &paramNode{}

	params := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	node.SetParams(params)

	assert.Equal(t, "value1", node.GetParam("key1"))
	assert.Equal(t, 42, node.GetParam("key2"))
	assert.Equal(t, true, node.GetParam("key3"))
	assert.Nil(t, node.GetParam("nonexistent"))
}

// TestPrepExecPost tests the prep-exec-post lifecycle with data flow
func TestPrepExecPost(t *testing.T) {
	shared := NewSharedState()
	shared.Set("initial", 10)

	node := &dataFlowNode{}
	result := node.Run(shared)

	// Verify data transformation through lifecycle
	assert.Equal(t, "processed", result)
	assert.Equal(t, 30, shared.Get("final")) // 10 * 2 + 10
}

// TestMethodChaining tests fluent API for node connections
func TestMethodChaining(t *testing.T) {
	node1 := &BaseNode{}
	node2 := &BaseNode{}
	node3 := &BaseNode{}
	node4 := &BaseNode{}

	// Test Next method chaining
	result := node1.Next(node2).Next(node3)
	assert.Equal(t, node3, result)
	assert.Equal(t, node2, node1.GetSuccessor("default"))
	assert.Equal(t, node3, node2.GetSuccessor("default"))

	// Test Then method (alias for Next)
	node1.Then(node4)
	assert.Equal(t, node4, node1.GetSuccessor("default"))

	// Test conditional transitions
	node1.Next(node2, "action1")
	node1.Next(node3, "action2")
	assert.Equal(t, node2, node1.GetSuccessor("action1"))
	assert.Equal(t, node3, node1.GetSuccessor("action2"))
}

// TestNodeSuccessors tests successor management
func TestNodeSuccessors(t *testing.T) {
	node := &BaseNode{}
	successor1 := &BaseNode{}
	successor2 := &BaseNode{}

	// Add successors
	node.Next(successor1, "path1")
	node.Next(successor2, "path2")

	// Verify successors
	assert.Equal(t, successor1, node.GetSuccessor("path1"))
	assert.Equal(t, successor2, node.GetSuccessor("path2"))
	assert.Nil(t, node.GetSuccessor("nonexistent"))

	// Test successor overwrite warning
	warnings := NewWarningCollector()
	node.WithWarningCollector(warnings)

	newSuccessor := &BaseNode{}
	node.Next(newSuccessor, "path1") // Should warn about overwrite

	assert.Equal(t, newSuccessor, node.GetSuccessor("path1"))
	assert.Contains(t, warnings.Warnings(), "path1")
}

// TestNodeInterface tests that custom nodes properly implement the interface
func TestNodeInterface(t *testing.T) {
	// Ensure various node types implement Node interface
	var _ Node = &BaseNode{}
	var _ Node = &instrumentedNode{}
	var _ Node = &paramNode{}
	var _ Node = &dataFlowNode{}
}

// TestNodeWarnings tests warning generation in nodes
func TestNodeWarnings(t *testing.T) {
	shared := NewSharedState()
	warnings := NewWarningCollector()

	node := &warningNode{
		warnings: []string{
			"Warning 1",
			"Warning 2",
		},
	}
	node.WithWarningCollector(warnings)

	// Add successors to trigger successor warning
	node.Next(&BaseNode{})
	node.Run(shared)

	allWarnings := warnings.Warnings()
	assert.Contains(t, allWarnings, "Warning 1")
	assert.Contains(t, allWarnings, "Warning 2")
	assert.Contains(t, allWarnings, "successors") // Warning about not running successors
}

// instrumentedNode allows tracking of lifecycle method calls
type instrumentedNode struct {
	BaseNode
	onPrep func(*SharedState) interface{}
	onExec func(interface{}) (string, error)
	onPost func(*SharedState, interface{}, string) string
}

func (n *instrumentedNode) Prep(shared *SharedState) interface{} {
	if n.onPrep != nil {
		return n.onPrep(shared)
	}
	return nil
}

func (n *instrumentedNode) Exec(prep interface{}) (string, error) {
	if n.onExec != nil {
		return n.onExec(prep)
	}
	return "default", nil
}

func (n *instrumentedNode) Post(shared *SharedState, prep interface{}, exec string) string {
	if n.onPost != nil {
		return n.onPost(shared, prep, exec)
	}
	return exec
}

// paramNode tests parameter functionality
type paramNode struct {
	BaseNode
}

// dataFlowNode demonstrates data transformation through lifecycle
type dataFlowNode struct {
	BaseNode
}

func (n *dataFlowNode) Prep(shared *SharedState) interface{} {
	initial := shared.Get("initial").(int)
	return initial * 2
}

func (n *dataFlowNode) Exec(prep interface{}) (string, error) {
	doubled := prep.(int)
	result := doubled + 10
	return fmt.Sprintf("%d", result), nil
}

func (n *dataFlowNode) Post(shared *SharedState, prep interface{}, exec string) string {
	// Convert string back to int
	var result int
	fmt.Sscanf(exec, "%d", &result)
	shared.Set("final", result)
	return "processed"
}

// warningNode generates warnings during execution
type warningNode struct {
	BaseNode
	warnings []string
}

func (n *warningNode) Exec(prep interface{}) (string, error) {
	for _, w := range n.warnings {
		n.Warn(w)
	}
	return "done", nil
}
