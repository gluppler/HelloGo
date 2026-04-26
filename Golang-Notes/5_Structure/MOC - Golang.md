---
tags:
  - type/structure
  - structure/moc
  - theme/golang
aliases:
  - Go MOC
lead: Navigation index for all Golang notes. Covers fundamentals, concurrency, standard library, and idiomatic patterns.
created: 2026-04-26
modified: 2026-04-26
---

# MOC — Golang

## Fundamentals

- [[Go Types and Variables]] — built-in types, zero values, type inference, constants
- [[Go Functions]] — multiple return values, named returns, variadic functions, defer
- [[Go Structs and Interfaces]] — struct embedding, interface satisfaction, DI via interfaces, compile-time verification
- [[Go Packages and Modules]] — module system, import paths, project layout, build tags, go.work
- [[Go Generics]] — type parameters, constraints, ~T approximation, generic data structures

## Concurrency

- [[Goroutines]] — lightweight threads, the Go scheduler, stack growth
- [[Channels]] — typed message passing, buffered vs unbuffered, direction constraints
- [[Select and Sync]] — select statement, sync.Mutex, sync.WaitGroup, sync.Once

## Standard library

- [[Go IO and Files]] — io.Reader, io.Writer, os package, bufio
- [[Go HTTP]] — net/http server and client, handlers, middleware patterns
- [[Go Error Handling]] — error interface, errors.Is, errors.As, wrapping

## Patterns

- [[Go Idiomatic Patterns]] — table-driven tests, functional options, embedding over inheritance
- [[Go Testing]] — subtests, t.Parallel, t.Helper, benchmarks, fuzzing, race detector

## AI and machine learning

- [[AI-ML Neural Network Foundations]] — neurons, activation functions, forward pass, backpropagation
- [[AI-ML Neural Network in Go]] — gonum matrices, forward pass, backprop, MNIST and iris examples
- [[AI-ML Go Data Science Tooling]] — gonum, gorgonia, golearn, gota, visualization, pipelines

---

# Back Matter

**Source**
- based_on:: [[Home]]

**References**
- see:: [[MOC - The Little Book of Deep Learning]] — parallel topic structure for comparison
