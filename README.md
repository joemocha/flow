# GoFlow

**Minimalist workflow orchestration library for Go** — A Go port of [PocketFlow](https://github.com/The-Pocket/PocketFlow)

GoFlow provides a powerful node-based execution system for building AI agents, complex workflows, and data processing pipelines with extreme simplicity and type safety.

## Features

- **🪶 Lightweight**: Minimal dependencies, maximum expressiveness
- **🔄 Node-Based Architecture**: Build complex workflows from simple, composable nodes  
- **⚡ Async & Parallel**: Full support for concurrent execution patterns
- **🔁 Retry Logic**: Built-in retry mechanisms with configurable fallback strategies
- **📦 Batch Processing**: Handle collections with chunking and parallel execution
- **🧵 Thread-Safe**: SharedState management for safe data sharing between nodes
- **🎯 Type Safety**: Full Go type system benefits with interface-driven design

## Installation

```bash
go get github.com/sam/goflow
```

## Quick Start

### Basic Node and Flow

```go
package main

import (
    "fmt"
    "github.com/sam/goflow/flow"
)

// Create a custom node
type GreetingNode struct {
    *flow.BaseNode
}

func (n *GreetingNode) Exec(prepResult interface{}) (string, error) {
    name := n.GetParam("name").(string)
    fmt.Printf("Hello, %s!\n", name)
    return "greeted", nil
}

func main() {
    // Create shared state
    state := flow.NewSharedState()
    
    // Create and configure node
    node := &GreetingNode{BaseNode: flow.NewBaseNode()}
    node.SetParams(map[string]interface{}{
        "name": "World",
    })
    
    // Execute
    result := node.Run(state)
    fmt.Printf("Result: %s\n", result)
}
```

### Chaining Nodes with Flow

```go
func main() {
    state := flow.NewSharedState()
    
    // Create nodes
    step1 := &ProcessingNode{BaseNode: flow.NewBaseNode()}
    step2 := &ValidationNode{BaseNode: flow.NewBaseNode()}
    step3 := &OutputNode{BaseNode: flow.NewBaseNode()}
    
    // Chain them together
    step1.Then(step2).Then(step3)
    
    // Create and run flow
    f := flow.NewFlow().Start(step1)
    result := f.Run(state)
}
```

### Conditional Branching

```go
type DecisionNode struct {
    *flow.BaseNode
}

func (n *DecisionNode) Exec(prepResult interface{}) (string, error) {
    value := n.GetParam("value").(int)
    if value > 10 {
        return "high", nil
    }
    return "low", nil
}

func main() {
    decision := &DecisionNode{BaseNode: flow.NewBaseNode()}
    highPath := &HighValueNode{BaseNode: flow.NewBaseNode()}
    lowPath := &LowValueNode{BaseNode: flow.NewBaseNode()}
    
    // Branch based on decision result
    decision.Next(highPath, "high")
    decision.Next(lowPath, "low")
    
    flow.NewFlow().Start(decision).Run(state)
}
```

## Advanced Examples

### Retry Logic

```go
type APICallNode struct {
    *flow.RetryNode
}

func (n *APICallNode) Exec(prepResult interface{}) (string, error) {
    // Simulate API call that might fail
    if rand.Float32() < 0.7 {
        return "", fmt.Errorf("API temporarily unavailable")
    }
    return "success", nil
}

func main() {
    apiNode := &APICallNode{
        RetryNode: &flow.RetryNode{
            BaseNode:   *flow.NewBaseNode(),
            MaxRetries: 3,
            RetryDelay: time.Second,
        },
    }
    
    // Will retry up to 3 times with 1-second delays
    result := apiNode.Run(state)
}
```

### Batch Processing

```go
func main() {
    batch := &flow.BatchNode{
        BaseNode: *flow.NewBaseNode(),
    }
    
    // Configure batch processing
    batch.SetExecFunc(func(item interface{}) (interface{}, error) {
        // Process each item
        return fmt.Sprintf("processed-%v", item), nil
    })
    
    // Add items to shared state
    state.Set("items", []int{1, 2, 3, 4, 5})
    
    result := batch.Run(state)
}
```

### Async Workflows

```go
func main() {
    ctx := context.Background()
    
    asyncNode := &flow.AsyncNode{
        BaseNode: *flow.NewBaseNode(),
    }
    
    // Configure async execution
    asyncNode.SetExecAsyncFunc(func(ctx context.Context, prep interface{}) (string, error) {
        // Async operation
        time.Sleep(100 * time.Millisecond)
        return "async-complete", nil
    })
    
    result, err := asyncNode.RunAsync(ctx, state)
    if err != nil {
        log.Fatal(err)
    }
}
```

## AI Agent Patterns

GoFlow is particularly well-suited for building AI agents:

```go
// Agent reasoning pipeline
func buildReasoningAgent() *flow.Flow {
    // Parse user input
    parser := &InputParserNode{BaseNode: flow.NewBaseNode()}
    
    // Load context and memory
    contextLoader := &ContextLoaderNode{BaseNode: flow.NewBaseNode()}
    
    // LLM reasoning step
    reasoner := &LLMReasoningNode{
        RetryNode: &flow.RetryNode{
            BaseNode:   *flow.NewBaseNode(),
            MaxRetries: 3,
            RetryDelay: time.Second,
        },
    }
    
    // Tool selection and execution
    toolSelector := &ToolSelectorNode{BaseNode: flow.NewBaseNode()}
    toolExecutor := &ToolExecutorNode{BaseNode: flow.NewBaseNode()}
    
    // Response generation
    responseGen := &ResponseGeneratorNode{BaseNode: flow.NewBaseNode()}
    
    // Chain the pipeline
    parser.Then(contextLoader).Then(reasoner)
    
    // Branch based on reasoning result
    reasoner.Next(toolSelector, "needs_tools")
    reasoner.Next(responseGen, "direct_response")
    
    toolSelector.Then(toolExecutor).Then(responseGen)
    
    return flow.NewFlow().Start(parser)
}
```

## Core Types

### Node Interface
```go
type Node interface {
    SetParams(params map[string]interface{})
    GetParam(key string) interface{}
    Next(node Node, actions ...string) Node
    Then(node Node) Node
    Prep(shared *SharedState) interface{}
    Exec(prepResult interface{}) (string, error)
    Post(shared *SharedState, prepResult interface{}, execResult string) string
    Run(shared *SharedState) string
}
```

### Available Node Types
- `BaseNode` - Foundation for all nodes
- `RetryNode` - Automatic retry with configurable strategies
- `BatchNode` - Process collections of items
- `ChunkBatchNode` - Process items in configurable chunks
- `AsyncNode` - Asynchronous execution with context support
- `ParallelBatchNode` - Concurrent batch processing

### SharedState
Thread-safe state container for data sharing between nodes:
```go
state := flow.NewSharedState()
state.Set("key", value)
value := state.Get("key")
state.Append("list", item)
items := state.GetSlice("list")
```

## Python vs Go Implementation

| Feature | Python (100 lines) | Go (650+ lines) |
|---------|-------------------|-----------------|
| **Type Safety** | Runtime | Compile-time |
| **Concurrency** | asyncio | Goroutines + channels |
| **Performance** | Interpreted | Compiled |
| **Error Handling** | Exceptions | Explicit errors |
| **Memory Safety** | GC | GC + compile-time checks |
| **Deployment** | Requires runtime | Single binary |

The Go implementation provides the same conceptual model as PocketFlow but leverages Go's strengths for production deployments.

## Testing

```bash
# Run all tests
go test ./flow/...

# Run with verbose output
go test -v ./flow/...

# Run specific test pattern
go test ./flow/ -run TestFlow

# Run benchmarks
go test -bench=. ./flow/...
```

## Development

```bash
# Build the project
go build ./...

# Format code
gofmt -w .

# Run linter
go vet ./...

# Tidy dependencies
go mod tidy
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -am 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License - see LICENSE file for details.

## Related Projects

- [PocketFlow](https://github.com/The-Pocket/PocketFlow) - Original Python implementation
- [GoFlow Examples](./examples/) - Extended examples and patterns

---

**Built with ❤️ for the Go community**