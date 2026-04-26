---
tags:
  - type/note
  - theme/golang
  - theme/concurrency
aliases: []
lead: Goroutines — the scheduler, lifecycle management with context and errCh, worker pools, fan-out/fan-in, pipelines, and goroutine leak prevention.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. golang-pro reference: concurrency."
---

# Goroutines

## What they are

A goroutine is a function executing concurrently with other goroutines in the same address space. Not OS threads, not coroutines in the traditional sense — the name is intentionally distinct.

They're cheap. A goroutine starts with a few kilobytes of stack that grows and shrinks dynamically. You can have hundreds of thousands of them without trouble.

```go
go list.Sort()  // start sorting concurrently, don't wait
```

When the function returns, the goroutine exits silently. There's no handle, no goroutine ID.

## The scheduler

Go's runtime uses M:N scheduling — M goroutines multiplexed onto N OS threads. If a goroutine blocks on I/O or a syscall, the runtime parks it and runs another on the same thread. This is why goroutines are cheap even when thousands are waiting.

`GOMAXPROCS` controls how many OS threads can execute Go code simultaneously. It defaults to `runtime.NumCPU()`. Query without setting:

```go
var numCPU = runtime.GOMAXPROCS(0)
```

Concurrency (structuring a program as independently executing components) is different from parallelism (executing on multiple CPUs). Go is a concurrent language. Parallelism falls out of it when `GOMAXPROCS > 1`.

## Goroutine lifecycle — the right way

Every goroutine should have a clear exit condition. The idiomatic approach: pass a context, return errors via a channel.

```go
// worker runs until ctx is cancelled or an error occurs.
// Errors are returned via errCh; the caller must drain it.
func worker(ctx context.Context, jobs <-chan Job, errCh chan<- error) {
    for {
        select {
        case <-ctx.Done():
            errCh <- fmt.Errorf("worker cancelled: %w", ctx.Err())
            return
        case job, ok := <-jobs:
            if !ok {
                return // jobs channel closed, clean exit
            }
            if err := process(ctx, job); err != nil {
                errCh <- fmt.Errorf("process job %v: %w", job.ID, err)
                return
            }
        }
    }
}

func runPipeline(ctx context.Context, jobs []Job) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    jobCh := make(chan Job, len(jobs))
    errCh := make(chan error, 1)

    go worker(ctx, jobCh, errCh)

    for _, j := range jobs {
        jobCh <- j
    }
    close(jobCh)

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return fmt.Errorf("pipeline timed out: %w", ctx.Err())
    }
}
```

Key properties: bounded lifetime via context, error propagation with `%w`, no goroutine leak on cancellation.

## Worker pool

Fixed number of goroutines processing from a shared queue. This bounds memory and CPU usage regardless of how many tasks come in.

```go
type WorkerPool struct {
    workers int
    tasks   chan func()
    wg      sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    wp := &WorkerPool{
        workers: workers,
        tasks:   make(chan func(), workers*2),
    }
    for i := 0; i < workers; i++ {
        wp.wg.Add(1)
        go func() {
            defer wp.wg.Done()
            for task := range wp.tasks {
                task()
            }
        }()
    }
    return wp
}

func (wp *WorkerPool) Submit(task func()) { wp.tasks <- task }

func (wp *WorkerPool) Shutdown() {
    close(wp.tasks)
    wp.wg.Wait()
}
```

## Generator pattern

A goroutine that produces values and closes the output channel when done:

```go
func generateNumbers(ctx context.Context, max int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 0; i < max; i++ {
            select {
            case out <- i:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

Always check `ctx.Done()` in the send select. Without it the goroutine leaks when the consumer stops listening.

## Fan-out / fan-in

Fan-out: one input channel read by multiple worker goroutines. Fan-in: results from multiple goroutines merged into one output channel.

```go
func fanOut(ctx context.Context, input <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = processStage(ctx, input)
    }
    return channels
}

func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                select {
                case out <- val:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

## Pipeline pattern

Each stage reads from an input channel, transforms values, and writes to an output channel:

```go
func pipeline(ctx context.Context, input <-chan int) <-chan int {
    // Stage 1: square
    stage1 := make(chan int)
    go func() {
        defer close(stage1)
        for num := range input {
            select {
            case stage1 <- num * num:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Stage 2: filter evens
    stage2 := make(chan int)
    go func() {
        defer close(stage2)
        for num := range stage1 {
            if num%2 == 0 {
                select {
                case stage2 <- num:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return stage2
}
```

## Closure capture

Function literals capture variables by reference. The classic loop bug:

```go
// Bug (Go < 1.22): all goroutines print the same final value of i
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }()
}

// Fix: pass i as an argument
for i := 0; i < 5; i++ {
    go func(n int) { fmt.Println(n) }(i)
}
```

Since Go 1.22, each loop iteration has its own variable. Still good to pass as argument when targeting older versions or when clarity matters.

## Goroutine leaks

A goroutine blocked with no way to unblock is a leak. In a server, leaks accumulate. Common causes:

- Sending to an unbuffered channel with no receiver
- Receiving from a channel that's never closed or sent to
- Waiting on a mutex that's never unlocked

The fix is always the same: give every goroutine a way out via context cancellation, a done channel, or closing its input.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Channels]] — how goroutines communicate; generator and fan-in patterns
- see:: [[Select and Sync]] — select, WaitGroup, worker pool coordination

**Terms**
- Goroutine, M:N scheduler, GOMAXPROCS, worker pool, generator pattern, fan-out, fan-in, pipeline, closure capture, goroutine leak, errCh
