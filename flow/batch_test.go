package goflow

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestBatchNodeExecution tests basic batch node processing
func TestBatchNodeExecution(t *testing.T) {
	shared := NewSharedState()

	// Create batch node that doubles each item
	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return s.Get("items").([]int)
		},
		execFunc: func(item interface{}) (interface{}, error) {
			return item.(int) * 2, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("results", results)
			return "done"
		},
	}

	shared.Set("items", []int{1, 2, 3, 4, 5})
	result := batchNode.Run(shared)

	assert.Equal(t, "done", result)
	expected := []interface{}{2, 4, 6, 8, 10}
	assert.Equal(t, expected, shared.Get("results"))
}

// TestUnevenChunks tests batch processing with uneven chunk sizes
func TestUnevenChunks(t *testing.T) {
	shared := NewSharedState()

	// Chunking node
	chunkNode := &ChunkBatchNode{
		BatchNode: BatchNode{
			BaseNode: BaseNode{},
		},
		ChunkSize: 3,
		prepFunc: func(s *SharedState) interface{} {
			items := s.Get("items").([]int)
			// Convert to chunks
			var chunks [][]int
			for i := 0; i < len(items); i += 3 {
				end := i + 3
				if end > len(items) {
					end = len(items)
				}
				chunks = append(chunks, items[i:end])
			}
			return chunks
		},
		execFunc: func(chunk interface{}) (interface{}, error) {
			items := chunk.([]int)
			sum := 0
			for _, v := range items {
				sum += v
			}
			return sum, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("chunk_sums", results)
			return "done"
		},
	}

	// Test with 10 items, chunk size 3 -> [3, 3, 3, 1]
	shared.Set("items", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	chunkNode.Run(shared)

	expected := []interface{}{6, 15, 24, 10} // [1+2+3, 4+5+6, 7+8+9, 10]
	assert.Equal(t, expected, shared.Get("chunk_sums"))
}

// TestEmptyBatch tests handling of empty batches
func TestEmptyBatch(t *testing.T) {
	shared := NewSharedState()

	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return []interface{}{}
		},
		execFunc: func(item interface{}) (interface{}, error) {
			t.Fatal("Should not execute on empty batch")
			return nil, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("results", results)
			return "empty"
		},
	}

	result := batchNode.Run(shared)

	assert.Equal(t, "empty", result)
	assert.Equal(t, []interface{}{}, shared.Get("results"))
}

// TestSingleItemBatch tests batch with single item
func TestSingleItemBatch(t *testing.T) {
	shared := NewSharedState()

	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return []string{"single"}
		},
		execFunc: func(item interface{}) (interface{}, error) {
			return item.(string) + "_processed", nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("results", results)
			return "done"
		},
	}

	result := batchNode.Run(shared)

	assert.Equal(t, "done", result)
	assert.Equal(t, []interface{}{"single_processed"}, shared.Get("results"))
}

// TestCustomChunkSize tests configurable chunk sizes
func TestCustomChunkSize(t *testing.T) {
	testCases := []struct {
		name      string
		items     []int
		chunkSize int
		expected  [][]int
	}{
		{
			name:      "chunk_size_1",
			items:     []int{1, 2, 3, 4},
			chunkSize: 1,
			expected:  [][]int{{1}, {2}, {3}, {4}},
		},
		{
			name:      "chunk_size_2",
			items:     []int{1, 2, 3, 4, 5},
			chunkSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:      "chunk_size_larger_than_items",
			items:     []int{1, 2, 3},
			chunkSize: 10,
			expected:  [][]int{{1, 2, 3}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shared := NewSharedState()
			shared.Set("items", tc.items)

			chunkNode := &ChunkBatchNode{
				BatchNode: BatchNode{BaseNode: BaseNode{}},
				ChunkSize: tc.chunkSize,
			}

			chunks := chunkNode.CreateChunks(tc.items)
			assert.Equal(t, tc.expected, chunks)
		})
	}
}

// TestBatchFlow tests batch processing within a flow
func TestBatchFlow(t *testing.T) {
	shared := NewSharedState()

	// Prepare node - generates items
	prepNode := &testNode{
		name: "prep",
		execFunc: func(s *SharedState) (string, error) {
			items := make([]int, 100)
			for i := range items {
				items[i] = i + 1
			}
			s.Set("items", items)
			return "default", nil
		},
	}

	// Batch processing node - squares each item
	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return s.Get("items").([]int)
		},
		execFunc: func(item interface{}) (interface{}, error) {
			n := item.(int)
			return n * n, nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("squared", results)
			return "default"
		},
	}

	// Reduce node - sums all squared values
	reduceNode := &testNode{
		name: "reduce",
		execFunc: func(s *SharedState) (string, error) {
			squared := s.Get("squared").([]interface{})
			sum := 0
			for _, v := range squared {
				sum += v.(int)
			}
			s.Set("sum", sum)
			return "done", nil
		},
	}

	// Build flow
	prepNode.Next(batchNode).Next(reduceNode)

	flow := NewFlow().Start(prepNode)
	flow.Run(shared)

	// Sum of squares from 1 to 100: n(n+1)(2n+1)/6
	expectedSum := 100 * 101 * 201 / 6
	assert.Equal(t, expectedSum, shared.Get("sum"))
}

// TestBatchNodeError tests error handling in batch processing
func TestBatchNodeError(t *testing.T) {
	shared := NewSharedState()

	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return []int{1, 2, 3, 4, 5}
		},
		execFunc: func(item interface{}) (interface{}, error) {
			n := item.(int)
			if n == 3 {
				return nil, fmt.Errorf("error on item %d", n)
			}
			return n * 2, nil
		},
		stopOnError: true,
	}

	// Should panic or handle error based on implementation
	assert.Panics(t, func() {
		batchNode.Run(shared)
	})
}

// TestBatchNodeContinueOnError tests continuing batch processing despite errors
func TestBatchNodeContinueOnError(t *testing.T) {
	shared := NewSharedState()

	batchNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return []int{1, 2, 3, 4, 5}
		},
		execFunc: func(item interface{}) (interface{}, error) {
			n := item.(int)
			if n == 3 {
				return nil, fmt.Errorf("error on item %d", n)
			}
			return n * 2, nil
		},
		stopOnError: false,
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			// Filter out nil results from errors
			var validResults []interface{}
			for _, r := range results {
				if r != nil {
					validResults = append(validResults, r)
				}
			}
			s.Set("results", validResults)
			return "done"
		},
	}

	result := batchNode.Run(shared)

	assert.Equal(t, "done", result)
	expected := []interface{}{2, 4, 8, 10} // 3 is skipped due to error
	assert.Equal(t, expected, shared.Get("results"))
}

// TestMapReducePattern tests map-reduce pattern with batch nodes
func TestMapReducePattern(t *testing.T) {
	shared := NewSharedState()

	// Map node - transforms strings to lengths
	mapNode := &BatchNode{
		BaseNode: BaseNode{},
		prepFunc: func(s *SharedState) interface{} {
			return []string{"hello", "world", "test", "batch", "processing"}
		},
		execFunc: func(item interface{}) (interface{}, error) {
			return len(item.(string)), nil
		},
		postFunc: func(s *SharedState, prep interface{}, results []interface{}) string {
			s.Set("lengths", results)
			return "default"
		},
	}

	// Reduce node - calculates average length
	reduceNode := &testNode{
		name: "reduce",
		execFunc: func(s *SharedState) (string, error) {
			lengths := s.Get("lengths").([]interface{})
			sum := 0
			for _, l := range lengths {
				sum += l.(int)
			}
			avg := float64(sum) / float64(len(lengths))
			s.Set("average_length", avg)
			return "done", nil
		},
	}

	mapNode.Next(reduceNode)

	flow := NewFlow().Start(mapNode)
	flow.Run(shared)

	// Average of [5, 5, 4, 5, 10] = 29/5 = 5.8
	assert.Equal(t, 5.8, shared.Get("average_length"))
}
