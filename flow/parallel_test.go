package goflow

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// TestParallelBatchExecution tests parallel execution of batch items
func TestParallelBatchExecution(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Track execution order to verify parallelism
	var orderMu sync.Mutex
	executionOrder := []int{}

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			return []int{1, 2, 3, 4, 5}
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			n := item.(int)

			// Simulate variable processing time
			delay := time.Duration((6-n)*10) * time.Millisecond
			time.Sleep(delay)

			// Record completion order
			orderMu.Lock()
			executionOrder = append(executionOrder, n)
			orderMu.Unlock()

			return n * n, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("results", results)
			s.Set("order", executionOrder)
			return "done"
		},
	}

	result, err := node.RunAsync(ctx, shared)

	require.NoError(t, err)
	assert.Equal(t, "done", result)

	// Results should be in original order despite parallel execution
	expected := []interface{}{1, 4, 9, 16, 25}
	assert.Equal(t, expected, shared.Get("results"))

	// Execution order should reflect parallel processing (not sequential)
	order := shared.Get("order").([]int)
	assert.NotEqual(t, []int{1, 2, 3, 4, 5}, order)
	// Item 5 should complete first (shortest delay), item 1 last
	assert.Equal(t, 5, order[0])
	assert.Equal(t, 1, order[len(order)-1])
}

// TestParallelTiming tests that parallel execution is faster than sequential
func TestParallelTiming(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	itemCount := 10
	delayPerItem := 50 * time.Millisecond

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			items := make([]int, itemCount)
			for i := range items {
				items[i] = i
			}
			return items
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			time.Sleep(delayPerItem)
			return item.(int) * 2, nil
		},
	}

	start := time.Now()
	_, err := node.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// Parallel execution should complete in roughly delayPerItem time
	// Sequential would take itemCount * delayPerItem
	sequentialTime := time.Duration(itemCount) * delayPerItem
	assert.Less(t, elapsed, sequentialTime/2)
	assert.GreaterOrEqual(t, elapsed, delayPerItem)
}

// TestConcurrencyOrder tests that concurrent execution maintains result order
func TestConcurrencyOrder(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Process 100 items with random delays
	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			items := make([]int, 100)
			for i := range items {
				items[i] = i
			}
			return items
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			n := item.(int)
			// Variable delay based on item value
			delay := time.Duration(n%10) * time.Millisecond
			time.Sleep(delay)
			return n * 10, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("results", results)
			return "done"
		},
	}

	_, err := node.RunAsync(ctx, shared)
	require.NoError(t, err)

	results := shared.Get("results").([]interface{})
	assert.Len(t, results, 100)

	// Verify order is maintained
	for i, r := range results {
		assert.Equal(t, i*10, r)
	}
}

// TestParallelErrorHandling tests error handling in parallel execution
func TestParallelErrorHandling(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	var errorsMu sync.Mutex
	capturedErrors := []error{}

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			return []int{1, 2, 3, 4, 5}
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			n := item.(int)
			if n == 2 || n == 4 {
				err := fmt.Errorf("error on item %d", n)
				errorsMu.Lock()
				capturedErrors = append(capturedErrors, err)
				errorsMu.Unlock()
				return nil, err
			}
			return n * 2, nil
		},
		continueOnError: true,
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			// Filter successful results
			var successful []interface{}
			for _, r := range results {
				if r != nil {
					successful = append(successful, r)
				}
			}
			s.Set("successful", successful)
			s.Set("errors", capturedErrors)
			return "done"
		},
	}

	result, err := node.RunAsync(ctx, shared)

	require.NoError(t, err)
	assert.Equal(t, "done", result)

	successful := shared.Get("successful").([]interface{})
	assert.Equal(t, []interface{}{2, 6, 10}, successful)

	errors := shared.Get("errors").([]error)
	assert.Len(t, errors, 2)
}

// TestParallelBatchFlow tests parallel batch flow processing
func TestParallelBatchFlow(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Track concurrent executions
	var concurrentCount int32
	var maxConcurrent int32
	var mu sync.Mutex

	processNode := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			// Track concurrent executions
			mu.Lock()
			concurrentCount++
			if concurrentCount > maxConcurrent {
				maxConcurrent = concurrentCount
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			concurrentCount--
			mu.Unlock()

			params := prep.(map[string]interface{})
			batchId := params["batch_id"].(int)
			shared.Append("processed", batchId)

			return "done", nil
		},
	}

	flow := &ParallelBatchFlow{
		AsyncBatchFlow: AsyncBatchFlow{
			AsyncFlow: AsyncFlow{
				Flow: *NewFlow().Start(processNode),
			},
		},
	}

	flow.prepAsyncFunc = func(ctx context.Context, s *SharedState) ([]map[string]interface{}, error) {
		// Create 5 batches
		batches := make([]map[string]interface{}, 5)
		for i := range batches {
			batches[i] = map[string]interface{}{
				"batch_id": i,
			}
		}
		return batches, nil
	}

	shared.Set("processed", []int{})

	start := time.Now()
	_, err := flow.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// All batches should be processed
	processed := shared.Get("processed").([]int)
	assert.Len(t, processed, 5)

	// Should have multiple concurrent executions
	assert.Greater(t, int(maxConcurrent), 1)

	// Should be faster than sequential (5 * 50ms = 250ms)
	assert.Less(t, elapsed, 150*time.Millisecond)
}

// TestParallelContextCancellation tests context cancellation during parallel execution
func TestParallelContextCancellation(t *testing.T) {
	shared := NewSharedState()

	ctx, cancel := context.WithCancel(context.Background())

	var completed []int
	var mu sync.Mutex

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			n := item.(int)

			// Different delays for different items
			delay := time.Duration(n*20) * time.Millisecond

			select {
			case <-time.After(delay):
				mu.Lock()
				completed = append(completed, n)
				mu.Unlock()
				return n, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := node.RunAsync(ctx, shared)

	assert.Error(t, err)

	// Only items 1 and 2 should complete (20ms and 40ms delays)
	mu.Lock()
	defer mu.Unlock()
	assert.LessOrEqual(t, len(completed), 3)
	assert.Contains(t, completed, 1)
	assert.Contains(t, completed, 2)
}

// TestParallelResourceLimiting tests limiting concurrent executions
func TestParallelResourceLimiting(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Semaphore to limit concurrency
	maxConcurrency := 3
	sem := make(chan struct{}, maxConcurrency)

	var activeConcurrent int32
	var peakConcurrent int32
	var mu sync.Mutex

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		MaxConcurrency: maxConcurrency,
		prepFunc: func(s *SharedState) interface{} {
			items := make([]int, 10)
			for i := range items {
				items[i] = i
			}
			return items
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Track concurrent executions
			mu.Lock()
			activeConcurrent++
			if activeConcurrent > peakConcurrent {
				peakConcurrent = activeConcurrent
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(20 * time.Millisecond)

			mu.Lock()
			activeConcurrent--
			mu.Unlock()

			return item.(int) * 2, nil
		},
	}

	_, err := node.RunAsync(ctx, shared)
	require.NoError(t, err)

	// Peak concurrent should not exceed limit
	assert.LessOrEqual(t, int(peakConcurrent), maxConcurrency)
}

// TestParallelLargeScale tests parallel processing at scale
func TestParallelLargeScale(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	itemCount := 1000

	node := &ParallelBatchNode{
		AsyncBatchNode: AsyncBatchNode{
			BatchNode: BatchNode{
				BaseNode: BaseNode{},
			},
		},
		prepFunc: func(s *SharedState) interface{} {
			items := make([]int, itemCount)
			for i := range items {
				items[i] = i
			}
			return items
		},
		execAsyncFunc: func(ctx context.Context, item interface{}) (interface{}, error) {
			n := item.(int)
			// Simple computation
			return n*n + n + 1, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			// Verify all results
			sum := 0
			for _, r := range results {
				sum += r.(int)
			}
			s.Set("sum", sum)
			return "done"
		},
	}

	start := time.Now()
	_, err := node.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// Verify correctness: sum of (i*i + i + 1) for i=0 to 999
	// = sum(i^2) + sum(i) + sum(1)
	// = (n-1)*n*(2n-1)/6 + (n-1)*n/2 + n
	n := itemCount
	expectedSum := (n-1)*n*(2*n-1)/6 + (n-1)*n/2 + n
	assert.Equal(t, expectedSum, shared.Get("sum"))

	// Should complete quickly even with 1000 items
	assert.Less(t, elapsed, 1*time.Second)
}
