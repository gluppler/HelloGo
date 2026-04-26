---
tags:
  - type/note
  - theme/golang
  - theme/concurrency
aliases: []
lead: Channels — typed message passing between goroutines. Buffered vs unbuffered semantics, directional types, closing, ranging, and channels as semaphores.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Channels

## The basics

A channel is a typed conduit for passing values between goroutines. Like maps, channels are created with `make`:

```go
ci := make(chan int)            // unbuffered
cs := make(chan *os.File, 100)  // buffered, capacity 100
```

Send with `<-` on the right, receive with `<-` on the left:

```go
ch <- value    // send
value = <-ch   // receive
```

An uninitialized channel is `nil`. Sending to or receiving from a nil channel blocks forever. Closing a nil channel panics.

## Unbuffered vs buffered

**Unbuffered** (capacity 0): send blocks until a receiver is ready. Receive blocks until a sender sends. The send and receive happen simultaneously — it's a rendezvous, not just data transfer. This gives you synchronization for free.

**Buffered** (capacity N): send only blocks when the buffer is full. Receive only blocks when the buffer is empty. Useful when the sender and receiver work at different rates.

```go
// Wait for a background sort to complete
c := make(chan int)
go func() {
    list.Sort()
    c <- 1  // signal done
}()
doOtherWork()
<-c  // wait for signal
```

## Directional channel types

Channels can be typed to allow only sending or only receiving:

```go
chan<- int   // send-only
<-chan int   // receive-only
chan int     // bidirectional
```

A bidirectional channel converts implicitly to a directional one. Use directional types in function signatures to document intent and catch mistakes at compile time:

```go
func producer(out chan<- int) {
    for i := 0; i < 10; i++ {
        out <- i
    }
    close(out)
}

func consumer(in <-chan int) {
    for v := range in {
        fmt.Println(v)
    }
}
```

## Closing channels

`close(ch)` signals that no more values will be sent. Receivers drain the remaining values, then get zero values with `ok == false`:

```go
v, ok := <-ch
// ok is false when channel is closed and drained
```

Ranging over a channel reads until close:

```go
for v := range ch {
    process(v)
}
```

Rules to live by:
- Only the sender should close a channel, never the receiver.
- Closing a closed channel panics.
- You don't have to close a channel. Closing is only needed to signal "no more data" to a range loop or a receiving goroutine.

## Channel as semaphore

A buffered channel with capacity N limits concurrency to N:

```go
var sem = make(chan struct{}, MaxOutstanding)

func handle(r *Request) {
    sem <- struct{}{}  // acquire slot
    process(r)
    <-sem              // release slot
}
```

`struct{}` is idiomatic for signal-only channels — it carries no data and allocates nothing.

## Channels of channels

Channels are first-class values, so a channel can carry channels. This lets clients provide their own reply path:

```go
type Request struct {
    args       []int
    f          func([]int) int
    resultChan chan int
}

// client
req := &Request{args, sumFunc, make(chan int)}
requestQueue <- req
result := <-req.resultChan  // wait for answer
```

The server puts the result directly into the channel the client provided. No shared state, no polling.

## Leaky buffer pattern

A free list of buffers using a buffered channel:

```go
var freeList = make(chan *Buffer, 100)

func getBuffer() *Buffer {
    select {
    case b := <-freeList:
        return b
    default:
        return new(Buffer)
    }
}

func returnBuffer(b *Buffer) {
    select {
    case freeList <- b:
        // returned
    default:
        // list full, drop it — GC handles it
    }
}
```

The `default` case in `select` makes the operations non-blocking. If the free list is empty, allocate. If it's full, drop. Simple, no mutex needed.

## Generator pattern

A function that produces values into a channel and closes it when done. The caller treats it like an iterator.

```go
func generate(ctx context.Context, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

The `ctx.Done()` case in the send select is not optional. Without it, if the consumer stops reading, the goroutine blocks on `out <- n` and leaks. Context gives it a way out.

## Fan-out / fan-in over channels

Fan-out means multiple goroutines reading from the same input channel — each item is processed by exactly one of them. Fan-in means multiple goroutines writing to the same output channel, with results merged.

```go
func merge(ctx context.Context, cs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    output := func(c <-chan int) {
        defer wg.Done()
        for n := range c {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }

    wg.Add(len(cs))
    for _, c := range cs {
        go output(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

The WaitGroup tracks when all input channels are drained. The closer goroutine waits on it so `out` is only closed when no more values are coming. Without the `ctx.Done()` arm, cancellation would leave the output goroutines stuck.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Goroutines]] — goroutines are what you connect with channels; generator and fan-in patterns
- see:: [[Select and Sync]] — select multiplexes across multiple channels

**Terms**
- Buffered channel, unbuffered channel, directional channel, close, range over channel, semaphore pattern, channels of channels, generator pattern, fan-out, fan-in
