# HelloGo

Go repository for learning and testing golang projects.

## Project Structure

```
HelloGo/
├── hello/              # Main Go module with examples
│   └── morestrings/    # String utilities package
├── neural_net_iris/   # ML neural network (gonum)
├── README.md
└── AGENTS.md         # Agent instructions
```

## Quick Start

```bash
# Clone and setup
go mod download

# Build
go build ./...

# Run
go run ./hello

# Test
go test -v -race ./...
```

## Best Practices

### Always Clear Cache

Run these before building/testing to ensure clean state:

```bash
# Clear module cache (use after go.mod changes)
go clean -modcache

# Clear build cache
go clean -cache

# Clear both
go clean -modcache -cache
go clean -cache -modcache
```

### Clean Build Process

```bash
# 1. Clear caches
go clean -modcache -cache

# 2. Tidy dependencies
go mod tidy

# 3. Verify module
go mod verify

# 4. Build
go build -v ./...

# 5. Run tests with race detector
go test -race -v ./...
```

## Standard Commands

### Module Management

```bash
# Initialize new module
go mod init github.com/user/project

# Add dependency
go get github.com/user/package

# Update all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy

# Download all modules
go mod download

# Verify module integrity
go mod verify

# Vendor dependencies
go mod vendor
```

### Building

```bash
# Build current module
go build -v ./...

# Build specific package
go build -v ./hello

# Build with version info
go build -ldflags "-X main.Version=1.0.0" ./...

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/server ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 ./...
GOOS=darwin GOARCH=amd64 go build -o bin/app-darwin-amd64 ./...
GOOS=windows GOARCH=amd64 go build -o bin/app-windows-amd64.exe ./...
```

### Running

```bash
# Run main package
go run .

# Run specific file
go run ./hello/main.go

# Run with race detector
go run -race ./...

# Run with environment variables
DEBUG=true go run .
```

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Run specific test
go test -v -run TestReverse ./...

# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory stats
go test -bench=. -benchmem ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests in specific directory
go test -v ./hello/...
go test -v ./neural_net_iris/...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run linter (if installed)
golangci-lint run ./...

# Check for unused imports
goimports -w .
```

### Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.out -bench=.
go tool pprof cpu.out

# Memory profile
go test -memprofile=mem.out -bench=.
go tool pprof mem.out
```

### Cache Management

```bash
# View cache size
go env GOCACHE

# Clear download cache
go clean -modcache

# Clear build cache
go clean -cache

# Clear test cache
go clean -testcache

# Clear everything
go clean -cache -modcache -testcache

# Disable cache (for CI)
GOCACHE=off go build ./...
```

## Development Workflow

### 1. New Project Setup

```bash
# Create directory
mkdir -p newproject/cmd/newproject
cd newproject

# Initialize module
go mod init github.com/user/newproject

# Create main.go in cmd/
go mod tidy
go build -v ./...
```

### 2. Making Changes

```bash
# Make code changes, then:
go clean -cache -modcache
go mod tidy
go vet ./...
go build -v ./...
go test -race -v ./...
```

### 3. Adding Dependencies

```bash
go get github.com/user/package@latest
go mod tidy
go mod verify
go build -v ./...
go test -v ./...
```

### 4. Running in Production

```bash
go clean -cache -modcache
go build -v -ldflags "-s -w" ./...
./binary
```

## Per-Module Commands

### hello/

```bash
cd hello
go build -v ./...
go run .
go test -v ./...
go test -v ./morestrings/...
```

### neural_net_iris/

```bash
cd neural_net_iris
go build -v -o neural_net_iris .
./neural_net_iris
go test -v -race ./...
go test -bench=. -benchmem ./...
```

## Quick Reference

| Action | Command |
|--------|---------|
| Build | `go build -v ./...` |
| Run | `go run .` |
| Test | `go test -race -v ./...` |
| Test (fast) | `go test -race ./...` |
| Benchmarks | `go test -bench=. -benchmem ./...` |
| Lint | `go vet ./...` |
| Format | `go fmt ./...` |
| Tidy | `go mod tidy` |
| Clean cache | `go clean -cache -modcache` |
| Full rebuild | `go clean -cache -modcache && go mod tidy && go build -v ./...` |

## Go Version

Go 1.26.2 required (per go.mod files).

## Notes

- Always use `-race` flag when running tests to detect race conditions
- Run `go clean -cache` after dependency changes
- Use `go mod tidy` to manage dependencies
- Run `go vet ./...` before committing
- Check `go.mod` for required Go version

## Creating Packages (Not package main)

Instead of using `package main` for everything, create reusable packages:

### Basic Package Structure

```bash
mkdir -p myproject/pkg/mypackage
cd myproject
```

```go
// pkg/mypackage/package.go
package mypackage

import "context"

type Service struct {
    name string
}

func NewService(name string) *Service {
    return &Service{name: name}
}

func (s *Service) DoSomething(ctx context.Context) error {
    // logic here
    return nil
}
```

### Using Internal Packages

```go
// internal/ packages are private to the module
myproject/
├── cmd/
│   └── server/
│       └── main.go        # Entry point
├── internal/
│   └── mypackage/       # Private package
│       └── package.go
└── go.mod
```

```go
// internal/mypackage/package.go
package mypackage

// Unexported type (private)
type service struct{}

// Exported constructor
func New() *service {
    return &service{}
}
```

### Creating a Library Package

```go
// mylib/myfunc.go
package mylib

import "errors"

func DoSomething() error {
    if false {
        return errors.New("failed")
    }
    return nil
}
```

```go
// mylib/myfunc_test.go
package mylib

import "testing"

func TestDoSomething(t *testing.T) {
    if err := DoSomething(); err != nil {
        t.Errorf("DoSomething() failed: %v", err)
    }
}
```

### Importing Your Package

```go
// cmd/server/main.go
package main

import (
    "myproject/internal/mypackage"
)

func main() {
    svc := mypackage.New()
    // use service
}
```

### Multi-Package Structure

```
myproject/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── pkg/
│   ├── models/
│   │   └── user.go
│   └── utils/
│       └── helpers.go
├── internal/
│   ├── auth/
│   │   └── jwt.go
│   └── db/
│       └── postgres.go
├── go.mod
└── README.md
```

### Package Naming Conventions

| Package | Purpose |
|---------|--------|
| `cmd/` | Executable entry points |
| `pkg/` | Public libraries (importable externally) |
| `internal/` | Private to this module |
| `internal/auth/` | Authentication logic |
| `internal/db/` | Database layer |
| `internal/api/` | API handlers |
| `internal/service/` | Business logic |

### Example: Creating morestrings Package

```bash
cd hello
mkdir -p morestrings
```

```go
// morestrings/reverse.go
package morestrings

func Reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
```

```go
// morestrings/reverse_test.go
package morestrings

import "testing"

func TestReverse(t *testing.T) {
    tests := []struct {
        input string
        want  string
    }{
        {"hello", "olleh"},
        {"abc", "cba"},
        {"a", "a"},
        {"", ""},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got := Reverse(tt.input)
            if got != tt.want {
                t.Errorf("Reverse(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}
```

```go
// Using the package
package main

import (
    "fmt"
    "github.com/gluppler/hello/morestrings"
)

func main() {
    fmt.Println(morestrings.Reverse("hello")) // prints: olleh
}
```

### Internal Package Example

```go
// internal/handler/user.go
package handler

import "context"

type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

type handler struct {
    svc UserService
}

func New(svc UserService) *handler {
    return &handler{svc: svc}
}
```

### Exporting vs Unexporting

```go
package mypackage

// Exported (public) - starts with uppercase
func NewService() *Service {
    return &Service{}
}

// Unexported (private) - starts with lowercase
type service struct{}
```

### Testing Private Functions (In-Package)

```go
// Test in same package to access unexported functions
package mypackage

import "testing"

func TestPrivateFunction(t *testing.T) {
    // Can test unexported things
    result := unexportedHelper()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Build and Test Package

```bash
# Build specific package
go build -v ./pkg/mypackage
go build -v ./internal/mypackage

# Test package
go test -v ./pkg/mypackage/...
go test -v ./internal/...

# Test all packages
go test -race -v ./...
```

### Relative Imports (Within Same Module)

```go
// Using relative path within module
import "github.com/user/project/internal/mypackage"
```

### Quick Reference

| Action | Command |
|--------|---------|
| Create package | `mkdir -p pkg/mypackage` |
| Create internal | `mkdir -p internal/mypackage` |
| Create cmd entry | `mkdir -p cmd/myapp` |
| Build package | `go build -v ./pkg/mypackage` |
| Test package | `go test -v ./pkg/mypackage` |
| Test all | `go test -race -v ./...` |