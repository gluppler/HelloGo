---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Generics in Go 1.18+ — type parameters, constraints, the ~T approximation element, generic data structures, and common utility functions.
created: 2026-04-26
modified: 2026-04-26
source: "The Go Programming Language Specification, go1.26. golang-pro reference: generics."
---

# Go generics

## What they are

Generics let you write functions and types that work over a family of types rather than a specific one. Before Go 1.18, the options were `interface{}` (lose type safety) or code generation (lose readability). Type parameters are the third option.

```go
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

Min(3, 5)       // T inferred as int
Min(3.0, 5.0)   // T inferred as float64
Min("a", "b")   // T inferred as string
```

The `[T constraints.Ordered]` part is the type parameter list. `T` is the name; `constraints.Ordered` is the constraint — the set of types `T` can be.

## Constraints

A constraint is an interface used as a bound on a type parameter. It can list methods or specific types.

The built-in constraints:

```go
any          // any type (equivalent to interface{})
comparable   // types that support == and !=
```

From `golang.org/x/exp/constraints`:

```go
constraints.Ordered    // integers, floats, strings — types that support < > <= >=
constraints.Integer    // all integer types
constraints.Float      // all float types
constraints.Signed     // all signed integer types
constraints.Unsigned   // all unsigned integer types
```

You can define your own:

```go
type Number interface {
    int | int64 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}
```

The `|` in an interface means "union of these types." A type parameter constrained by `Number` can only be one of `int`, `int64`, or `float64`.

## The ~T approximation element

`~T` means "T and all types whose underlying type is T." Without it, custom types built on a base type don't satisfy the constraint:

```go
type Celsius float64

// This fails — Celsius is not float64, it just has float64 as its underlying type
func Freeze[T float64](c T) bool { return c <= 0 }

// This works — ~float64 includes Celsius
func Freeze[T ~float64](c T) bool { return c <= 0 }
```

Most numeric constraints use `~` for exactly this reason: you want to accept both the primitive type and any named types built on it.

## Generic data structures

Type parameters work on structs too:

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) Len() int { return len(s.items) }
```

Usage:

```go
s := Stack[int]{}
s.Push(1)
s.Push(2)
v, ok := s.Pop()  // v = 2, ok = true
```

The zero-value return trick (`var zero T; return zero, false`) is how you return the zero value of a type parameter — you can't write `return nil, false` because `T` may not be a pointer type.

## Map, Filter, Reduce

The classic functional trio, finally expressible without `interface{}`:

```go
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s {
        result[i] = f(v)
    }
    return result
}

func Filter[T any](s []T, f func(T) bool) []T {
    var result []T
    for _, v := range s {
        if f(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
    acc := init
    for _, v := range s {
        acc = f(acc, v)
    }
    return acc
}
```

These are useful, but don't reach for them by default. A plain loop is usually clearer. Reach for generics when the abstraction removes real repetition across multiple callers, not just to avoid writing a for loop once.

## Generic channels

Type parameters on channels let you write typed pipeline stages:

```go
func OrDone[T any](ctx context.Context, c <-chan T) <-chan T {
    out := make(chan T)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case v, ok := <-c:
                if !ok {
                    return
                }
                select {
                case out <- v:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}
```

`OrDone` wraps any receive channel and short-circuits when the context is cancelled, so callers don't have to write the nested select themselves.

## When to use generics

Use type parameters when:
- You're writing a container (stack, queue, set, cache) that should work over multiple types
- You have a utility function used across multiple call sites with different concrete types
- The alternative would be `interface{}` with type assertions spread everywhere

Don't use them when:
- A single concrete type is enough
- An interface with methods covers the use case cleanly
- The function is only called once or twice — the abstraction costs clarity

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Structs and Interfaces]] — interfaces underpin constraints
- see:: [[Go Types and Variables]] — underlying types and the ~T element
- see:: [[Channels]] — generic channels in pipeline stages

**Terms**
- Type parameter, constraint, comparable, any, constraints.Ordered, union constraint, approximation element ~T, generic data structure, Stack, Map Filter Reduce, zero value of type parameter
