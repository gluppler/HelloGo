---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Testing in Go — table-driven tests with subtests, parallel execution, t.Helper, mocking via interfaces, benchmarks, fuzzing, and the race detector.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. golang-pro reference: testing."
---

# Go testing

## The basics

Test files end in `_test.go`. The `testing` package provides everything you need. No external framework is required.

```go
package mypackage

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    if got != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", got)
    }
}
```

Run tests:

```bash
go test ./...           # all packages
go test -v ./pkg/...    # verbose output
go test -run TestAdd    # run specific test
go test -count=1 ./...  # disable test caching
```

## Table-driven tests with subtests

Table-driven tests are the Go standard. `t.Run` turns each case into a named subtest:

```go
func TestDivide(t *testing.T) {
    cases := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"positive", 10, 2, 5, false},
        {"negative dividend", -10, 2, -5, false},
        {"divide by zero", 10, 0, 0, true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Divide(tc.a, tc.b)
            if (err != nil) != tc.wantErr {
                t.Fatalf("Divide(%v, %v) error = %v, wantErr %v", tc.a, tc.b, err, tc.wantErr)
            }
            if !tc.wantErr && got != tc.want {
                t.Errorf("Divide(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

Run a single subtest: `go test -run TestDivide/divide_by_zero` (spaces become underscores in the filter).

## Parallel tests

`t.Parallel()` inside a subtest lets cases run concurrently. This is safe when cases don't share mutable state:

```go
for _, tc := range cases {
    tc := tc  // capture range variable (required in Go < 1.22)
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()
        // test body
    })
}
```

In Go 1.22+, loop variables are per-iteration, so `tc := tc` is no longer needed. Still useful to know for codebases that target older versions.

Parallel tests run faster and surface data races. Run with `-race` to catch them:

```bash
go test -race ./...
```

## t.Helper

Any function that calls `t.Errorf` or `t.Fatal` should call `t.Helper()` at its top so that failures report the line in the calling test, not inside the helper:

```go
func requireNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

Without `t.Helper()`, the stack frame shown in the error output points inside `requireNoError` rather than the test that called it.

## Mocking via interfaces

Go's standard approach: define the interface next to the consumer, write a hand-rolled fake in the test file. No mocking library required.

```go
// In production code:
type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// In _test.go:
type fakeSender struct {
    sent []struct{ to, subject string }
    err  error
}

func (f *fakeSender) Send(_ context.Context, to, subject, body string) error {
    if f.err != nil {
        return f.err
    }
    f.sent = append(f.sent, struct{ to, subject string }{to, subject})
    return nil
}
```

Assertions then check `fake.sent` directly — no mock framework assertion DSL to learn. For interfaces with many methods, `gomock` or `testify/mock` is worth the dependency. For most cases, a fake struct is enough.

## Benchmarks

Benchmark functions are named `BenchmarkXxx` and receive `*testing.B`:

```go
func BenchmarkJSON(b *testing.B) {
    data := generatePayload()  // setup
    b.ResetTimer()             // don't count setup time
    for i := 0; i < b.N; i++ {
        json.Marshal(data)
    }
}
```

`b.N` is set by the testing framework — it runs the loop enough times to get a stable measurement. `b.ResetTimer()` discards any setup time before the loop. `b.StopTimer()` / `b.StartTimer()` can pause timing for cleanup inside the loop.

Run benchmarks:

```bash
go test -bench=. -benchmem ./...
go test -bench=BenchmarkJSON -benchtime=10s ./...
```

`-benchmem` shows allocations per op, which matters as much as time for hot paths.

## Fuzzing

Fuzz tests find inputs that cause panics or unexpected behavior by generating random mutations of a seed corpus:

```go
func FuzzParseURL(f *testing.F) {
    // seed corpus — representative valid inputs
    f.Add("https://example.com/path?q=1")
    f.Add("http://localhost:8080")
    f.Add("")

    f.Fuzz(func(t *testing.T, rawURL string) {
        u, err := url.Parse(rawURL)
        if err != nil {
            return  // parse errors are expected
        }
        // roundtrip invariant — parsed URL should re-serialize to itself
        if u.String() != rawURL {
            // not necessarily a bug, but worth logging
        }
        // the real target: this must never panic
        _ = u.Hostname()
    })
}
```

Run fuzz tests:

```bash
go test -fuzz=FuzzParseURL          # run indefinitely, Ctrl-C to stop
go test -fuzz=FuzzParseURL -fuzztime=30s  # run for 30 seconds
go test ./...                       # runs seed corpus only (no mutations)
```

When the fuzzer finds a failure, it writes the input to `testdata/fuzz/FuzzParseURL/` so the regression case is reproduced by future `go test` runs.

## Race detector

The race detector instruments memory accesses and reports data races at runtime:

```bash
go test -race ./...
go run -race main.go
go build -race -o myapp ./cmd/server
```

A data race is when two goroutines access the same variable concurrently and at least one is a write, with no synchronization. Races are undefined behavior — they can silently corrupt data or cause intermittent crashes. The race detector catches them.

The detector has a ~5x memory and ~10x CPU overhead, so don't ship a `-race` binary to production. Run it in CI on your test suite.

## Golden files

For outputs that are complex to assert inline (HTML, JSON blobs, compiled output), store the expected value in a file and compare:

```go
func TestRender(t *testing.T) {
    got := render(input)
    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        os.WriteFile(golden, []byte(got), 0644)
    }

    want, err := os.ReadFile(golden)
    if err != nil {
        t.Fatalf("reading golden file: %v", err)
    }
    if got != string(want) {
        t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, want)
    }
}

var update = flag.Bool("update", false, "update golden files")
```

Run `go test -update` when the output intentionally changes. The golden file shows in code review, so regressions are visible in diffs.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Idiomatic Patterns]] — table-driven tests, t.Helper, mock via interface patterns
- see:: [[Go Structs and Interfaces]] — mocking relies on interface definitions
- see:: [[Go Packages and Modules]] — test file conventions, build tags for integration tests

**Terms**
- t.Run, t.Parallel, t.Helper, t.Fatalf, table-driven test, fake struct, gomock, benchmark, b.ResetTimer, b.N, fuzz test, seed corpus, race detector, golden file
