package goflow

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// SharedState represents the shared state passed between nodes
type SharedState struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSharedState creates a new shared state instance
func NewSharedState() *SharedState {
	return &SharedState{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the shared state
func (s *SharedState) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the shared state
func (s *SharedState) Get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// GetInt retrieves an int value, returning 0 if not found or not an int
func (s *SharedState) GetInt(key string) int {
	val := s.Get(key)
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}

// Append adds an item to a slice in shared state
func (s *SharedState) Append(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.data[key].([]interface{}); ok {
		s.data[key] = append(existing, value)
	} else if existing, ok := s.data[key].([]int); ok {
		if v, ok := value.(int); ok {
			s.data[key] = append(existing, v)
		}
	} else {
		s.data[key] = []interface{}{value}
	}
}

// WarningCollector collects warnings during execution
type WarningCollector struct {
	warnings []string
	mu       sync.Mutex
}

// NewWarningCollector creates a new warning collector
func NewWarningCollector() *WarningCollector {
	return &WarningCollector{
		warnings: []string{},
	}
}

// Add adds a warning to the collector
func (w *WarningCollector) Add(warning string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.warnings = append(w.warnings, warning)
}

// Warnings returns all collected warnings
func (w *WarningCollector) Warnings() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string{}, w.warnings...)
}

// assertWarning checks if a warning contains the expected message
func assertWarning(t *testing.T, warnings *WarningCollector, expectedMsg string) {
	found := false
	for _, w := range warnings.Warnings() {
		if contains(w, expectedMsg) {
			found = true
			break
		}
	}
	assert.True(t, found, fmt.Sprintf("Expected warning containing '%s' not found in %v", expectedMsg, warnings.Warnings()))
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) >= len(substr) && contains(s[1:], substr)
}

// createTestFlow creates a simple test flow
func createTestFlow() *Flow {
	node := &testNode{
		name: "test",
		execFunc: func(s *SharedState) (string, error) {
			return "done", nil
		},
	}
	return NewFlow().Start(node)
}

// waitForCompletion waits for async operations to complete with timeout
func waitForCompletion(ctx context.Context, done chan struct{}, timeout time.Duration) error {
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for completion")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// assertSharedState asserts a value in shared state
func assertSharedState(t *testing.T, state *SharedState, key string, expected interface{}) {
	actual := state.Get(key)
	assert.Equal(t, expected, actual, fmt.Sprintf("SharedState[%s] mismatch", key))
}

// TestNodeBuilder helps build test nodes with common patterns
type TestNodeBuilder struct {
	name     string
	prepFunc func(*SharedState) interface{}
	execFunc func(interface{}) (string, error)
	postFunc func(*SharedState, interface{}, string) string
	warnings []string
}

// NewTestNodeBuilder creates a new test node builder
func NewTestNodeBuilder(name string) *TestNodeBuilder {
	return &TestNodeBuilder{
		name: name,
	}
}

// WithPrep sets the prep function
func (b *TestNodeBuilder) WithPrep(f func(*SharedState) interface{}) *TestNodeBuilder {
	b.prepFunc = f
	return b
}

// WithExec sets the exec function
func (b *TestNodeBuilder) WithExec(f func(interface{}) (string, error)) *TestNodeBuilder {
	b.execFunc = f
	return b
}

// WithPost sets the post function
func (b *TestNodeBuilder) WithPost(f func(*SharedState, interface{}, string) string) *TestNodeBuilder {
	b.postFunc = f
	return b
}

// WithWarnings adds warnings to generate
func (b *TestNodeBuilder) WithWarnings(warnings ...string) *TestNodeBuilder {
	b.warnings = warnings
	return b
}

// Build creates the test node
func (b *TestNodeBuilder) Build() *testNode {
	node := &testNode{
		name:     b.name,
		warnings: b.warnings,
	}

	if b.prepFunc != nil {
		node.prepFunc = b.prepFunc
	}
	if b.execFunc != nil {
		node.execFunc = func(s *SharedState) (string, error) {
			var prep interface{} = s
			if node.prepFunc != nil {
				prep = node.prepFunc(s)
			}
			return b.execFunc(prep)
		}
	}
	if b.postFunc != nil {
		node.postFunc = b.postFunc
	}

	return node
}

// FlowTestHarness provides utilities for testing flows
type FlowTestHarness struct {
	t        *testing.T
	shared   *SharedState
	flow     *Flow
	warnings *WarningCollector
}

// NewFlowTestHarness creates a new test harness
func NewFlowTestHarness(t *testing.T) *FlowTestHarness {
	return &FlowTestHarness{
		t:        t,
		shared:   NewSharedState(),
		warnings: NewWarningCollector(),
	}
}

// WithFlow sets the flow to test
func (h *FlowTestHarness) WithFlow(flow *Flow) *FlowTestHarness {
	h.flow = flow.WithWarningCollector(h.warnings)
	return h
}

// WithSharedState sets initial shared state
func (h *FlowTestHarness) WithSharedState(key string, value interface{}) *FlowTestHarness {
	h.shared.Set(key, value)
	return h
}

// Run executes the flow and returns the result
func (h *FlowTestHarness) Run() string {
	return h.flow.Run(h.shared)
}

// AssertState checks a value in shared state
func (h *FlowTestHarness) AssertState(key string, expected interface{}) *FlowTestHarness {
	assertSharedState(h.t, h.shared, key, expected)
	return h
}

// AssertWarning checks for a warning
func (h *FlowTestHarness) AssertWarning(expectedMsg string) *FlowTestHarness {
	assertWarning(h.t, h.warnings, expectedMsg)
	return h
}

// AssertNoWarnings checks that no warnings were generated
func (h *FlowTestHarness) AssertNoWarnings() *FlowTestHarness {
	assert.Empty(h.t, h.warnings.Warnings())
	return h
}

// measureExecutionTime measures how long a function takes to execute
func measureExecutionTime(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

// createSequentialTestNodes creates a sequence of test nodes
func createSequentialTestNodes(count int, execDelay time.Duration) []*testNode {
	nodes := make([]*testNode, count)
	for i := 0; i < count; i++ {
		idx := i
		nodes[i] = &testNode{
			name: fmt.Sprintf("node_%d", idx),
			execFunc: func(s *SharedState) (string, error) {
				time.Sleep(execDelay)
				s.Append("executed", idx)
				if idx < count-1 {
					return "default", nil
				}
				return "done", nil
			},
		}
	}

	// Chain nodes
	for i := 0; i < count-1; i++ {
		nodes[i].Next(nodes[i+1])
	}

	return nodes
}

// Mock implementations for testing

// MockAsyncNode provides a mock async node for testing
type MockAsyncNode struct {
	AsyncNode
	PrepCalled bool
	ExecCalled bool
	PostCalled bool
	PrepResult interface{}
	ExecResult string
	ExecError  error
	PostResult string
}

func (m *MockAsyncNode) PrepAsync(ctx context.Context, s *SharedState) (interface{}, error) {
	m.PrepCalled = true
	return m.PrepResult, nil
}

func (m *MockAsyncNode) ExecAsync(ctx context.Context, prep interface{}) (string, error) {
	m.ExecCalled = true
	return m.ExecResult, m.ExecError
}

func (m *MockAsyncNode) PostAsync(ctx context.Context, s *SharedState, prep interface{}, exec string) (string, error) {
	m.PostCalled = true
	return m.PostResult, nil
}
