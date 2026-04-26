---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Go's type system — built-in types, zero values, short declarations, constants, iota, and the new vs make distinction.
created: 2026-04-26
modified: 2026-04-26
source: "The Go Programming Language Specification, go1.26. Effective Go, 2024."
---

# Go types and variables

## Built-in types

Go has a fixed set of predeclared types. The numeric ones are explicit about size:

```go
uint8, uint16, uint32, uint64
int8, int16, int32, int64
float32, float64
complex64, complex128
byte   // alias for uint8
rune   // alias for int32
```

Then the architecture-dependent ones: `int`, `uint`, and `uintptr` are 32 or 64 bits depending on the platform. In practice you use `int` unless you have a specific reason not to.

`bool`, `string`. That's the full primitive set.

## Zero values

Every variable gets a zero value if you don't initialize it. This matters more than it sounds. The zero for `int` is `0`, for `bool` is `false`, for `string` is `""`, and for pointers, slices, maps, and channels is `nil`.

Go's standard library leans into this. `sync.Mutex` requires no initialization — the zero value is an unlocked mutex. `bytes.Buffer` works the same way. When you design your own types, making the zero value useful saves callers from needing a constructor.

```go
type SyncedBuffer struct {
    lock   sync.Mutex
    buffer bytes.Buffer
}

var v SyncedBuffer  // ready to use without any setup
```

## Declaring variables

Three ways, depending on context:

```go
var i int               // explicit type, zero value
var x float64 = 3.14   // explicit type and value
y := 42                // short declaration, type inferred
```

Short declarations (`:=`) only work inside functions. The type is inferred from the right side. If the right side is an untyped integer constant like `42`, it becomes `int`. If it's `42.0`, it becomes `float64`.

`:=` can redeclare a variable from the same scope if at least one variable on the left is new:

```go
f, err := os.Open(name)
d, err := f.Stat()   // err is reassigned here, not redeclared
```

This is pragmatic, not elegant. It lets you reuse `err` across a chain of calls without cluttering the code.

## Constants and iota

Constants are evaluated at compile time. They can be numbers, booleans, runes, or strings — nothing else.

`iota` is a counter that resets to 0 at the start of each `const` block and increments by 1 for each spec:

```go
const (
    Sunday = iota  // 0
    Monday         // 1
    Tuesday        // 2
)
```

The expression repeats, so this gives you powers of 2:

```go
type ByteSize float64

const (
    _           = iota  // skip zero
    KB ByteSize = 1 << (10 * iota)
    MB
    GB
    TB
)
```

`iota` within the same ConstSpec has the same value across all names:

```go
const (
    bit0, mask0 = 1 << iota, 1<<iota - 1  // 1, 0  (iota=0)
    bit1, mask1                           // 2, 1  (iota=1)
)
```

## new vs make

Two allocation built-ins that do different things.

`new(T)` allocates memory for a `T`, zeros it, and returns `*T`. You rarely need it directly because composite literals do the same thing more readably.

`make(T, args)` is for slices, maps, and channels only. It creates the internal data structure and returns an initialized value of type `T` (not `*T`).

```go
var p *[]int = new([]int)       // pointer to nil slice — rarely useful
var v  []int = make([]int, 100) // a real slice backed by an array
```

The distinction exists because slices, maps, and channels are descriptors pointing to underlying data. A zeroed slice descriptor is `nil` and useless. `make` sets up the internals so the value is ready.

## Slices

A slice is a three-field struct: pointer to array, length, capacity. Slices are passed by value, but the value contains a pointer to the underlying array — so modifying elements inside a function is visible to the caller.

```go
s := make([]int, 5, 10)
s = append(s, 6, 7)  // len=7, still same array if cap allows
```

Arrays in Go are values: assigning one array to another copies all elements. For that reason idiomatic Go almost always uses slices instead.

## Maps

Maps need `make` before you can write to them. A nil map is readable (returns zero values) but panics on write.

```go
m := make(map[string]int)
m["key"] = 1

// comma-ok distinguishes missing key from zero value
val, ok := m["key"]
```

The comma-ok idiom is how you distinguish "key not present" from "key present with zero value." You'll use it constantly.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Functions]] — functions are how you work with these types
- see:: [[Go Structs and Interfaces]] — composite types built on top of these primitives

**Terms**
- Zero value, type inference, short declaration, iota, const block, new, make, slice descriptor, comma-ok idiom
