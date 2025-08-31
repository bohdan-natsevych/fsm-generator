package fsm

// TransitionKey represents a state-symbol pair for transition lookup
type TransitionKey[S, Sym comparable] struct {
	From   S
	Symbol Sym
}

// Machine is an immutable deterministic finite state machine.
// States and symbols are generic and must be comparable (hashable) to be used as map keys.
type Machine[S comparable, Sym comparable] struct {
	initialState S
	accepting    map[S]struct{}
	// Flat map with composite key for O(1) lookup
	transitions map[TransitionKey[S, Sym]]S
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

// Get all states in the machine
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
	for key, to := range m.transitions {
		// Add 'from' state
		if _, exists := seen[key.From]; !exists {
			states = append(states, key.From)
			seen[key.From] = struct{}{}
		}
		// Add 'to' state
		if _, exists := seen[to]; !exists {
			states = append(states, to)
			seen[to] = struct{}{}
		}
	}

	return states
}

// Get the initial state
func (m *Machine[S, Sym]) InitialState() S {
	return m.initialState
}

// GetTransition returns the target state for a transition, if it exists
func (m *Machine[S, Sym]) GetTransition(from S, symbol Sym) (S, bool) {
	to, ok := m.transitions[TransitionKey[S, Sym]{From: from, Symbol: symbol}]
	return to, ok
}

// HasTransition reports whether a transition exists from the given state on the given symbol
func (m *Machine[S, Sym]) HasTransition(from S, symbol Sym) bool {
	_, exists := m.GetTransition(from, symbol)
	return exists
}


