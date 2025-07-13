# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoFlow is a workflow orchestration library implemented in both Go and Python. It provides a node-based execution system for building complex workflows with support for:
- Sequential execution flows
- Parallel and batch processing 
- Asynchronous execution patterns
- Retry logic with fallback handling
- Conditional branching between nodes

## Architecture

### Core Components

**BaseNode** (`flow/node.go:24-122`): Foundation for all nodes with parameter management, successor chaining, and lifecycle methods (Prep/Exec/Post). Includes warning collection system for debugging.

**Flow** (`flow/flow.go:8-254`): Orchestrates node execution starting from a designated start node. Handles sequential traversal based on node return values and action-based routing.

**Node Types**:
- **RetryNode** (`flow/node.go:124-164`): Implements retry logic with configurable max attempts, delays, and fallback functions
- **BatchNode** (`flow/node.go:166-218`): Processes collections of items sequentially or in chunks
- **AsyncNode** (`flow/node.go:262-301`): Provides async execution with context support
- **ParallelBatchNode** (`flow/node.go:351-400`): Concurrent batch processing with configurable concurrency limits

**SharedState** (`flow/test_helpers.go:13-62`): Thread-safe state container passed between nodes for data sharing with typed getters and collection operations.

### Execution Patterns

1. **Synchronous Flows**: Standard sequential execution through connected nodes
2. **Async Flows**: Context-aware async execution with proper error propagation  
3. **Batch Processing**: Handle collections with optional chunking and parallel execution
4. **Retry Mechanisms**: Configurable retry attempts with exponential backoff support

### Flow Construction

Nodes are chained using:
- `Next(node, action)`: Connect nodes with specific action triggers
- `Then(node)`: Shorthand for default action chaining
- Conditional routing based on node execution results

## Development Commands

### Testing
```bash
go test ./flow/...              # Run all tests
go test -v ./flow/...           # Verbose test output
go test ./flow/ -run TestFlow   # Run specific test pattern
```

### Build and Validation
```bash
go build ./...                  # Build all packages
go mod tidy                     # Clean up dependencies  
go vet ./...                    # Static analysis
```

### Code Quality
```bash
gofmt -w .                      # Format all Go files
go mod verify                   # Verify dependencies
```

## Key Implementation Details

- **Node Lifecycle**: All nodes follow Prep → Exec → Post pattern for consistent execution
- **Error Handling**: Panics are used to match Python behavior; async variants return errors
- **State Management**: SharedState provides thread-safe data sharing between nodes
- **Warning System**: Centralized warning collection for debugging flow execution issues
- **Type Reflection**: Flow orchestrator uses type switching to handle different node implementations

## Testing Utilities

The codebase includes comprehensive testing helpers in `flow/test_helpers.go`:
- **FlowTestHarness**: Simplified flow testing with state assertions
- **TestNodeBuilder**: Fluent interface for creating test nodes
- **Mock implementations**: For async nodes and complex testing scenarios

## Language Implementations

- **Go**: Primary implementation with full feature set (`flow/` directory)
- **Python**: Compact reference implementation with similar API (`main.py`)

Both implementations share the same conceptual model but leverage language-specific idioms and patterns.