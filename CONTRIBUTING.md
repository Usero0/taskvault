# Contributing to TaskVault

We welcome contributions! TaskVault is MIT-licensed open-source, and we believe in community-driven development.

## Code of Conduct

- Be respectful and inclusive
- Assume good intent
- Focus on the code, not the person
- Report violations to maintainers

## Getting Started

### Fork & Clone
```bash
git clone https://github.com/YOUR_USERNAME/taskvault.git
cd taskvault
```

### Set Up Development Environment
```bash
# Requires Go 1.21+
go version

# Download dependencies
go mod download

# Run tests
go test ./...

# Build CLI
go build -o taskvault ./cmd/taskvault

# Create your local config
./taskvault init
```

## Development Workflow

### 1. Create Feature Branch
```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bugfix
```

### 2. Make Changes

**Code Standards**:
- Follow Go idioms ([Effective Go](https://golang.org/doc/effective_go))
- Run `go fmt` before commit
- Add tests for new functionality
- Keep functions small and focused
- Use meaningful variable names

**Example**:
```go
// ✓ Good
func (m *Manager) InvalidateTask(taskName string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Clear cached entries for this task
    entries, err := m.store.FindByTask(taskName)
    if err != nil {
        return fmt.Errorf("cannot find entries: %w", err)
    }
    
    for _, entry := range entries {
        if err := m.store.Delete(entry.Hash); err != nil {
            // Log but continue
            m.auditLog.LogError("delete_error", taskName, err)
        }
    }
    
    return nil
}

// ✗ Avoid
func ClearCache(t string) {
    // Unclear what this does
    db.Exec("DELETE FROM cache WHERE task = ?", t)
}
```

### 3. Write Tests

All new features must include tests:
```bash
go test ./... -v -race -coverprofile=coverage.out
```

**Test Template**:
```go
package cache

import "testing"

func TestManagerSaveAndGet(t *testing.T) {
    manager, err := NewManager(".test_cache", 1, Blake3)
    if err != nil {
        t.Fatalf("setup failed: %v", err)
    }
    defer manager.Close()
    
    input := []byte("test input")
    output := []byte("test output")
    
    // Save
    _, err = manager.SaveResult("test_task", input, output, nil)
    if err != nil {
        t.Fatalf("save failed: %v", err)
    }
    
    // Get
    result, _, hit, err := manager.GetResult("test_task", input)
    if err != nil {
        t.Fatalf("get failed: %v", err)
    }
    
    if !hit {
        t.Fatal("expected cache hit")
    }
    
    if string(result) != string(output) {
        t.Errorf("expected %s, got %s", output, result)
    }
}
```

### 4. Commit & Push
```bash
git add .
git commit -m "feat: add task invalidation

- Implements InvalidateTask(taskName)
- Clears all cached entries for specified task
- Includes full audit trail
- Tested with concurrent access patterns"

git push origin feature/my-feature
```

**Commit Message Format**:
```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

### 5. Create Pull Request

Provide:
- **Title**: Clear description of change
- **Description**: Why this change? What problem does it solve?
- **Testing**: How did you test this?
- **Checklist**:
  - [ ] Tests pass (`go test ./...`)
  - [ ] Code formatted (`go fmt ./...`)
  - [ ] No race conditions (`go test -race ./...`)
  - [ ] Documentation updated

## Areas for Contribution

### High Priority
- **PostgreSQL backend** for distributed deployments
- **Python SDK** for ML/data engineering workflows
- **REST API** for remote cache access
- **Metrics export** (Prometheus format)

### Medium Priority
- Cloud storage backends (S3, GCS, Azure)
- Improved eviction strategies (adaptive TTL)
- Cross-platform testing improvements
- Documentation translations

### Low Priority
- Performance optimizations
- UI/Dashboard (future SaaS feature)
- Additional hash algorithms

## Code Review Checklist

Before submitting, verify:
- [ ] **Functionality**: Does it work as intended?
- [ ] **Tests**: Coverage > 80%? Includes edge cases?
- [ ] **Performance**: No regressions? Benchmarked if relevant?
- [ ] **Security**: Input validation? Proper error handling?
- [ ] **Documentation**: Comments where needed? README updated?
- [ ] **Compatibility**: Works on Linux/macOS/Windows?
- [ ] **Dependencies**: Only necessary imports?

## Reporting Issues

Use GitHub Issues with:
- **Title**: Concise description
- **Steps to Reproduce**: Exact commands/code
- **Expected**: What should happen
- **Actual**: What happened
- **Environment**: OS, Go version, TaskVault version

## Questions?

- **GitHub Discussions**: General questions
- **GitHub Issues**: Bug reports
- **Pull Requests**: Feature proposals

---

**Thank you for contributing to TaskVault! 🙏**
