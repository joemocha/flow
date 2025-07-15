# Changelog

All notable changes to the Flow project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-07-14

### Added
- **Revolutionary adaptive node system** - Single node type that adapts behavior based on parameters
- **Automatic batch processing** - Set `batch: true` with `data` parameter for collection processing
- **Automatic retry logic** - Set `retries > 0` for exponential backoff retry behavior
- **Automatic parallel execution** - Set `parallel: true` for concurrent batch processing
- **Composable patterns** - Mix retry + batch + parallel in single node declarations
- **Thread-safe SharedState** - Concurrent data sharing between nodes with typed getters
- **Flow orchestration** - Chain nodes with action-based routing for complex workflows
- **Zero boilerplate design** - Parameter-driven behavior instead of inheritance patterns

### Features
- **Core Components**:
  - `Node` - Adaptive node with parameter-driven behavior detection
  - `Flow` - Workflow orchestrator for node chaining and execution
  - `SharedState` - Thread-safe state management with typed accessors

- **Adaptive Behaviors**:
  - Single execution (default)
  - Batch processing with `batch: true`
  - Parallel execution with `parallel: true` and `parallel_limit`
  - Retry logic with `retries` and `retry_delay` parameters
  - Full composability of all patterns

- **Developer Experience**:
  - Comprehensive Go documentation with examples
  - 11 test functions covering all adaptive behaviors
  - Race condition testing for thread safety
  - Performance benchmarks
  - 6 working examples demonstrating all patterns

- **Quality Assurance**:
  - Complete CI/CD pipeline with GitHub Actions
  - Multi-version Go testing (1.21, 1.22, 1.23)
  - Code coverage reporting with Codecov
  - Security scanning with Gosec
  - Comprehensive linting with golangci-lint
  - Automated release process

- **Documentation**:
  - Comprehensive README with usage examples
  - Complete API reference documentation
  - Contributing guidelines for community development
  - CC0 license (Public Domain) for maximum freedom

### Examples
- **basic-greeting** - Simple parameter usage and execution
- **batch-pattern** - Automatic batch processing demonstration
- **retry-pattern** - Retry logic with exponential backoff
- **composed-pattern** - All patterns combined (batch + parallel + retry)
- **workflow-pattern** - Multi-node workflow with conditional branching
- **chatbot** - Interactive chatbot using adaptive nodes

### Performance
- **Ultra-lightweight**: ~440 lines total implementation
- **High performance**: 197M ops/sec for basic operations
- **Efficient batching**: 995K ops/sec for sequential batch processing
- **Scalable parallel**: 17K ops/sec for parallel batch processing with proper resource management

### Architecture
- **Parameter Detection Priority**:
  1. Batch Processing: `batch: true` → process each item in `data`
  2. Retry Logic: `retries > 0` → wrap execution with exponential backoff
  3. Single Execution: Default behavior
- **Thread Safety**: All SharedState operations protected by RWMutex
- **Composability**: All adaptive behaviors can be combined seamlessly
- **Extensibility**: Easy to add new parameter-driven behaviors

[Unreleased]: https://github.com/joemocha/flow/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/joemocha/flow/releases/tag/v1.0.0
