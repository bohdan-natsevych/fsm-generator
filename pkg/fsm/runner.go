package fsm

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


