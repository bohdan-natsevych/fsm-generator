## FSM Generator (Go) — Advanced Modulo Three Example

This project provides a small, generic finite state machine (FSM) library in Go, plus an example modulo-three FSM and a CLI.

- Library: `pkg/fsm`
- Example: `examples/mod3`
- CLI: `cmd/mod3`

### Requirements
- Go 1.22+

### Install and Build

```bash
# From repo root
go mod tidy

# Build CLI
go build -o bin/mod3 ./cmd/mod3

# Run with flag
./bin/mod3 -in 1111   # => 0

# Or via stdin
echo 1101 | ./bin/mod3   # => 1
```

### Library Overview

The library implements a generic deterministic FSM with a fluent builder.

```go
b := fsm.NewBuilder[string, byte](
	fsm.WithPreventOverwriteTransitions(),
)

b.AddState("S0", true).AddState("S1", true).AddState("S2", true)
b.SetInitial("S0")
b.AddSymbol('0').AddSymbol('1')
b.On("S0", '0', "S0").On("S0", '1', "S1")
b.On("S1", '0', "S2").On("S1", '1', "S0")
b.On("S2", '0', "S1").On("S2", '1', "S2")

m, err := b.Build()
```

Evaluate input:
```go
state, err := m.Eval([]byte("1110")) // => final state "S2"
```

Use the `Runner` to step manually:
```go
r := m.Start()
_ = r.Step('1')
_ = r.Step('1')
_ = r.Step('1')
_ = r.Step('0')
_ = r.State() // "S2"
```

### Mod-3 Example API

```go
rem, err := mod3.ModThree("1111") // => 0
```

### Notes
- Input is processed MSB first (left to right)

### Testing

```bash
go test ./...
```

