---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Error handling in Go — the error interface, custom error types, wrapping with %w, errors.Is and errors.As, and when to use panic/recover.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Go error handling

## The error interface

`error` is a built-in interface with one method:

```go
type error interface {
    Error() string
}
```

Functions signal failure by returning a non-nil `error` as the last return value. The caller checks it. This is verbose by design — errors are values, not exceptions, and handling them explicitly is intentional.

```go
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()
```

Never ignore an error by assigning to `_` unless you genuinely don't care. Discarded errors are a common source of subtle bugs.

## Custom error types

Implement the `error` interface on any struct to add context:

```go
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}
```

`os.PathError` works exactly this way. When `os.Open` fails, you get a `*PathError` that tells you the operation, the path, and the underlying OS error — not just "file not found."

Error strings should identify their origin. Prefix them with the package or operation name:

```go
// good
return fmt.Errorf("image: unknown format")
return fmt.Errorf("sql: no rows in result set")

// bad — no context
return errors.New("not found")
```

## Wrapping errors

`fmt.Errorf` with `%w` wraps an error while adding context:

```go
if err := db.Query(query); err != nil {
    return fmt.Errorf("fetchUser %d: %w", id, err)
}
```

This builds an error chain. Each layer adds context. The resulting message reads like a stack of what went wrong:

```
fetchUser 42: sql: no rows in result set
```

## errors.Is and errors.As

Wrapping changes the error's type, but you still want to inspect it. `errors.Is` checks if any error in the chain matches a target:

```go
if errors.Is(err, os.ErrNotExist) {
    // file doesn't exist
}
```

`errors.As` finds the first error in the chain of a given type and assigns it:

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("operation:", pathErr.Op)
    fmt.Println("path:", pathErr.Path)
}
```

Don't use `==` to compare errors directly unless you're comparing against a sentinel you control. Wrapping breaks `==`.

## Sentinel errors

A sentinel is a package-level error value used to signal a specific condition:

```go
var ErrNotFound = errors.New("not found")
var ErrPermission = errors.New("permission denied")
```

Callers compare with `errors.Is(err, ErrNotFound)`. Sentinels are part of your package's API — document them and don't remove them.

## panic and recover

`panic` stops normal execution, unwinds the call stack, and runs deferred functions as it goes. If nothing catches it, the program crashes.

Use `panic` for programming errors — the kind that should never happen if the code is correct: index out of range, nil pointer dereference, broken invariants. Don't use it for expected failure conditions like "file not found."

```go
func mustCompile(expr string) *regexp.Regexp {
    re, err := regexp.Compile(expr)
    if err != nil {
        panic(err)  // bad regex in a constant is a programmer error
    }
    return re
}
```

`recover` catches a panic inside a deferred function and returns the value passed to `panic`. It only works directly inside a defer:

```go
func safelyDo(work *Work) {
    defer func() {
        if err := recover(); err != nil {
            log.Println("work failed:", err)
        }
    }()
    do(work)
}
```

The classic use in servers: wrap each request handler in a recover so one bad request doesn't crash the entire process.

The pattern used in `regexp` and similar packages: use panic internally for error propagation within the package, recover at the package boundary, and convert back to a regular error for callers:

```go
func (p *parser) parse(s string) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("parse error: %v", r)
        }
    }()
    p.parseExpression()  // panics on syntax error
    return nil
}
```

This keeps the internal parsing code clean while presenting a normal `error` return to callers.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Functions]] — functions return (T, error) by convention
- see:: [[Go Structs and Interfaces]] — custom errors implement the error interface

**Terms**
- error interface, custom error type, fmt.Errorf %w, error wrapping, errors.Is, errors.As, sentinel error, panic, recover, error chain
