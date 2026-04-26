---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Functions in Go — multiple return values, named returns, variadic, defer, closures, and the pointer vs value receiver rule.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Go functions

## Multiple return values

Functions can return more than one value. This isn't a trick — it's how Go handles errors. The standard pattern is `(result, error)`:

```go
func (file *os.File) Write(b []byte) (n int, err error)
```

You get both pieces of information without having to inspect an out-parameter or check a global error state. In C you'd stuff the error into a return code and the real value into a side effect. Go just returns both.

A practical example — scanning a number from a byte slice and returning where the number ended:

```go
func nextInt(b []byte, i int) (int, int) {
    for ; i < len(b) && !isDigit(b[i]); i++ {}
    x := 0
    for ; i < len(b) && isDigit(b[i]); i++ {
        x = x*10 + int(b[i]) - '0'
    }
    return x, i
}
```

## Named return values

Return values can have names. When named, they're initialized to zero at function entry and a bare `return` sends their current values.

This is mostly useful for documenting what the returns mean, and occasionally for simplifying logic:

```go
func ReadFull(r Reader, buf []byte) (n int, err error) {
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    return
}
```

Don't overuse named returns. Bare `return` in a long function is hard to follow.

## Variadic functions

The last parameter can be variadic with `...`:

```go
func Min(a ...int) int {
    min := int(^uint(0) >> 1)
    for _, i := range a {
        if i < min {
            min = i
        }
    }
    return min
}
```

Pass a slice to a variadic function by appending `...` at the call site:

```go
x := []int{1, 2, 3}
Min(x...)  // equivalent to Min(1, 2, 3)
```

## defer

`defer` schedules a function call to run when the surrounding function returns, regardless of how it returns. Arguments to the deferred function are evaluated immediately when `defer` executes, not when the deferred call happens.

The canonical use is cleanup paired with the thing being cleaned up:

```go
func Contents(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()  // will run no matter what happens below

    // ... read f ...
}
```

Deferred calls run in LIFO order. If you defer multiple things, the last deferred runs first.

One subtlety: the deferred call's arguments are captured at defer time, not execution time. This is useful for tracing:

```go
func trace(s string) string { fmt.Println("entering:", s); return s }
func un(s string)           { fmt.Println("leaving:", s) }

func a() {
    defer un(trace("a"))  // trace("a") runs now, un() runs on exit
    fmt.Println("in a")
}
```

## Closures

Function literals capture variables from the surrounding scope. The variable is shared, not copied:

```go
func Announce(message string, delay time.Duration) {
    go func() {
        time.Sleep(delay)
        fmt.Println(message)  // message captured by reference
    }()
}
```

This is where the classic goroutine loop bug comes from. If you close over a loop variable, all goroutines share the same variable:

```go
// Bug in Go < 1.22: all goroutines see the last value of i
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }()
}

// Fix: pass i as a parameter
for i := 0; i < 5; i++ {
    go func(n int) { fmt.Println(n) }(i)
}
```

Since Go 1.22, each loop iteration has its own variable, so the bug no longer exists in new code.

## Pointer vs value receivers

Methods can have either a pointer receiver or a value receiver.

Value methods can be invoked on both pointers and values. Pointer methods can only be invoked on pointers — with one exception: if the value is addressable, the compiler inserts the address automatically.

The rule: use a pointer receiver when the method needs to modify the receiver, or when the receiver is large. Use a value receiver otherwise. Be consistent within a type — don't mix pointer and value receivers on the same type unless you have a good reason.

```go
func (p *ByteSlice) Write(data []byte) (n int, err error) {
    // modifies *p — needs pointer receiver
    *p = append(*p, data...)
    return len(data), nil
}
```

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Types and Variables]] — parameter and return types
- see:: [[Go Error Handling]] — functions typically return (T, error)
- see:: [[Goroutines]] — defer and closures interact with goroutines in non-obvious ways

**Terms**
- Multiple return values, named returns, variadic, defer, LIFO, closure, capture by reference, pointer receiver, value receiver
