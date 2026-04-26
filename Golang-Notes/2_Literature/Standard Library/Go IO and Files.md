---
tags:
  - type/note
  - theme/golang
aliases: []
lead: IO in Go — the io.Reader and io.Writer interfaces, working with files via os, buffered IO with bufio, and the EOF convention.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024. The Go Programming Language Specification, go1.26."
---

# Go IO and files

## The core interfaces

Almost all IO in Go flows through two interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`Read` fills `p` with up to `len(p)` bytes. It returns how many bytes it read and an error. When there's no more data, it returns `io.EOF`. `Write` writes `p` and returns how many bytes were written. If `n < len(p)`, it must return a non-nil error.

These interfaces are implemented by files, network connections, HTTP request bodies, bytes.Buffer, gzip readers, and dozens of other things. A function that accepts `io.Reader` works on all of them.

```go
func (f *os.File) Read(buf []byte) (n int, err error)
func (f *os.File) Write(b []byte) (n int, err error)
```

## Reading a file

```go
f, err := os.Open("data.txt")
if err != nil {
    return err
}
defer f.Close()

buf := make([]byte, 1024)
for {
    n, err := f.Read(buf)
    if n > 0 {
        process(buf[:n])
    }
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
}
```

Check `n > 0` before checking the error. `Read` can return both data and `io.EOF` in the same call.

`defer f.Close()` right after the open is idiomatic — pairs the close with the open and runs no matter how the function returns.

## Writing a file

```go
f, err := os.Create("output.txt")
if err != nil {
    return err
}
defer f.Close()

fmt.Fprintf(f, "Hello, %s\n", name)
```

`os.Create` truncates if the file exists. Use `os.OpenFile` for more control:

```go
f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
```

## Buffered IO with bufio

Unbuffered reads hit the OS on every call. `bufio` wraps a reader or writer with a buffer, reducing syscalls:

```go
r := bufio.NewReader(f)
line, err := r.ReadString('\n')  // read up to newline
```

`bufio.Scanner` is cleaner for line-by-line reading:

```go
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    fmt.Println(scanner.Text())
}
if err := scanner.Err(); err != nil {
    return err
}
```

`bufio.Writer` buffers writes and flushes them in larger chunks. Always flush before the function returns:

```go
w := bufio.NewWriter(f)
fmt.Fprintf(w, "data\n")
if err := w.Flush(); err != nil {
    return err
}
```

## io utility functions

```go
io.Copy(dst, src)           // copy from reader to writer
io.ReadAll(r)               // read everything into []byte
io.ReadFull(r, buf)         // read exactly len(buf) bytes
io.LimitReader(r, n)        // read at most n bytes
io.MultiReader(r1, r2, ...) // concatenate readers
io.MultiWriter(w1, w2, ...) // write to multiple writers simultaneously
```

`io.Copy` is especially useful — it loops reading and writing until EOF, using an internal 32KB buffer.

## Standard streams

```go
os.Stdin   // io.Reader
os.Stdout  // io.Writer
os.Stderr  // io.Writer
```

These implement `io.Reader` and `io.Writer`, so they work everywhere those interfaces are accepted. `fmt.Fprintf(os.Stderr, ...)` is how you write to stderr without the log package's formatting.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Structs and Interfaces]] — io.Reader and io.Writer are the textbook example of Go interface design
- see:: [[Go HTTP]] — http.Request.Body is io.ReadCloser; ResponseWriter is io.Writer

**Terms**
- io.Reader, io.Writer, io.EOF, os.Open, os.Create, defer Close, bufio.Scanner, bufio.Writer, io.Copy, io.ReadAll
