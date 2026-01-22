# Testing Guide

## File format
```go
package <name>_test

import (
  "testing"

  "github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/<name>"
)

func Test<StructName>_<FuncName>(t *testing.T) {
  t.Parallel()

  tests := []struct{
    name string
    <variables ...>
    expected string // Or other types
  } {
    {
      name: "A name for case 1",
      <variable definition ...>,
      expected: "value",
    },
    { case 2 ...},
    ...
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      t.Parallel()

      got := tt.variable.func_name()
      if got != tt.expected {
        t.Fatalf("String() mismatch: \ngot: %s\nwant: %s", got, tt.expected)
      }
    })
  }
}
```

## Adding some Benchmark for important informations
```go
package server

import (
	"testing"
)

func Benchmark<func_Name>(b *testing.B) {
  // setup

	b.ResetTimer()
	for b.Loop() {
    // func call
	}
}
```

## Running tests
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/package
# e.g.
go test ./internal/protocol/identity

# Run tests with coverage
go test -cover ./internal/protocol/identity

# Generate detailed coverage report
go test -coverprofile=coverage.out ./internal/protocol/identity
go tool cover -func=coverage.out

# Run with race detector
go test -race ./...

# Run Benchmark
go test -bench=. -benchmem ./...
```
