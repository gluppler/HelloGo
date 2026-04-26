---
tags:
  - type/note
  - theme/golang
aliases: []
lead: Go's module system — go.mod, import paths, exported identifiers, package naming conventions, init functions, and blank imports.
created: 2026-04-26
modified: 2026-04-26
source: "How to Write Go Code, go.dev. Effective Go, 2024."
---

# Go packages and modules

## Code organization

A **package** is a directory of `.go` files compiled together. Everything declared in one file is visible to all other files in the same package — no explicit intra-package imports.

A **module** is a collection of related packages released together. It lives at a root directory containing a `go.mod` file, which declares the module path and the minimum Go version:

```
module github.com/you/myproject

go 1.22
```

The module path doubles as the import path prefix. A package at `myproject/utils/` is imported as `github.com/you/myproject/utils`.

You don't need to publish to a remote repository before building. A local module works fine.

## First program setup

```bash
mkdir hello && cd hello
go mod init example/user/hello
```

Then write `hello.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, world.")
}
```

Build and install:

```bash
go install example/user/hello
```

The binary lands in `$GOPATH/bin` (or `$HOME/go/bin` if GOPATH isn't set). Override with `go env -w GOBIN=/somewhere/else/bin`.

## Exported identifiers

Capitalization determines visibility. An identifier starting with an uppercase letter is exported — visible to other packages. Lowercase is unexported — package-private.

```go
func ExportedFunc() {}  // visible outside
func internalFunc() {}  // package-only
```

This applies to types, fields, constants, variables — everything. There's no `public`/`private` keyword; the case of the first letter does it.

## Package naming

Package names should be short, lowercase, single words. No underscores, no mixedCaps.

```
bufio, not buf_io or bufIO
strconv, not stringConversion
```

The package name is the default name at import. Since callers prefix everything with it, don't repeat the package name in exported identifiers:

```go
// bad — bufio.BufReader
// good — bufio.Reader, used as bufio.Reader
```

One-method interfaces get the method name plus `-er`: `io.Reader`, `io.Writer`, `fmt.Stringer`.

## Importing packages

```go
import (
    "fmt"
    "os"
    "github.com/google/go-cmp/cmp"
)
```

Go resolves external modules via `go mod tidy`, which downloads missing dependencies and updates `go.mod` and `go.sum`.

```bash
go mod tidy
```

`go.sum` records the cryptographic hash of every dependency version. Commit both files.

## Blank import (side-effect import)

Sometimes you want a package only for its `init` function — to register a driver, handler, or codec — without using any of its exported names:

```go
import _ "net/http/pprof"   // registers pprof HTTP handlers
import _ "image/png"        // registers PNG decoder
```

The `_` makes explicit that you're importing for the side effect only.

## The init function

Each source file can define one or more `init` functions. They run after all package-level variable initializations, in the order files are presented to the compiler, after all imported packages have initialized.

`init` is the right place for setup that can't be expressed as a variable initializer:

```go
func init() {
    if user == "" {
        log.Fatal("$USER not set")
    }
    if gopath == "" {
        gopath = home + "/go"
    }
}
```

You can't call `init` directly. It's called automatically.

## Testing

Test files end in `_test.go`. Test functions are `TestXxx(t *testing.T)`:

```go
package morestrings

import "testing"

func TestReverseRunes(t *testing.T) {
    cases := []struct{ in, want string }{
        {"Hello, world", "dlrow ,olleH"},
        {"", ""},
    }
    for _, c := range cases {
        got := ReverseRunes(c.in)
        if got != c.want {
            t.Errorf("ReverseRunes(%q) = %q, want %q", c.in, got, c.want)
        }
    }
}
```

Run with `go test`. The table-driven structure here is the standard pattern — see [[Go Idiomatic Patterns]].

## Project layout

The `cmd/`, `internal/`, and `pkg/` split is the de facto standard for anything beyond a single binary:

```
myproject/
├── cmd/
│   ├── server/main.go   # entry point for the HTTP server
│   └── migrate/main.go  # entry point for DB migrations
├── internal/
│   ├── auth/            # private to this module
│   └── db/
├── pkg/
│   └── middleware/      # importable by external projects
├── go.mod
└── go.sum
```

`internal/` is enforced by the compiler — code outside the module cannot import it. `pkg/` has no enforcement; the name is a convention meaning "library code that's OK for others to import."

Each binary lives in its own subdirectory under `cmd/` with its own `main.go`. Don't put business logic in `main.go`. Pull it into `internal/` and call it from main.

## Build tags

Build tags control which files are compiled. The modern syntax (Go 1.17+):

```go
//go:build linux && amd64

package myapp
```

The constraint must appear before the package declaration, with a blank line separating them. Multiple constraints combine with `&&` (and), `||` (or), `!` (not).

Common uses: platform-specific code, integration tests that shouldn't run in CI:

```go
//go:build integration

package tests
```

Run integration tests explicitly: `go test -tags integration ./...`

## go.work — multi-module workspaces

When you're working across multiple modules locally (e.g., a library and an app that uses it), `go.work` tells the toolchain to use local versions instead of fetching from a registry:

```
go 1.22

use (
    ./myapp
    ./mylib
)
```

Initialize with `go work init ./myapp ./mylib`. Add modules with `go work use ./newmod`.

`go.work` is a development tool. Don't commit it to repos that will be built by others — it overrides module resolution in ways that may not make sense outside your machine.

## Cross-compilation

Go compiles for any supported target from any host. Set `GOOS` and `GOARCH`:

```bash
GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o bin/app-darwin-arm64 ./cmd/server
GOOS=windows GOARCH=amd64 go build -o bin/app.exe ./cmd/server
```

No cross-compilation toolchain needed unless you're using cgo. Pure Go code just works.

## go generate

`//go:generate` embeds shell commands in source files that `go generate` runs when you invoke it:

```go
//go:generate stringer -type=Direction
//go:generate mockgen -source=service.go -destination=mocks/service_mock.go
```

Run with `go generate ./...`. It's just a convention for code generation — the command can be anything. Track tool dependencies in a `tools.go` with a build tag so they don't bloat your production binary:

```go
//go:build tools

package tools

import (
    _ "golang.org/x/tools/cmd/stringer"
    _ "github.com/golang/mock/mockgen"
)
```

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[Go Types and Variables]] — exported vs unexported follows from capitalization
- see:: [[Go Idiomatic Patterns]] — table-driven tests shown here are standard Go
- see:: [[Go Testing]] — full testing reference

**Terms**
- go.mod, go.sum, module path, package, exported identifier, init function, blank import, go mod tidy, internal package, build tag, go.work, cross-compilation, go generate
