---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Patterns that feel native to Go — table-driven tests, functional options, the accept interfaces/return structs rule, constructor functions, and the blank identifier.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. How to Write Go Code, go.dev."
---

# Go idiomatic patterns

## Table-driven tests

The standard Go test pattern. One slice of test cases, one loop, one call. Adding a new case is a single line:

```go
func TestReverseRunes(t *testing.T) {
    cases := []struct {
        name string
        in   string
        want string
    }{
        {"ascii", "Hello, world", "dlrow ,olleH"},
        {"unicode", "Hello, 世界", "界世 ,olleH"},
        {"empty", "", ""},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            t.Parallel()
            got := ReverseRunes(c.in)
            if got != c.want {
                t.Errorf("got %q, want %q", got, c.want)
            }
        })
    }
}
```

`t.Run` gives each case its own subtest — failures show the name, and you can run a single case with `go test -run TestReverseRunes/unicode`. `t.Parallel()` inside the subtest lets cases run concurrently, which catches races and speeds up slow tests.

Use `t.Errorf` (not `t.Fatalf`) so all cases run even when one fails. Anonymous structs are fine; no need to define a named type.

## Accept interfaces, return structs

Functions should accept the broadest interface they need and return the most concrete type they have.

```go
// bad — forces caller to use *os.File specifically
func process(f *os.File) error

// good — works with any reader
func process(r io.Reader) error
```

Accepting `io.Reader` lets callers pass a file, a network connection, a `bytes.Buffer`, a test fixture — anything. The function doesn't care.

Returning a concrete type is equally important. If you return an interface, callers can't access methods the concrete type has that aren't on the interface. If you later add methods to the concrete type, callers can use them without a breaking change.

The exception: if the implementation details genuinely shouldn't be exposed, return an interface. The `io`, `http`, and `sort` packages do this deliberately.

## Functional options

The standard way to handle optional configuration without a growing list of parameters or a config struct that's half-empty:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{
        host:    host,
        port:    8080,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

Callers apply only what they need:

```go
s := NewServer("localhost", WithTimeout(60*time.Second))
```

Adding a new option is backwards compatible — existing callers don't need updating.

## Constructor functions

Go has no constructors, but the convention is a `New` or `NewXxx` function that returns an initialized value:

```go
func NewFile(fd int, name string) *File {
    if fd < 0 {
        return nil
    }
    return &File{fd: fd, name: name}
}
```

When the package exports only one type, `New` is enough. When it exports several, use `NewFoo`, `NewBar`:

```go
ring.New(size)     // package ring, type Ring
list.New()         // package list, type List
```

Don't prefix with `Get` — `Owner()` not `GetOwner()`. The exported name already signals it's a getter.

## The blank identifier

`_` is a write-only placeholder. It discards values you don't need:

```go
for _, value := range slice { ... }        // discard index
_, err := os.Stat(path)                    // discard file info
```

It's also used to verify interface satisfaction at compile time:

```go
var _ json.Marshaler = (*RawMessage)(nil)
```

This assignment fails to compile if `*RawMessage` doesn't implement `json.Marshaler`. It's a compile-time assertion with no runtime cost.

And to import packages purely for their `init` side effects:

```go
import _ "net/http/pprof"
```

## Embedding over inheritance

Go has no subclassing. Use embedding when you want to reuse behavior:

```go
type Logger struct { /* ... */ }
func (l *Logger) Log(msg string) { /* ... */ }

type Server struct {
    *Logger         // embed, not inherit
    host string
}

s := &Server{Logger: log.New(...), host: "localhost"}
s.Log("starting")  // calls s.Logger.Log
```

Embedding promotes methods to the outer type, but the inner type's methods still have the inner type as receiver. There's no polymorphism — `Server` isn't a `Logger`, it has one.

## Error guard pattern

Go's `if err != nil` is verbose, but the style it enforces keeps the happy path left-aligned and the error cases tucked at the right:

```go
f, err := os.Open(name)
if err != nil {
    return err
}
d, err := f.Stat()
if err != nil {
    f.Close()
    return err
}
codeUsing(f, d)
```

The logic reads top to bottom. Error handling is local to where the error occurs. When the `else` clause would just return anyway, drop it:

```go
// verbose
if err != nil {
    return err
} else {
    doWork()
}

// idiomatic
if err != nil {
    return err
}
doWork()
```

## Mock via interface

The standard Go mocking pattern doesn't need a library. Define the interface next to the consumer, write a fake struct in the test:

```go
// production code
type Notifier interface {
    Notify(ctx context.Context, msg string) error
}

type AlertService struct {
    notifier Notifier
}

// test file
type fakeNotifier struct {
    sent []string
    err  error
}

func (f *fakeNotifier) Notify(_ context.Context, msg string) error {
    if f.err != nil {
        return f.err
    }
    f.sent = append(f.sent, msg)
    return nil
}

func TestAlertService_sends(t *testing.T) {
    fake := &fakeNotifier{}
    svc := AlertService{notifier: fake}
    svc.Alert(context.Background(), "disk full")
    if len(fake.sent) != 1 {
        t.Fatalf("expected 1 notification, got %d", len(fake.sent))
    }
}
```

Handwritten fakes are more readable than generated mocks and don't require importing a mock framework. Use `gomock` or `testify/mock` if you have a large interface and need call-count assertions, but for most cases a fake struct is enough.

## t.Helper

Mark helper functions with `t.Helper()` so test failures report the line where the helper was called, not the line inside the helper:

```go
func assertEqual(t *testing.T, got, want string) {
    t.Helper()
    if got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}
```

Without `t.Helper()`, the error points to `t.Errorf(...)` inside `assertEqual`, which tells you nothing useful about which test case failed.

## Benchmark pattern

Benchmarks live in `_test.go` files and are named `BenchmarkXxx`:

```go
func BenchmarkReverseRunes(b *testing.B) {
    s := "Hello, 世界"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ReverseRunes(s)
    }
}
```

`b.ResetTimer()` discards setup time. Call it after any expensive setup so the reported ns/op reflects only the code under test. Run with `go test -bench=. -benchmem`.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Structs and Interfaces]] — accept interfaces / return structs builds on interface design
- see:: [[Go Functions]] — functional options and closures
- see:: [[Go Packages and Modules]] — table-driven tests shown in testing section
- see:: [[Go Testing]] — full testing reference

**Terms**
- Table-driven test, t.Run, t.Parallel, t.Helper, functional options, accept interfaces return structs, constructor function, blank identifier, compile-time assertion, embedding, error guard pattern, mock via interface, benchmark, b.ResetTimer
