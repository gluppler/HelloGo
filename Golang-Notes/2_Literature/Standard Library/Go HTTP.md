---
tags:
  - type/note
  - theme/golang
aliases: []
lead: net/http in Go — the Handler interface, HandlerFunc adapter, ServeMux routing, the HTTP client, and composing middleware with function wrapping.
created: 2026-04-26
modified: 2026-04-26
source: "Effective Go, 2024."
---

# Go HTTP

## The Handler interface

The entire net/http server is built on one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type that implements `ServeHTTP` can serve HTTP requests. That's it. The rest of the package builds utilities on top of this.

`http.ResponseWriter` is itself an interface. It wraps `io.Writer`, so you can pass it to `fmt.Fprintf` directly:

```go
type Counter struct{ n int }

func (c *Counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c.n++
    fmt.Fprintf(w, "count: %d\n", c.n)
}
```

## HandlerFunc

Writing a new type just to implement `ServeHTTP` is verbose when you have a standalone function. `http.HandlerFunc` is a type that turns a function with the right signature into a `Handler`:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, req *Request) {
    f(w, req)
}
```

So you can convert any compatible function:

```go
func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello")
}

http.Handle("/hello", http.HandlerFunc(hello))
```

Or use `http.HandleFunc`, which does the conversion for you:

```go
http.HandleFunc("/hello", hello)
```

## ServeMux

`http.ServeMux` routes requests to handlers by URL pattern. The default mux (`http.DefaultServeMux`) is used by `http.HandleFunc` and `http.Handle` when no mux is specified.

For production code, create your own mux to avoid conflicts with other packages that register handlers on the default:

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/users", usersHandler)
mux.HandleFunc("/api/posts", postsHandler)
mux.Handle("/static/", http.FileServer(http.Dir("./static")))

http.ListenAndServe(":8080", mux)
```

Go 1.22 added richer pattern matching — you can match on method and path segments:

```go
mux.HandleFunc("GET /api/users/{id}", getUserHandler)
mux.HandleFunc("POST /api/users", createUserHandler)
```

## Middleware

Middleware wraps a handler to add behavior — logging, auth, rate limiting. The pattern is a function that takes a handler and returns a handler:

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

Chain multiple middleware by nesting:

```go
handler := logging(auth(rateLimiter(mux)))
http.ListenAndServe(":8080", handler)
```

## HTTP client

```go
resp, err := http.Get("https://api.example.com/data")
if err != nil {
    return err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
```

Always close `resp.Body` — if you don't, the connection can't be reused and you'll leak file descriptors.

For anything beyond a quick GET, use a custom client with timeouts. `http.DefaultClient` has no timeout, which is almost never what you want in production:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}

req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")

resp, err := client.Do(req)
```

`http.NewRequestWithContext` wires the context into the request so cancellations and timeouts propagate properly.

## Reading request data

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // query parameters
    name := r.URL.Query().Get("name")

    // path values (Go 1.22+)
    id := r.PathValue("id")

    // JSON body
    var payload struct{ Name string }
    json.NewDecoder(r.Body).Decode(&payload)

    // form values (after parsing)
    r.ParseForm()
    val := r.FormValue("field")
}
```

## Writing responses

```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)  // must be called before Write
    json.NewEncoder(w).Encode(result)
}
```

`WriteHeader` sets the status code. If you call `Write` without calling `WriteHeader` first, the status defaults to 200. Once you write anything to `w`, the headers are flushed and you can't change them.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Structs and Interfaces]] — Handler interface is a textbook Go interface
- see:: [[Go IO and Files]] — ResponseWriter is io.Writer; request body is io.Reader
- see:: [[Go Idiomatic Patterns]] — middleware uses the function adapter pattern
- see:: [[Select and Sync]] — use context.WithTimeout for HTTP client calls

**Terms**
- http.Handler, http.HandlerFunc, http.ServeMux, middleware, http.Client, http.ResponseWriter, WriteHeader, NewRequestWithContext
