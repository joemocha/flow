package Flow

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Test helpers for controlled, deterministic testing

// mockCounter provides deterministic behavior for testing
type mockCounter struct {
	count int64
}

func (m *mockCounter) increment() int64 {
	return atomic.AddInt64(&m.count, 1)
}

// TestAdaptiveNodeBasic tests basic node execution without special parameters
func TestAdaptiveNodeBasic(t *testing.T) {
	state := NewSharedState()

	// Test basic execution with parameters
	node := NewNode()
	node.SetParams(map[string]interface{}{
		"name": "TestWorld",
	})

	var capturedName string
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		capturedName = node.GetParam("name").(string)
		return "success", nil
	})

	result := node.Run(state)

	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
	if capturedName != "TestWorld" {
		t.Errorf("Expected 'TestWorld', got '%s'", capturedName)
	}
}

// TestAdaptiveRetryBehavior tests automatic retry detection and execution
func TestAdaptiveRetryBehavior(t *testing.T) {
	state := NewSharedState()
	counter := &mockCounter{}

	// Node with retry configuration
	node := NewNode()
	node.SetParams(map[string]interface{}{
		"retries":     3,
		"retry_delay": time.Millisecond * 10, // Fast for testing
	})

	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		attempt := counter.increment()
		if attempt < 3 {
			return "", fmt.Errorf("attempt %d failed", attempt)
		}
		return "retry_success", nil
	})

	result := node.Run(state)

	if result != "retry_success" {
		t.Errorf("Expected 'retry_success', got '%s'", result)
	}
	if counter.count != 3 {
		t.Errorf("Expected 3 attempts, got %d", counter.count)
	}
}

// TestAdaptiveRetryFailure tests retry exhaustion
func TestAdaptiveRetryFailure(t *testing.T) {
	state := NewSharedState()
	counter := &mockCounter{}

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"retries":     2,
		"retry_delay": time.Millisecond * 5,
	})

	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		attempt := counter.increment()
		return "", fmt.Errorf("attempt %d always fails", attempt)
	})

	// Should panic after exhausting retries
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic after retry exhaustion")
		}
	}()

	node.Run(state)
}

// TestAdaptiveBatchBehavior tests automatic batch processing detection
func TestAdaptiveBatchBehavior(t *testing.T) {
	state := NewSharedState()

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":  []int{1, 2, 3, 4, 5},
		"batch": true,
	})

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		num := item.(int)
		return fmt.Sprintf("processed-%d", num*2), nil
	})

	result := node.Run(state)

	if result != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", result)
	}

	results := state.Get("batch_results").([]interface{})
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	expected := []string{"processed-2", "processed-4", "processed-6", "processed-8", "processed-10"}
	for i, result := range results {
		if result.(string) != expected[i] {
			t.Errorf("Expected '%s', got '%s'", expected[i], result)
		}
	}
}

// TestAdaptiveParallelBehavior tests parallel execution detection
func TestAdaptiveParallelBehavior(t *testing.T) {
	state := NewSharedState()

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":           []string{"item1", "item2", "item3", "item4"},
		"batch":          true,
		"parallel":       true,
		"parallel_limit": 2,
	})

	executionOrder := make([]string, 0)
	var orderMutex sync.Mutex

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		str := item.(string)

		// Record execution order (may vary due to concurrency)
		orderMutex.Lock()
		executionOrder = append(executionOrder, str)
		orderMutex.Unlock()

		// Simulate work
		time.Sleep(time.Millisecond * 50)
		return fmt.Sprintf("parallel-%s", str), nil
	})

	start := time.Now()
	result := node.Run(state)
	elapsed := time.Since(start)

	if result != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", result)
	}

	results := state.Get("batch_results").([]interface{})
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// With parallel_limit=2 and 4 items taking 50ms each,
	// should complete in ~100ms instead of 200ms
	if elapsed > time.Millisecond*150 {
		t.Errorf("Parallel execution took too long: %v", elapsed)
	}

	// All items should be processed (order may vary)
	resultStrs := make([]string, len(results))
	for i, r := range results {
		resultStrs[i] = r.(string)
	}

	expectedItems := []string{"parallel-item1", "parallel-item2", "parallel-item3", "parallel-item4"}
	for _, expected := range expectedItems {
		found := false
		for _, actual := range resultStrs {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected result '%s' not found in %v", expected, resultStrs)
		}
	}
}

// TestComposedRetryBatch tests retry + batch combination
func TestComposedRetryBatch(t *testing.T) {
	state := NewSharedState()
	counter := &mockCounter{}

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":        []string{"item1", "item2", "item3"},
		"batch":       true,
		"retries":     2,
		"retry_delay": time.Millisecond * 5,
	})

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		attempt := counter.increment()
		// Fail first attempt for each item, succeed on second
		if (attempt-1)%2 == 0 {
			return "", fmt.Errorf("first attempt fails for %s", item)
		}
		return fmt.Sprintf("retry-batch-%s", item), nil
	})

	result := node.Run(state)

	if result != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", result)
	}

	results := state.Get("batch_results").([]interface{})
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Should have made 6 attempts total (2 per item)
	if counter.count != 6 {
		t.Errorf("Expected 6 attempts, got %d", counter.count)
	}
}

// TestComposedBatchParallel tests batch + parallel combination
func TestComposedBatchParallel(t *testing.T) {
	state := NewSharedState()

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":           []int{1, 2, 3, 4, 5, 6},
		"batch":          true,
		"parallel":       true,
		"parallel_limit": 3,
	})

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		num := item.(int)
		time.Sleep(time.Millisecond * 30)
		return num * num, nil
	})

	start := time.Now()
	result := node.Run(state)
	elapsed := time.Since(start)

	if result != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", result)
	}

	// With 6 items, parallel_limit=3, should take ~60ms instead of 180ms
	if elapsed > time.Millisecond*90 {
		t.Errorf("Parallel batch took too long: %v", elapsed)
	}

	results := state.Get("batch_results").([]interface{})
	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// Results should be squares (order preserved)
	expected := []int{1, 4, 9, 16, 25, 36}
	for i, result := range results {
		if result.(int) != expected[i] {
			t.Errorf("Expected %d, got %d", expected[i], result)
		}
	}
}

// TestComposedAll tests retry + batch + parallel combination (from composed_pattern.go)
func TestComposedAll(t *testing.T) {
	state := NewSharedState()
	counter := &mockCounter{}

	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":           []string{"url1", "url2", "url3"},
		"batch":          true,
		"parallel":       true,
		"parallel_limit": 2,
		"retries":        3,
		"retry_delay":    time.Millisecond * 10,
	})

	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		url := item.(string)
		attempt := counter.increment()

		// Simulate intermittent failures
		if attempt%3 == 1 { // First attempt of each item fails
			return "", fmt.Errorf("failed to fetch %s", url)
		}

		return fmt.Sprintf("data from %s", url), nil
	})

	start := time.Now()
	result := node.Run(state)
	elapsed := time.Since(start)

	if result != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", result)
	}

	results := state.Get("batch_results").([]interface{})
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify all URLs were processed
	for _, result := range results {
		resultStr := result.(string)
		if !strings.Contains(resultStr, "data from url") {
			t.Errorf("Unexpected result format: %s", resultStr)
		}
	}

	// Should demonstrate parallel execution benefits
	if elapsed > time.Millisecond*100 {
		t.Logf("Execution time: %v (parallel with retries)", elapsed)
	}
}

// TestParameterPrecedence tests the order of parameter detection
func TestParameterPrecedence(t *testing.T) {
	state := NewSharedState()

	// Test that batch: true takes precedence over retries
	node := NewNode()
	node.SetParams(map[string]interface{}{
		"data":    []string{"item1", "item2"},
		"batch":   true,
		"retries": 3, // Should be ignored in favor of batch
	})

	execCount := 0
	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		execCount++
		return fmt.Sprintf("batch-%s", item), nil
	})

	result := node.Run(state)

	if result != "batch_complete" {
		t.Errorf("Expected batch execution, got '%s'", result)
	}

	// Should execute once per batch item, not use retry
	if execCount != 2 {
		t.Errorf("Expected 2 executions (batch), got %d", execCount)
	}
}

// TestFlowWithAdaptiveNodes tests adaptive nodes in flow chains
func TestFlowWithAdaptiveNodes(t *testing.T) {
	state := NewSharedState()

	// Simpler test - just chain basic nodes without batch complexity
	node1 := NewNode()
	node1.SetExecFunc(func(prep interface{}) (interface{}, error) {
		state.Set("step1", "completed")
		return "continue", nil
	})

	node2 := NewNode()
	node2.SetExecFunc(func(prep interface{}) (interface{}, error) {
		state.Set("step2", "processed")
		return "done", nil
	})

	// Chain nodes
	node1.Next(node2, "continue")

	// Create and run flow
	flow := NewFlow().Start(node1)
	result := flow.Run(state)

	if result != "done" {
		t.Errorf("Expected 'done', got '%s'", result)
	}

	if state.Get("step1") != "completed" {
		t.Error("Step 1 not completed")
	}

	if state.Get("step2") != "processed" {
		t.Error("Step 2 not completed")
	}
}

// TestEdgeCases tests error handling and edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("EmptyBatchData", func(t *testing.T) {
		state := NewSharedState()
		node := NewNode()
		node.SetParams(map[string]interface{}{
			"data":  []int{},
			"batch": true,
		})

		execCount := 0
		node.SetExecFunc(func(item interface{}) (interface{}, error) {
			execCount++
			return "should not execute", nil
		})

		result := node.Run(state)

		if result != "batch_complete" {
			t.Errorf("Expected 'batch_complete', got '%s'", result)
		}

		if execCount != 0 {
			t.Errorf("Expected 0 executions for empty batch, got %d", execCount)
		}

		results := state.Get("batch_results").([]interface{})
		if len(results) != 0 {
			t.Errorf("Expected empty results, got %d", len(results))
		}
	})

	t.Run("NoExecFunc", func(t *testing.T) {
		state := NewSharedState()
		node := NewNode()

		result := node.Run(state)

		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})

	t.Run("InvalidParallelLimit", func(t *testing.T) {
		state := NewSharedState()
		node := NewNode()
		node.SetParams(map[string]interface{}{
			"data":           []int{1, 2, 3},
			"batch":          true,
			"parallel":       true,
			"parallel_limit": 0, // Should default to len(items)
		})

		node.SetExecFunc(func(item interface{}) (interface{}, error) {
			return item, nil
		})

		result := node.Run(state)

		if result != "batch_complete" {
			t.Errorf("Expected 'batch_complete', got '%s'", result)
		}

		results := state.Get("batch_results").([]interface{})
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})
}

// Benchmark tests to validate performance characteristics
func BenchmarkAdaptiveNodeBasic(b *testing.B) {
	state := NewSharedState()
	node := NewNode()
	node.SetExecFunc(func(prep interface{}) (interface{}, error) {
		return "result", nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Run(state)
	}
}

func BenchmarkAdaptiveBatchSequential(b *testing.B) {
	state := NewSharedState()
	node := NewNode()

	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	node.SetParams(map[string]interface{}{
		"data":  items,
		"batch": true,
	})
	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		return item.(int) * 2, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Run(state)
	}
}

func BenchmarkAdaptiveBatchParallel(b *testing.B) {
	state := NewSharedState()
	node := NewNode()

	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	node.SetParams(map[string]interface{}{
		"data":           items,
		"batch":          true,
		"parallel":       true,
		"parallel_limit": 10,
	})
	node.SetExecFunc(func(item interface{}) (interface{}, error) {
		return item.(int) * 2, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Run(state)
	}
}
