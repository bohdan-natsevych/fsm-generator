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
		row = make(map[Sym]S, len(b.symbols))
		b.transitions[from] = row
	}
	if _, exists := row[sym]; exists && b.options.preventOverwriteTransitions {
		panic(fmt.Sprintf("transition already defined for (%v,%v)", from, sym))
	}
	row[sym] = to
	return b
}

// CURSOR: Optional checks are extracted to helpers to keep Build concise.
func (b *Builder[S, Sym]) checkRequireTotalTransitions(verr *ValidationErrors) {
    if !b.options.requireTotalTransitions {
        return
    }
    for s := range b.states {
        row := b.transitions[s]
        for sym := range b.symbols {
            if row == nil {
                verr.Append(newBuildError("missing transitions for state %v", s))
                break
            }
            if _, ok := row[sym]; !ok {
                verr.Append(newBuildError("missing transition from %v on %v", s, sym))
            }
        }
    }
}

func (b *Builder[S, Sym]) checkRequireAtLeastOneAccepting(verr *ValidationErrors) {
    if b.options.requireAtLeastOneAccepting && len(b.accepting) == 0 {
        verr.Append(newBuildError("at least one accepting state required"))
    }
}

func (b *Builder[S, Sym]) checkReachability(verr *ValidationErrors) {
    if !b.initialSet || !(b.options.errorOnUnreachableStates || b.options.errorWhenNoAcceptingReachable) {
        return
    }
    reached := make(map[S]struct{})
    queue := []S{b.initialState}
    reached[b.initialState] = struct{}{}
    // CURSOR: Use index-based queue to avoid retaining backing array when slicing
    for i := 0; i < len(queue); i++ {
        cur := queue[i]
        for _, to := range b.transitions[cur] {
            if _, ok := reached[to]; !ok {
                reached[to] = struct{}{}
                queue = append(queue, to)
            }
        }
    }
    if b.options.errorOnUnreachableStates {
        for s := range b.states {
            if _, ok := reached[s]; !ok {
                verr.Append(newBuildError("unreachable state %v", s))
            }
        }
    }
    if b.options.errorWhenNoAcceptingReachable {
        any := false
        for s := range b.accepting {
            if _, ok := reached[s]; ok {
                any = true
                break
            }
        }
        if !any {
            verr.Append(newBuildError("no accepting state reachable from initial"))
        }
    }
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

	// Ensure F ⊆ Q: every accepting state must be a registered state
	for s := range b.accepting {
		if _, ok := b.states[s]; !ok {
			verr.Append(newBuildError("accepting state unknown %v", s))
		}
	}

	// Ensure all transitions reference known states/symbols.
	for from, row := range b.transitions {
		if _, ok := b.states[from]; !ok {
			verr.Append(newBuildError("transition from unknown state %v", from))
		}
		for sym, to := range row {
			if _, ok := b.symbols[sym]; !ok {
				verr.Append(newBuildError("transition uses unknown symbol %v", sym))
			}
			if _, ok := b.states[to]; !ok {
				verr.Append(newBuildError("transition to unknown state %v", to))
			}
		}
	}

	// Optional checks controlled by flags
	b.checkRequireTotalTransitions(verr)
	b.checkRequireAtLeastOneAccepting(verr)
	b.checkReachability(verr)

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

// Convenience method for checking if final state after evaluation is accepting
func (m *Machine[S, Sym]) EvalAccepting(input []Sym) (bool, error) {
	finalState, err := m.Eval(input)
	if err != nil {
		return false, err
	}
	return m.Accepting(finalState), nil
}

// CURSOR: Get all states in the machine
func (m *Machine[S, Sym]) States() []S {
	states := make([]S, 0, len(m.accepting)+1)
	seen := make(map[S]struct{})
	
	// Add initial state first
	states = append(states, m.initialState)
	seen[m.initialState] = struct{}{}
	
	// Add accepting states
	for state := range m.accepting {
		if _, exists := seen[state]; !exists {
			states = append(states, state)
			seen[state] = struct{}{}
		}
	}
	
	// Add any other states from transitions
	for from := range m.transitions {
		if _, exists := seen[from]; !exists {
			states = append(states, from)
			seen[from] = struct{}{}
		}
		for _, to := range m.transitions[from] {
			if _, exists := seen[to]; !exists {
				states = append(states, to)
				seen[to] = struct{}{}
			}
		}
	}
	
	return states
}

// Get the initial state
func (m *Machine[S, Sym]) InitialState() S {
	return m.initialState
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


