package fsm

// Machine is an immutable deterministic finite state machine.
// States and symbols are generic and must be comparable (hashable) to be used as map keys.
type Machine[S comparable, Sym comparable] struct {
	initialState S
	accepting    map[S]struct{}
	// transitions maps (state, symbol) -> next state
	transitions map[S]map[Sym]S
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


