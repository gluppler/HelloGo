---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Structs and interfaces in Go — field definitions, embedding, promoted methods, interface type sets, implicit satisfaction, type assertions, and type switches.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Go structs and interfaces

## Structs

A struct is a sequence of named fields:

```go
type Point struct {
    x, y float64
}
```

Fields with the same type can be grouped. Unexported fields (lowercase) are only accessible within the defining package.

Initialize with positional arguments or named fields. Named fields are almost always better — they're self-documenting and don't break when the struct gains a field:

```go
p1 := Point{1.0, 2.0}       // positional — fragile
p2 := Point{x: 1.0, y: 2.0} // named — better
p3 := Point{y: 3.0}         // x gets zero value
```

Returning a pointer to a local struct is fine. The compiler will escape it to the heap:

```go
func NewPoint(x, y float64) *Point {
    return &Point{x: x, y: y}
}
```

## Embedding

Go has no inheritance. It has embedding, which is different.

When you embed a type in a struct, its fields and methods are promoted to the outer type. The outer type doesn't inherit anything — the embedded type's methods still have the embedded type as their receiver:

```go
type Job struct {
    Command string
    *log.Logger
}

job := &Job{"build", log.New(os.Stderr, "Job: ", log.Ldate)}
job.Println("starting")  // calls job.Logger.Println
```

This means `bufio.ReadWriter` can embed `*bufio.Reader` and `*bufio.Writer` and automatically satisfy `io.Reader`, `io.Writer`, and `io.ReadWriter` without writing any forwarding methods.

When names conflict, the shallower field wins. If `Job` had its own `Print` method, it would shadow `Logger.Print`. If two embedded types define the same method at the same depth, it's a compile error if you try to call it.

## Interfaces

An interface in Go is a set of method signatures. A type satisfies an interface by implementing all its methods — no explicit declaration needed.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Any type with a `Write([]byte) (int, error)` method automatically implements `Writer`. This implicit satisfaction is what makes Go's interface system so composable. The `http.Handler`, `io.Reader`, `fmt.Stringer` interfaces are implemented by dozens of unrelated types without any of them knowing about each other.

One-method interfaces are named by the method plus `-er`: `Reader`, `Writer`, `Closer`, `Stringer`. When you name a method `Read` or `Write` or `String`, it must have the same signature as the canonical version or you'll confuse everyone.

## Interface composition

Interfaces can embed other interfaces:

```go
type ReadWriter interface {
    Reader
    Writer
}
```

The type set of `ReadWriter` is the intersection — only types that implement both. This is how the `io` package is organized: small single-method interfaces that compose into larger ones.

## Type assertions

A type assertion extracts the concrete value from an interface:

```go
var w io.Writer = os.Stdout
f := w.(*os.File)  // panics if w doesn't hold *os.File
```

Use the comma-ok form to avoid panics:

```go
f, ok := w.(*os.File)
if ok {
    // f is *os.File
}
```

## Type switches

A type switch dispatches on the dynamic type of an interface value:

```go
var t interface{} = someValue()
switch v := t.(type) {
case bool:
    fmt.Printf("boolean: %t\n", v)
case int:
    fmt.Printf("integer: %d\n", v)
case *os.File:
    fmt.Printf("file: %s\n", v.Name())
default:
    fmt.Printf("unknown type: %T\n", v)
}
```

Inside each case, `v` has the type of that case, not `interface{}`. The `%T` verb prints the dynamic type, which is useful for debugging.

## The empty interface

`interface{}` (or `any`, its predeclared alias) is satisfied by every type. It lets you write generic-feeling code that works on any value — but you give up type safety and need assertions to get anything useful back out.

Before generics (Go 1.18), `any` was how you wrote containers and utilities. Now you can often use type parameters instead.

## Generality pattern

If a type exists only to implement an interface, export the interface and not the type. Return the interface from constructors:

```go
// good: clients get io.ReadCloser, not *myReader
func NewReader(r io.Reader) io.ReadCloser {
    return &myReader{r}
}
```

This lets you swap implementations without changing callers.

## Interface segregation

Keep interfaces small. A function that needs to read should accept `io.Reader`, not `io.ReadWriteCloser`. The caller passes in exactly what they have, and the function can't accidentally call methods it doesn't need.

```go
// fat interface — function doesn't need Write or Close
func process(rw io.ReadWriteCloser) error

// segregated — takes only what it needs
func process(r io.Reader) error
```

When you're designing a service layer, the pattern is: define the interface next to the code that consumes it, not next to the implementation. Two packages can both define an `interface{ Save(ctx, item) error }` without knowing about each other. Any type satisfying one satisfies the other, because Go compares method sets structurally.

## Dependency injection via interfaces

Interfaces make dependencies explicit and swappable. Rather than hardcoding a concrete database client, accept an interface:

```go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, u *User) error
}

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

In production code you inject a real database. In tests you inject a fake. The service doesn't know or care which it got.

## Compile-time interface verification

If a type is supposed to implement an interface, prove it at compile time:

```go
var _ io.Reader = (*MyReader)(nil)
var _ http.Handler = (*MyHandler)(nil)
```

This line does nothing at runtime — it's a zero-value pointer assigned to an interface variable. If the method set is wrong, the build fails immediately. Put these declarations near the type definition, not in tests. Finding out at build time beats finding out at test time.

## Custom Reader and Writer

Any type with the right signature satisfies `io.Reader` or `io.Writer`. You don't import anything special or register anywhere.

```go
type rot13Reader struct {
    r io.Reader
}

func (r rot13Reader) Read(p []byte) (n int, err error) {
    n, err = r.r.Read(p)
    for i := range p[:n] {
        if p[i] >= 'A' && p[i] <= 'Z' {
            p[i] = 'A' + (p[i]-'A'+13)%26
        } else if p[i] >= 'a' && p[i] <= 'z' {
            p[i] = 'a' + (p[i]-'a'+13)%26
        }
    }
    return
}
```

Wrap it around any reader:

```go
r := rot13Reader{strings.NewReader("Hello, World!")}
io.Copy(os.Stdout, r)
```

The entire `io` package is built this way — transformations are just wrappers that satisfy the same interface.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Types and Variables]] — structs build on primitive types
- see:: [[Go Idiomatic Patterns]] — embedding over inheritance, accept interfaces return structs
- see:: [[Go HTTP]] — http.Handler is a great example of interface design

**Terms**
- Struct embedding, promoted method, implicit interface satisfaction, type assertion, type switch, empty interface, any, interface composition, interface segregation, dependency injection, compile-time verification, custom Reader, custom Writer
