package goflow

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestAsyncNodeExecution tests basic async node execution
func TestAsyncNodeExecution(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	node := &AsyncNode{
		BaseNode: BaseNode{},
		prepAsyncFunc: func(ctx context.Context, s *SharedState) (interface{}, error) {
			s.Set("prep", "async_prep")
			return "prep_data", nil
		},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			// Simulate async work
			time.Sleep(50 * time.Millisecond)
			return "async_result", nil
		},
		postAsyncFunc: func(ctx context.Context, s *SharedState, prep interface{}, exec string) (string, error) {
			s.Set("post", "async_post")
			s.Set("exec_result", exec)
			return "final", nil
		},
	}

	result, err := node.RunAsync(ctx, shared)
	require.NoError(t, err)

	assert.Equal(t, "final", result)
	assert.Equal(t, "async_prep", shared.Get("prep"))
	assert.Equal(t, "async_post", shared.Get("post"))
	assert.Equal(t, "async_result", shared.Get("exec_result"))
}

// TestAsyncFlowOrchestration tests async flow with mixed sync/async nodes
func TestAsyncFlowOrchestration(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Sync node
	syncNode := &testNode{
		BaseNode: *NewBaseNode(),
		name:     "sync",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("sync", "executed")
			return "default", nil
		},
	}

	// Async node
	asyncNode := &AsyncNode{
		BaseNode: *NewBaseNode(),
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			time.Sleep(30 * time.Millisecond)
			shared.Set("async", "executed")
			return "default", nil
		},
	}

	// Another sync node
	endNode := &testNode{
		BaseNode: *NewBaseNode(),
		name:     "end",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("end", "executed")
			return "done", nil
		},
	}

	// Build flow: sync -> async -> sync
	syncNode.Next(asyncNode).Next(endNode)

	flow := &AsyncFlow{
		Flow: *NewFlow().Start(syncNode),
	}

	result, err := flow.RunAsync(ctx, shared)
	require.NoError(t, err)

	assert.Equal(t, "done", result)
	assert.Equal(t, "executed", shared.Get("sync"))
	assert.Equal(t, "executed", shared.Get("async"))
	assert.Equal(t, "executed", shared.Get("end"))
}

// TestContextCancellation tests proper context cancellation handling
func TestContextCancellation(t *testing.T) {
	shared := NewSharedState()

	ctx, cancel := context.WithCancel(context.Background())

	node := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			select {
			case <-time.After(5 * time.Second):
				return "should_not_complete", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		},
	}

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := node.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Less(t, elapsed, 200*time.Millisecond)
}

// TestAsyncRetry tests retry mechanism in async nodes
func TestAsyncRetry(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	node := &AsyncRetryNode{
		AsyncNode: AsyncNode{
			BaseNode: BaseNode{},
		},
		MaxRetries: 3,
		RetryDelay: 50 * time.Millisecond,
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			attempt := shared.GetInt("attempts") + 1
			shared.Set("attempts", attempt)

			if attempt < 3 {
				return "", errors.New("retry needed")
			}
			return "success", nil
		},
	}

	start := time.Now()
	result, err := node.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, shared.Get("attempts"))
	// Should have 2 delays
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
}

// TestAsyncFallback tests fallback in async context
func TestAsyncFallback(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	node := &AsyncRetryNode{
		AsyncNode: AsyncNode{
			BaseNode: BaseNode{},
		},
		MaxRetries: 2,
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			return "", errors.New("always fails")
		},
		fallbackAsyncFunc: func(ctx context.Context, prep interface{}, err error) (string, error) {
			shared.Set("fallback_executed", true)
			shared.Set("original_error", err.Error())
			return "recovered", nil
		},
	}

	result, err := node.RunAsync(ctx, shared)

	require.NoError(t, err)
	assert.Equal(t, "recovered", result)
	assert.Equal(t, true, shared.Get("fallback_executed"))
	assert.Equal(t, "always fails", shared.Get("original_error"))
}

// TestAsyncBatchFlow tests async batch flow processing
func TestAsyncBatchFlow(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	batchFlow := &AsyncBatchFlow{
		AsyncFlow: AsyncFlow{
			Flow: Flow{},
		},
		prepAsyncFunc: func(ctx context.Context, s *SharedState) ([]map[string]interface{}, error) {
			// Generate batch parameters
			batches := []map[string]interface{}{
				{"id": 1, "value": 10},
				{"id": 2, "value": 20},
				{"id": 3, "value": 30},
			}
			return batches, nil
		},
		postAsyncFunc: func(ctx context.Context, s *SharedState, batches []map[string]interface{}, results interface{}) (string, error) {
			s.Set("batch_count", len(batches))
			return "done", nil
		},
	}

	// Node that processes each batch
	processNode := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			params := prep.(map[string]interface{})
			id := params["id"].(int)
			value := params["value"].(int)

			// Store result for this batch
			results := shared.Get("results").([]interface{})
			results = append(results, map[string]int{
				"id":     id,
				"result": value * 2,
			})
			shared.Set("results", results)

			return "processed", nil
		},
	}

	shared.Set("results", []interface{}{})
	batchFlow.Start(processNode)

	result, err := batchFlow.RunAsync(ctx, shared)

	require.NoError(t, err)
	assert.Equal(t, "done", result)
	assert.Equal(t, 3, shared.Get("batch_count"))

	results := shared.Get("results").([]interface{})
	assert.Len(t, results, 3)
}

// TestAsyncTimeout tests timeout handling in async operations
func TestAsyncTimeout(t *testing.T) {
	shared := NewSharedState()

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	node := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			// This should timeout
			time.Sleep(200 * time.Millisecond)
			return "should_not_complete", nil
		},
	}

	start := time.Now()
	_, err := node.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, elapsed, 150*time.Millisecond)
}

// TestAsyncNodeInSyncFlow tests error when async node is used in sync flow
func TestAsyncNodeInSyncFlow(t *testing.T) {
	shared := NewSharedState()

	asyncNode := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			return "async", nil
		},
	}

	flow := NewFlow().Start(asyncNode)

	// Should panic or error when trying to run async node in sync flow
	assert.Panics(t, func() {
		flow.Run(shared)
	})
}

// TestAsyncChainedOperations tests chained async operations
func TestAsyncChainedOperations(t *testing.T) {
	ctx := context.Background()
	shared := NewSharedState()

	// Chain of async operations that depend on each other
	fetchNode := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			time.Sleep(20 * time.Millisecond)
			shared.Set("data", []int{1, 2, 3, 4, 5})
			return "default", nil
		},
	}

	processNode := &AsyncNode{
		BaseNode: BaseNode{},
		prepAsyncFunc: func(ctx context.Context, s *SharedState) (interface{}, error) {
			data := s.Get("data").([]int)
			return data, nil
		},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			data := prep.([]int)
			sum := 0
			for _, v := range data {
				sum += v
			}
			time.Sleep(20 * time.Millisecond)
			shared.Set("sum", sum)
			return "default", nil
		},
	}

	saveNode := &AsyncNode{
		BaseNode: BaseNode{},
		execAsyncFunc: func(ctx context.Context, prep interface{}) (string, error) {
			sum := shared.Get("sum").(int)
			time.Sleep(20 * time.Millisecond)
			shared.Set("saved", true)
			shared.Set("final_sum", sum)
			return "done", nil
		},
	}

	fetchNode.Next(processNode).Next(saveNode)

	flow := &AsyncFlow{
		Flow: *NewFlow().Start(fetchNode),
	}

	start := time.Now()
	result, err := flow.RunAsync(ctx, shared)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, "done", result)
	assert.Equal(t, 15, shared.Get("final_sum"))
	assert.Equal(t, true, shared.Get("saved"))
	// Should take at least 60ms (3 operations * 20ms each)
	assert.GreaterOrEqual(t, elapsed, 60*time.Millisecond)
}
