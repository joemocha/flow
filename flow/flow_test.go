package goflow

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// newTestNode creates a new test node with initialized BaseNode
func newTestNode(name string, execFunc func(*SharedState) (string, error)) *testNode {
	return &testNode{
		BaseNode: *NewBaseNode(),
		name:     name,
		execFunc: execFunc,
	}
}

// TestFlowInitialization tests basic flow creation and start node setting
func TestFlowInitialization(t *testing.T) {
	flow := NewFlow()
	assert.NotNil(t, flow)
	assert.Nil(t, flow.StartNode())

	node := newTestNode("start", nil)
	flow.Start(node)
	assert.Equal(t, node, flow.StartNode())
}

// TestLinearPipeline tests sequential node execution
func TestLinearPipeline(t *testing.T) {
	shared := NewSharedState()

	// Create nodes
	start := newTestNode("start", func(s *SharedState) (string, error) {
		s.Set("value", 1)
		return "default", nil
	})

	middle := newTestNode("middle", func(s *SharedState) (string, error) {
		val, _ := s.Get("value").(int)
		s.Set("value", val*2)
		return "default", nil
	})

	end := newTestNode("end", func(s *SharedState) (string, error) {
		val, _ := s.Get("value").(int)
		s.Set("value", val+10)
		return "done", nil
	})

	// Build pipeline: start -> middle -> end
	start.Next(middle)
	middle.Next(end)

	flow := NewFlow().Start(start)
	result := flow.Run(shared)

	assert.Equal(t, "done", result)
	assert.Equal(t, 12, shared.Get("value")) // (1 * 2) + 10
}

// TestConditionalBranching tests flow branching based on node execution results
func TestConditionalBranching(t *testing.T) {
	shared := NewSharedState()
	shared.Set("value", 5)

	checkNode := newTestNode("check", func(s *SharedState) (string, error) {
		val := s.Get("value").(int)
		if val > 0 {
			return "positive", nil
		}
		return "negative", nil
	})

	positiveNode := newTestNode("positive", func(s *SharedState) (string, error) {
		s.Set("branch", "positive branch")
		return "done", nil
	})

	negativeNode := newTestNode("negative", func(s *SharedState) (string, error) {
		s.Set("branch", "negative branch")
		return "done", nil
	})

	// Setup branching
	checkNode.Next(positiveNode, "positive")
	checkNode.Next(negativeNode, "negative")

	flow := NewFlow().Start(checkNode)
	flow.Run(shared)

	assert.Equal(t, "positive branch", shared.Get("branch"))

	// Test negative branch
	shared.Set("value", -5)
	flow.Run(shared)
	assert.Equal(t, "negative branch", shared.Get("branch"))
}

// TestDefaultTransitions tests default action routing
func TestDefaultTransitions(t *testing.T) {
	shared := NewSharedState()

	node1 := newTestNode("node1", func(s *SharedState) (string, error) {
		s.Set("visited", "node1")
		return "", nil // Empty action should trigger default
	})

	node2 := newTestNode("node2", func(s *SharedState) (string, error) {
		s.Set("visited", "node2")
		return "done", nil
	})

	node1.Next(node2) // Default transition

	flow := NewFlow().Start(node1)
	flow.Run(shared)

	assert.Equal(t, "node2", shared.Get("visited"))
}

// TestMissingTransitionWarning tests warning generation for missing transitions
func TestMissingTransitionWarning(t *testing.T) {
	shared := NewSharedState()
	warnings := NewWarningCollector()

	node := newTestNode("node", func(s *SharedState) (string, error) {
		return "undefined_action", nil
	})

	// Add a different action to have successors but not the one we return
	dummy := newTestNode("dummy", nil)
	node.Next(dummy, "other_action")

	flow := NewFlow().Start(node).WithWarningCollector(warnings)
	flow.Run(shared)

	assert.Contains(t, warnings.Warnings(), "undefined_action")
}

// TestSharedStateManagement tests shared state isolation and management
func TestSharedStateManagement(t *testing.T) {
	shared1 := NewSharedState()
	shared2 := NewSharedState()

	node := newTestNode("node", func(s *SharedState) (string, error) {
		s.Set("key", "value")
		return "done", nil
	})

	flow := NewFlow().Start(node)

	flow.Run(shared1)
	flow.Run(shared2)

	assert.Equal(t, "value", shared1.Get("key"))
	assert.Equal(t, "value", shared2.Get("key"))

	// Modify one shouldn't affect the other
	shared1.Set("key", "modified")
	assert.Equal(t, "modified", shared1.Get("key"))
	assert.Equal(t, "value", shared2.Get("key"))
}

// TestNodeChaining tests the >> operator equivalent (method chaining)
func TestNodeChaining(t *testing.T) {
	shared := NewSharedState()

	node1 := newTestNode("node1", func(s *SharedState) (string, error) {
		s.Set("path", []string{"node1"})
		return "default", nil
	})

	node2 := newTestNode("node2", func(s *SharedState) (string, error) {
		path := s.Get("path").([]string)
		s.Set("path", append(path, "node2"))
		return "default", nil
	})

	node3 := newTestNode("node3", func(s *SharedState) (string, error) {
		path := s.Get("path").([]string)
		s.Set("path", append(path, "node3"))
		return "done", nil
	})

	// Chain using fluent API
	node1.Then(node2).Then(node3)

	flow := NewFlow().Start(node1)
	flow.Run(shared)

	expected := []string{"node1", "node2", "node3"}
	assert.Equal(t, expected, shared.Get("path"))
}

// TestFlowComposition tests nested flow execution
func TestFlowComposition(t *testing.T) {
	shared := NewSharedState()

	// Inner flow
	innerStart := newTestNode("inner_start", func(s *SharedState) (string, error) {
		s.Set("inner", "started")
		return "default", nil
	})

	innerEnd := newTestNode("inner_end", func(s *SharedState) (string, error) {
		s.Set("inner", "completed")
		return "inner_done", nil
	})

	innerStart.Next(innerEnd)
	innerFlow := &FlowNode{
		BaseNode: *NewBaseNode(),
		Flow:     NewFlow().Start(innerStart),
	}

	// Outer flow
	outerStart := newTestNode("outer_start", func(s *SharedState) (string, error) {
		s.Set("outer", "started")
		return "default", nil
	})

	outerEnd := newTestNode("outer_end", func(s *SharedState) (string, error) {
		s.Set("outer", "completed")
		return "done", nil
	})

	outerStart.Next(innerFlow).Next(outerEnd)

	flow := NewFlow().Start(outerStart)
	flow.Run(shared)

	assert.Equal(t, "completed", shared.Get("inner"))
	assert.Equal(t, "completed", shared.Get("outer"))
}

// testNode is a test implementation of Node interface
type testNode struct {
	BaseNode
	name     string
	execFunc func(*SharedState) (string, error)
	prepFunc func(*SharedState) interface{}
	postFunc func(*SharedState, interface{}, string) string
	warnings []string
}

func (n *testNode) Exec(prep interface{}) (string, error) {
	if n.execFunc != nil {
		return n.execFunc(prep.(*SharedState))
	}
	return "default", nil
}

func (n *testNode) Prep(shared *SharedState) interface{} {
	if n.prepFunc != nil {
		return n.prepFunc(shared)
	}
	return shared
}

func (n *testNode) Post(shared *SharedState, prep interface{}, exec string) string {
	if n.postFunc != nil {
		return n.postFunc(shared, prep, exec)
	}
	return exec
}
