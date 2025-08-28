package fsm

import (
	"fmt"
)


// Machine is an immutable deterministic finite state machine.
// States and symbols are generic and must be comparable (hashable) to be used as map keys.
type Machine[S comparable, Sym comparable] struct {
	initialState S
	accepting    map[S]struct{}
	// transitions maps (state, symbol) -> next state
	transitions map[S]map[Sym]S
}

// Builder incrementally constructs a Machine.
type Builder[S comparable, Sym comparable] struct {
	states       map[S]struct{}
	symbols      map[Sym]struct{}
	initialSet   bool
	initialState S
	accepting    map[S]struct{}
	transitions  map[S]map[Sym]S
	options      buildOptions
}

// NewBuilder creates a new FSM builder.
func NewBuilder[S comparable, Sym comparable](opts ...Option) *Builder[S, Sym] {
	b := &Builder[S, Sym]{
		states:      make(map[S]struct{}),
		symbols:     make(map[Sym]struct{}),
		accepting:   make(map[S]struct{}),
		transitions: make(map[S]map[Sym]S),
	}
	for _, o := range opts {
		o(&b.options)
	}
	return b
}

// AddState registers a state. If isAccepting is true, it is added to the accepting set.
func (b *Builder[S, Sym]) AddState(state S, isAccepting bool) *Builder[S, Sym] {
	b.states[state] = struct{}{}
	if isAccepting {
		b.accepting[state] = struct{}{}
	}
	return b
}

// SetInitial sets the initial state. The state is implicitly registered.
func (b *Builder[S, Sym]) SetInitial(state S) *Builder[S, Sym] {
	b.initialSet = true
	b.initialState = state
	b.states[state] = struct{}{}
	return b
}

// AddSymbol registers an input symbol.
func (b *Builder[S, Sym]) AddSymbol(sym Sym) *Builder[S, Sym] {
	b.symbols[sym] = struct{}{}
	return b
}

// On adds a transition: from --sym--> to. States and symbol are implicitly registered.
func (b *Builder[S, Sym]) On(from S, sym Sym, to S) *Builder[S, Sym] {
	b.states[from] = struct{}{}
	b.states[to] = struct{}{}
	b.symbols[sym] = struct{}{}
	row := b.transitions[from]
	if row == nil {
		row = make(map[Sym]S)
		b.transitions[from] = row
	}
	if _, exists := row[sym]; exists && b.options.preventOverwriteTransitions {
		panic(fmt.Sprintf("transition already defined for (%v,%v)", from, sym))
	}
	row[sym] = to
	return b
}

// Build validates and returns an immutable Machine.
func (b *Builder[S, Sym]) Build() (*Machine[S, Sym], error) {
	verr := &ValidationErrors{}
	if !b.initialSet {
		verr.Append(newBuildError("initial state must be set"))
	}
	if len(b.states) == 0 {
		verr.Append(newBuildError("at least one state is required"))
	}
	if len(b.symbols) == 0 {
		verr.Append(newBuildError("at least one input symbol is required"))
	}

	// Ensure all transitions reference known states/symbols and optionally determinism.
	for from, row := range b.transitions {
		if _, ok := b.states[from]; !ok {
			verr.Append(newBuildError("transition from unknown state %v", from))
		}
		seen := make(map[Sym]struct{})
		for sym, to := range row {
			if _, ok := b.symbols[sym]; !ok {
				verr.Append(newBuildError("transition uses unknown symbol %v", sym))
			}
			if _, ok := b.states[to]; !ok {
				verr.Append(newBuildError("transition to unknown state %v", to))
			}
			if b.options.enforceDeterministic {
				if _, dup := seen[sym]; dup {
					verr.Append(newBuildError("multiple transitions from %v on %v", from, sym))
				} else {
					seen[sym] = struct{}{}
				}
			}
		}
	}

	if err := verr.AsError(); err != nil {
		return nil, err
	}

	// Copy into immutable machine.
	acc := make(map[S]struct{}, len(b.accepting))
	for s := range b.accepting {
		acc[s] = struct{}{}
	}
	trans := make(map[S]map[Sym]S, len(b.transitions))
	for from, row := range b.transitions {
		rowCopy := make(map[Sym]S, len(row))
		for sym, to := range row {
			rowCopy[sym] = to
		}
		trans[from] = rowCopy
	}
	return &Machine[S, Sym]{
		initialState: b.initialState,
		accepting:    acc,
		transitions:  trans,
	}, nil
}

// Start creates a new runner starting at the initial state.
func (m *Machine[S, Sym]) Start() *Runner[S, Sym] {
	return &Runner[S, Sym]{
		machine: m,
		state:   m.initialState,
	}
}

// Accepting reports whether the provided state is in the accepting set.
func (m *Machine[S, Sym]) Accepting(state S) bool {
	_, ok := m.accepting[state]
	return ok
}

// Eval consumes a sequence of symbols and returns the final state.
func (m *Machine[S, Sym]) Eval(input []Sym) (S, error) {
	r := m.Start()
	for _, sym := range input {
		if err := r.Step(sym); err != nil {
			var zero S
			return zero, err
		}
	}
	return r.State(), nil
}

// Runner is a mutable execution context for a Machine.
type Runner[S comparable, Sym comparable] struct {
	machine *Machine[S, Sym]
	state   S
}

// State returns the current state of the runner.
func (r *Runner[S, Sym]) State() S { return r.state }

// Step advances the machine using the provided input symbol.
func (r *Runner[S, Sym]) Step(sym Sym) error {
	row := r.machine.transitions[r.state]
	if row == nil {
		return &TransitionError{From: r.state, Symbol: sym}
	}
	next, ok := row[sym]
	if !ok {
		return &TransitionError{From: r.state, Symbol: sym}
	}
	r.state = next
	return nil
}


