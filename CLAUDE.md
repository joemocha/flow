# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoFlow is a revolutionary workflow orchestration library that evolved from a traditional OOP approach to a single adaptive node system. It represents an advancement of PocketFlow's constraint-based philosophy, providing parameter-driven behavior composition for building AI agents, complex workflows, and data processing pipelines.

**Core Philosophy**: Zero boilerplate, parameter-driven behavior composition over inheritance.

## Revolutionary Architecture

### Single Adaptive Node System

**Node** (`flow/node.go:9-313`): The revolutionary single node type that automatically adapts behavior based on parameters. Eliminates the need for multiple node types through intelligent parameter detection.

```go
type Node struct {
    params     map[string]interface{}
    successors map[string]*Node
    execFunc   func(interface{}) (interface{}, error)
    prepFunc   func(*SharedState) interface{}
    postFunc   func(*SharedState, interface{}, interface{}) string
}
```

**Flow** (`flow/flow.go:3-67`): Simplified orchestrator that works with the unified Node type. Handles sequential traversal and action-based routing.

**SharedState** (`flow/shared_state.go:8-59`): Thread-safe state container for data sharing between nodes with typed getters and collection operations.

### Adaptive Behavior Detection

The single Node automatically detects and applies patterns based on parameters:

1. **Batch Processing**: `batch_data` present → process each item sequentially
2. **Parallel Execution**: `parallel: true` → concurrent goroutine execution with semaphore limits
3. **Retry Logic**: `retry_max > 0` → automatic retry with exponential backoff
4. **Composability**: All patterns can be combined in a single node declaration

### Parameter-Driven Behaviors

| Parameter | Type | Effect | Composition |
|-----------|------|--------|-------------|
| `batch_data` | `[]interface{}` | Auto-batch processing | ✓ Combines with retry + parallel |
| `retry_max` | `int` | Auto-retry logic | ✓ Combines with batch + parallel |
| `retry_delay` | `time.Duration` | Retry delay timing | ✓ Works with retry_max |
| `parallel` | `bool` | Parallel batch execution | ✓ Combines with batch + retry |
| `parallel_limit` | `int` | Goroutine concurrency limit | ✓ Works with parallel |

## Development Commands

### Testing
```bash
go test ./flow/...              # Run all adaptive behavior tests
go test -v ./flow/...           # Verbose test output
go test ./flow/ -run TestAdaptive # Run adaptive-specific tests
go test -race ./flow/...        # Race detection for concurrency
go test -bench=. ./flow/...     # Performance benchmarks
```

### Build and Validation
```bash
go build ./...                  # Build all packages
go mod tidy                     # Clean up dependencies
go vet ./...                    # Static analysis
gofmt -w .                      # Format all Go files
```

## Key Implementation Details

- **Parameter Detection**: Node.Run() method detects behavior patterns through parameter inspection
- **Composable Execution**: Retry, batch, and parallel can be layered automatically
- **Zero Boilerplate**: Users write pure business logic; patterns applied automatically
- **Thread Safety**: Goroutine-based parallel execution with semaphore controls
- **Error Handling**: Panics for flow control; errors returned from business logic
- **State Management**: SharedState provides concurrent-safe data sharing

## Code Examples

### Basic Adaptive Node
```go
node := NewNode()
node.SetParams(map[string]interface{}{"name": "World"})
node.SetExecFunc(func(prep interface{}) (interface{}, error) {
    name := node.GetParam("name").(string)
    return fmt.Sprintf("Hello, %s!", name), nil
})
result := node.Run(state) // Returns: "Hello, World!"
```

### Composed Patterns (Retry + Batch + Parallel)
```go
node := NewNode()
node.SetParams(map[string]interface{}{
    "batch_data":     []string{"url1", "url2", "url3"},
    "parallel":       true,
    "parallel_limit": 2,
    "retry_max":      3,
    "retry_delay":    time.Millisecond * 200,
})
node.SetExecFunc(func(item interface{}) (interface{}, error) {
    // Pure business logic - all patterns applied automatically
    return fetchURL(item.(string)), nil
})
result := node.Run(state) // Automatic: batch + parallel + retry
```

## Comprehensive Test Suite

**Test Coverage** (`flow/node_test.go:1-532`): 11 test functions + 3 benchmarks covering all adaptive behaviors:

- ✅ Basic node execution and parameter handling
- ✅ Automatic retry detection and execution patterns
- ✅ Batch processing (sequential and parallel modes)
- ✅ Composed pattern combinations (retry + batch + parallel)
- ✅ Flow integration with adaptive nodes
- ✅ Edge cases and error handling scenarios
- ✅ Performance benchmarks for all execution modes

**Test Examples**: All working examples converted to test cases for living documentation.

## Performance Characteristics

**Code Reduction**: 60-85% fewer lines compared to traditional OOP approaches
**Memory Efficiency**: Single node type eliminates inheritance overhead
**Execution Speed**:
- Basic: 6.2 ns/op (197M ops/sec)
- Batch Sequential: 1085 ns/op (995K ops/sec)
- Batch Parallel: 65μs/op (17K ops/sec)

## Evolution from PocketFlow

| Aspect | PocketFlow | GoFlow Adaptive |
|--------|------------|-----------------|
| **Core Size** | 100 lines | ~440 lines |
| **Node Types** | 1 BaseNode | 1 Adaptive Node |
| **Patterns** | User-built | Parameter-driven |
| **Composability** | Limited | Unlimited |
| **Boilerplate** | Minimal | Zero |

## Usage Patterns for AI Agents

Perfect for intelligent agent construction:
- **Input Processing**: Retry-enabled parsing with fallback
- **Tool Execution**: Parallel batch processing with retry
- **Response Generation**: Flow chains with adaptive nodes

## Implementation

- **Go**: Revolutionary adaptive implementation with full parameter-driven composability (`flow/` directory)

GoFlow advances PocketFlow's constraint-based philosophy while maintaining its elegant simplicity through parameter-driven behavior composition.
