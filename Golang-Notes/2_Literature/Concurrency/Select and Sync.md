---
tags:
  - type/note
  - theme/golang
  - theme/concurrency
aliases: []
lead: Coordinating goroutines — select for multiplexing channels, sync.Mutex/RWMutex for shared state, WaitGroup for fan-out, Once for one-time init, and context for cancellation.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Select and sync

## select

`select` waits on multiple channel operations simultaneously and runs whichever is ready first. If multiple are ready, it picks one at random.

```go
select {
case msg := <-ch1:
    fmt.Println("received from ch1:", msg)
case ch2 <- value:
    fmt.Println("sent to ch2")
case <-time.After(1 * time.Second):
    fmt.Println("timed out")
}
```

A `default` clause runs immediately if no other case is ready — making the select non-blocking:

```go
select {
case v := <-ch:
    return v
default:
    return 0  // nothing ready, don't wait
}
```

`select` with no cases (`select {}`) blocks forever. Useful in `main` to keep the program alive when all work is in goroutines.

### Timeout pattern

```go
func doWithTimeout(ch <-chan Result, timeout time.Duration) (Result, bool) {
    select {
    case r := <-ch:
        return r, true
    case <-time.After(timeout):
        return Result{}, false
    }
}
```

`time.After` returns a channel that receives after the duration. This is the standard way to add timeouts without a separate context.

### Quit channel pattern

```go
func worker(work <-chan Job, quit <-chan struct{}) {
    for {
        select {
        case j := <-work:
            process(j)
        case <-quit:
            return
        }
    }
}
```

Close the `quit` channel to broadcast termination to all goroutines watching it. Closing a channel unblocks all receivers simultaneously — this is different from sending a value, which unblocks only one.

## sync.Mutex

Use a mutex when goroutines share state that can't be passed through channels cleanly — counters, caches, maps being modified concurrently.

```go
type SafeCounter struct {
    mu sync.Mutex
    v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
    c.mu.Lock()
    c.v[key]++
    c.mu.Unlock()
}
```

`defer c.mu.Unlock()` at the top of the function is cleaner and panic-safe:

```go
func (c *SafeCounter) Value(key string) int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.v[key]
}
```

`sync.RWMutex` lets multiple readers hold the lock simultaneously, but only one writer at a time. Use it when reads vastly outnumber writes:

```go
c.mu.RLock()
defer c.mu.RUnlock()
return c.v[key]
```

One rule: don't copy a mutex after first use. Always pass `*sync.Mutex`, not `sync.Mutex`.

## sync.WaitGroup

Fan out N goroutines and wait for all of them:

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(item)
    }(item)
}

wg.Wait()  // blocks until all Done() calls match Add() calls
```

Call `Add` before starting the goroutine, not inside it. If the goroutine starts before `Add`, a concurrent `Wait` might return too early.

## sync.Once

Runs a function exactly once, regardless of how many goroutines call it. Standard use: lazy initialization of something expensive.

```go
var (
    instance *Database
    once     sync.Once
)

func GetDB() *Database {
    once.Do(func() {
        instance = connectToDB()
    })
    return instance
}
```

Once `Do` returns, every subsequent call is a no-op. The function is guaranteed to complete before `Do` returns in any goroutine.

## context.Context

`context` is how you propagate cancellation, deadlines, and request-scoped values across API boundaries. Every function that might block or do I/O should accept a context as its first argument:

```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    // if ctx is cancelled, the request is cancelled
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

Create contexts:

```go
ctx := context.Background()                          // root, never cancelled
ctx, cancel := context.WithCancel(ctx)               // cancel manually
ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // deadline
defer cancel()  // always defer cancel to avoid leak
```

Check cancellation inside long-running loops:

```go
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        doWork()
    }
}
```

## Rate limiting

`golang.org/x/time/rate` implements a token bucket. At its core it's a `Limiter` that grants tokens at a steady rate, with a burst allowance:

```go
import "golang.org/x/time/rate"

// 10 requests per second, burst of 30
limiter := rate.NewLimiter(10, 30)

func makeRequest(ctx context.Context) error {
    if err := limiter.Wait(ctx); err != nil {
        return err  // ctx cancelled while waiting for a token
    }
    return doRequest()
}
```

`Wait` blocks until a token is available or ctx is cancelled. For non-blocking checks, use `Allow()` — it returns false immediately if no token is ready.

```go
if !limiter.Allow() {
    return ErrRateLimited
}
```

For per-user rate limiting, keep a `map[string]*rate.Limiter` behind a mutex:

```go
type RateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    r        rate.Limit
    b        int
}

func (rl *RateLimiter) Get(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    l, ok := rl.limiters[key]
    if !ok {
        l = rate.NewLimiter(rl.r, rl.b)
        rl.limiters[key] = l
    }
    return l
}
```

## sync.Map

`sync.Map` is a concurrent map optimized for two scenarios: write-once/read-many, or when goroutines access disjoint key sets. Outside those cases, a `map` + `sync.RWMutex` is usually faster and clearer.

```go
var m sync.Map

m.Store("key", "value")

if v, ok := m.Load("key"); ok {
    fmt.Println(v.(string))
}

m.Range(func(k, v any) bool {
    fmt.Println(k, v)
    return true  // return false to stop iteration
})
```

Don't use `sync.Map` for a general-purpose cache. Build a `struct` with an `RWMutex` and a regular map — that way you control eviction, sizing, and type safety.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Channels]] — select operates on channels; quit channel pattern
- see:: [[Goroutines]] — sync primitives coordinate goroutines; WaitGroup for fan-out

**Terms**
- select, default case, timeout pattern, quit channel, sync.Mutex, sync.RWMutex, sync.WaitGroup, sync.Once, context.Context, context cancellation, context.WithTimeout, rate.Limiter, sync.Map
