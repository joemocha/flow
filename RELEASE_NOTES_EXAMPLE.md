# Flow v1.0.0 Release Notes

We're excited to announce the first stable release of Flow! 🎉

## 🚀 What's New

### Revolutionary Adaptive Node System
- **Single Node Type**: One adaptive node that changes behavior based on parameters
- **Zero Boilerplate**: Parameter-driven behavior composition over inheritance
- **Automatic Pattern Detection**: Batch, retry, and parallel execution automatically applied

### Core Features
- **Batch Processing**: Set `batch: true` with `data` parameter
- **Retry Logic**: Set `retries > 0` for exponential backoff
- **Parallel Execution**: Set `parallel: true` for concurrent processing
- **Full Composability**: Mix all patterns in a single node declaration

### Developer Experience
- 🧪 **Comprehensive Test Suite**: 11 test functions covering all behaviors
- 📚 **Complete Documentation**: API reference and examples
- ⚡ **High Performance**: 197M ops/sec for basic operations
- 🛡️ **Thread Safety**: Concurrent-safe SharedState management

## 📦 Installation

```bash
go get github.com/joemocha/flow
```

## 🎯 Quick Start

```go
node := NewNode()
node.SetParams(map[string]interface{}{
    "data":     []string{"item1", "item2", "item3"},
    "batch":    true,
    "parallel": true,
    "retries":  3,
})
node.SetExecFunc(func(item interface{}) (interface{}, error) {
    // Your business logic here
    return processItem(item.(string)), nil
})
result := node.Run(state) // Automatic: batch + parallel + retry
```

## 🔧 Breaking Changes

None - this is the initial stable release!

## 🐛 Bug Fixes

- N/A - initial release

## 📈 Performance

- **Basic execution**: 6.2 ns/op (197M ops/sec)
- **Batch sequential**: 1085 ns/op (995K ops/sec)  
- **Batch parallel**: 65μs/op (17K ops/sec)

## 🙏 Acknowledgments

Thank you to all early testers and contributors who helped make this release possible!

## 📋 Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete details.
