package goflow

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestRetrySuccess tests successful execution without retries
func TestRetrySuccess(t *testing.T) {
	shared := NewSharedState()

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 3,
		execFunc: func(prep interface{}) (string, error) {
			shared.Set("attempts", shared.GetInt("attempts")+1)
			return "success", nil
		},
	}

	result := node.Run(shared)

	assert.Equal(t, "success", result)
	assert.Equal(t, 1, shared.Get("attempts"))
}

// TestRetryExhaustion tests retry exhaustion and error propagation
func TestRetryExhaustion(t *testing.T) {
	shared := NewSharedState()

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 3,
		execFunc: func(prep interface{}) (string, error) {
			attempt := shared.GetInt("attempts") + 1
			shared.Set("attempts", attempt)
			return "", errors.New("always fails")
		},
	}

	// Should panic or return error based on implementation
	assert.Panics(t, func() {
		node.Run(shared)
	})

	assert.Equal(t, 3, shared.Get("attempts"))
}

// TestFallbackExecution tests fallback mechanism after retry failure
func TestFallbackExecution(t *testing.T) {
	shared := NewSharedState()

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 2,
		execFunc: func(prep interface{}) (string, error) {
			attempt := shared.GetInt("attempts") + 1
			shared.Set("attempts", attempt)
			return "", errors.New("fails")
		},
		fallbackFunc: func(prep interface{}, err error) (string, error) {
			shared.Set("fallback", "executed")
			return "fallback_result", nil
		},
	}

	result := node.Run(shared)

	assert.Equal(t, "fallback_result", result)
	assert.Equal(t, 2, shared.Get("attempts"))
	assert.Equal(t, "executed", shared.Get("fallback"))
}

// TestRetryWithDelay tests retry with delays between attempts
func TestRetryWithDelay(t *testing.T) {
	shared := NewSharedState()

	start := time.Now()
	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
		execFunc: func(prep interface{}) (string, error) {
			attempt := shared.GetInt("attempts") + 1
			shared.Set("attempts", attempt)
			if attempt < 3 {
				return "", errors.New("retry needed")
			}
			return "success", nil
		},
	}

	result := node.Run(shared)
	elapsed := time.Since(start)

	assert.Equal(t, "success", result)
	assert.Equal(t, 3, shared.Get("attempts"))
	// Should have 2 delays (between attempts 1-2 and 2-3)
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond)
}

// TestNoFallbackError tests error propagation when no fallback is defined
func TestNoFallbackError(t *testing.T) {
	shared := NewSharedState()

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 1,
		execFunc: func(prep interface{}) (string, error) {
			return "", errors.New("error without fallback")
		},
	}

	assert.Panics(t, func() {
		node.Run(shared)
	})
}

// TestRetryCounter tests current retry counter accessibility
func TestRetryCounter(t *testing.T) {
	shared := NewSharedState()
	retryAttempts := []int{}

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 4,
	}

	node.execFunc = func(prep interface{}) (string, error) {
		// Access current retry counter
		retryAttempts = append(retryAttempts, node.CurrentRetry())
		if node.CurrentRetry() < 3 {
			return "", errors.New("retry")
		}
		return "success", nil
	}

	result := node.Run(shared)

	assert.Equal(t, "success", result)
	assert.Equal(t, []int{0, 1, 2, 3}, retryAttempts)
}

// TestRetryWithTransientErrors tests retry with errors that eventually succeed
func TestRetryWithTransientErrors(t *testing.T) {
	shared := NewSharedState()
	errors := []string{}

	node := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 5,
		execFunc: func(prep interface{}) (string, error) {
			attempt := shared.GetInt("attempts") + 1
			shared.Set("attempts", attempt)

			// Fail on attempts 1 and 3
			if attempt == 1 || attempt == 3 {
				err := fmt.Sprintf("transient error %d", attempt)
				errors = append(errors, err)
				return "", fmt.Errorf(err)
			}

			// Succeed on attempt 2, but we'll keep failing to test more
			if attempt == 2 {
				return "", fmt.Errorf("still failing")
			}

			// Finally succeed on attempt 4
			return "success", nil
		},
	}

	result := node.Run(shared)

	assert.Equal(t, "success", result)
	assert.Equal(t, 4, shared.Get("attempts"))
	assert.Contains(t, errors, "transient error 1")
	assert.Contains(t, errors, "transient error 3")
}

// TestRetryInFlow tests retry nodes within a flow
func TestRetryInFlow(t *testing.T) {
	shared := NewSharedState()

	// First node always succeeds
	start := &testNode{
		name: "start",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("start", "ok")
			return "default", nil
		},
	}

	// Retry node that fails twice then succeeds
	retryNode := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 3,
		execFunc: func(prep interface{}) (string, error) {
			attempt := shared.GetInt("retry_attempts") + 1
			shared.Set("retry_attempts", attempt)
			if attempt < 3 {
				return "", errors.New("retry needed")
			}
			return "default", nil
		},
	}

	// Final node
	end := &testNode{
		name: "end",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("end", "ok")
			return "done", nil
		},
	}

	// Build flow
	start.Next(retryNode).Next(end)

	flow := NewFlow().Start(start)
	result := flow.Run(shared)

	assert.Equal(t, "done", result)
	assert.Equal(t, "ok", shared.Get("start"))
	assert.Equal(t, "ok", shared.Get("end"))
	assert.Equal(t, 3, shared.Get("retry_attempts"))
}

// TestFallbackIntegration tests fallback with flow integration
func TestFallbackIntegration(t *testing.T) {
	shared := NewSharedState()

	// Node that always fails but has fallback
	nodeWithFallback := &RetryNode{
		BaseNode:   BaseNode{},
		MaxRetries: 2,
		execFunc: func(prep interface{}) (string, error) {
			return "", errors.New("always fails")
		},
		fallbackFunc: func(prep interface{}, err error) (string, error) {
			shared.Set("fallback_error", err.Error())
			return "recovered", nil
		},
	}

	// Recovery node
	recoveryNode := &testNode{
		name: "recovery",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("recovered", true)
			return "default", nil
		},
	}

	// Normal continuation node
	normalNode := &testNode{
		name: "normal",
		execFunc: func(s *SharedState) (string, error) {
			s.Set("normal", true)
			return "done", nil
		},
	}

	// Setup branching based on fallback result
	nodeWithFallback.Next(recoveryNode, "recovered")
	nodeWithFallback.Next(normalNode, "default")

	flow := NewFlow().Start(nodeWithFallback)
	flow.Run(shared)

	assert.Equal(t, "always fails", shared.Get("fallback_error"))
	assert.Equal(t, true, shared.Get("recovered"))
	assert.Nil(t, shared.Get("normal"))
}
