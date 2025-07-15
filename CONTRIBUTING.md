# Contributing to Flow

Thank you for your interest in contributing to Flow! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- A GitHub account

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/flow.git
   cd flow
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/joemocha/flow.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Verify everything works:
   ```bash
   go test ./...
   go build ./...
   ```

## 🔄 Development Workflow

### Making Changes

1. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards
3. Add or update tests as needed
4. Ensure all tests pass:
   ```bash
   go test -v -race ./...
   ```

5. Run linting:
   ```bash
   golangci-lint run
   ```

6. Commit your changes with a clear message:
   ```bash
   git commit -m "feat: add new adaptive behavior for X"
   ```

### Commit Message Format

We follow conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

### Submitting Changes

1. Push your branch:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a Pull Request on GitHub
3. Fill out the PR template completely
4. Wait for review and address feedback

## 📝 Coding Standards

### Go Style Guide

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comprehensive documentation for exported functions
- Keep functions focused and small
- Use interfaces where appropriate

### Documentation

- All exported types, functions, and methods must have Go doc comments
- Include examples in documentation where helpful
- Update README.md if adding new features
- Add examples to the `examples/` directory for new patterns

### Testing

- Write tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Include both positive and negative test cases
- Test concurrent behavior with `-race` flag

## 🏗️ Architecture Guidelines

### Core Principles

Flow follows these architectural principles:

1. **Single Adaptive Node**: Avoid creating new node types; extend the existing Node
2. **Parameter-Driven Behavior**: Use parameters to control behavior, not inheritance
3. **Zero Boilerplate**: Keep the API simple and intuitive
4. **Thread Safety**: All shared state operations must be thread-safe
5. **Composability**: Features should work together seamlessly

### Adding New Features

When adding new adaptive behaviors:

1. Add parameter detection in `Node.Run()`
2. Implement the behavior in a separate method
3. Ensure it composes with existing behaviors
4. Add comprehensive tests
5. Update documentation and examples

## 🧪 Testing Guidelines

### Test Structure

```go
func TestNewFeature(t *testing.T) {
    // Arrange
    state := NewSharedState()
    node := NewNode()
    
    // Act
    result := node.Run(state)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Test Categories

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Concurrency Tests**: Test thread safety with `-race`
- **Example Tests**: Ensure examples compile and run

## 📋 Pull Request Checklist

Before submitting a PR, ensure:

- [ ] Code follows Go style guidelines
- [ ] All tests pass (`go test -v -race ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] Examples are added/updated if needed
- [ ] Commit messages follow conventional format
- [ ] PR description explains the change clearly

## 🐛 Reporting Issues

When reporting bugs:

1. Use the issue template
2. Provide a minimal reproduction case
3. Include Go version and OS information
4. Describe expected vs actual behavior

## 💡 Feature Requests

For new features:

1. Check existing issues first
2. Describe the use case clearly
3. Explain how it fits with Flow's philosophy
4. Consider if it can be implemented as parameters

## 📞 Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing documentation and examples first

## 🎉 Recognition

Contributors will be recognized in:
- CHANGELOG.md for releases
- README.md contributors section
- Release notes for significant contributions

Thank you for contributing to Flow! 🚀
