package fsm

import "testing"

func TestMachineEvalMod3States(t *testing.T) {
	b := NewBuilder[string, rune](WithPreventOverwriteTransitions())
	b.AddState("S0", true).AddState("S1", true).AddState("S2", true)
	b.SetInitial("S0")
	b.AddSymbol('0').AddSymbol('1')
	b.On("S0", '0', "S0").On("S0", '1', "S1")
	b.On("S1", '0', "S2").On("S1", '1', "S0")
	b.On("S2", '0', "S1").On("S2", '1', "S2")

	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	final, err := m.Eval([]rune("1110"))
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	if final != "S2" {
		t.Fatalf("expected final state S2, got %v", final)
	}
}

func TestEvalEmptyReturnsInitial(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.SetInitial("I")
	b.AddSymbol('x')
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	s, err := m.Eval(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "I" {
		t.Fatalf("expected initial state 'I', got %v", s)
	}
}

func TestEvalReturnsZeroOnError(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true)
	b.SetInitial("A")
	b.AddSymbol('y')
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	s, err := m.Eval([]rune{'x'}) // 'x' not in alphabet
	if err == nil {
		t.Fatalf("expected error from Eval on unknown symbol")
	}
	if s != "" { // zero value for string state type
		t.Fatalf("expected zero value state on error, got %q", s)
	}
}

func TestEvalAccepting(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", false)
	b.SetInitial("A")
	b.AddSymbol('x')
	b.On("A", 'x', "B").On("B", 'x', "A")
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	// Even number of steps should end in accepting state A
	accepting, err := m.EvalAccepting([]rune{})
	if err != nil || !accepting {
		t.Fatalf("expected accepting=true for empty input, got %v, err: %v", accepting, err)
	}

	// Odd number of steps should end in non-accepting state B
	accepting, err = m.EvalAccepting([]rune{'x'})
	if err != nil || accepting {
		t.Fatalf("expected accepting=false for single step, got %v, err: %v", accepting, err)
	}

	// Test error case - invalid symbol should return error
	accepting, err = m.EvalAccepting([]rune{'z'}) // 'z' not in alphabet
	if err == nil {
		t.Fatalf("expected error from EvalAccepting on unknown symbol")
	}
	if accepting {
		t.Fatalf("expected accepting=false when error occurs")
	}
}

func TestStatesMethod(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", false).AddState("C", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	b.On("A", 'x', "B").On("B", 'x', "C")
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	states := m.States()
	if len(states) != 3 {
		t.Fatalf("expected 3 states, got %d: %v", len(states), states)
	}

	// Check all expected states are present
	stateSet := make(map[string]struct{})
	for _, s := range states {
		stateSet[s] = struct{}{}
	}
	for _, expected := range []string{"A", "B", "C"} {
		if _, ok := stateSet[expected]; !ok {
			t.Fatalf("expected state %q not found in states: %v", expected, states)
		}
	}
}

func TestStatesMethodEdgeCases(t *testing.T) {
	// CURSOR: Test case where initial state is also accepting and appears in transitions
	b := NewBuilder[string, rune]()
	b.SetInitial("S") // Initial state that's also accepting
	b.AddState("S", true)
	b.AddSymbol('x')
	b.On("S", 'x', "S") // Self-transition
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	states := m.States()
	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d: %v", len(states), states)
	}
	if states[0] != "S" {
		t.Fatalf("expected state S, got %v", states[0])
	}

	// CURSOR: Test case with states only referenced in transitions
	b2 := NewBuilder[string, rune]()
	b2.SetInitial("Init")
	b2.AddSymbol('x')
	b2.On("Init", 'x', "TransitionOnly") // TransitionOnly not explicitly added
	m2, err := b2.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	states2 := m2.States()
	if len(states2) != 2 {
		t.Fatalf("expected 2 states, got %d: %v", len(states2), states2)
	}
	stateSet := make(map[string]struct{})
	for _, s := range states2 {
		stateSet[s] = struct{}{}
	}
	for _, expected := range []string{"Init", "TransitionOnly"} {
		if _, ok := stateSet[expected]; !ok {
			t.Fatalf("expected state %q not found in states: %v", expected, states2)
		}
	}
}

func TestInitialStateMethod(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("INIT", true)
	b.SetInitial("INIT")
	b.AddSymbol('x')
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if initial := m.InitialState(); initial != "INIT" {
		t.Fatalf("expected initial state INIT, got %v", initial)
	}
}

func TestGetTransitionMethod(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", false)
	b.SetInitial("A")
	b.AddSymbol('x').AddSymbol('y')
	b.On("A", 'x', "B")
	// No transition from A on 'y'
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	// Test existing transition
	to, exists := m.GetTransition("A", 'x')
	if !exists {
		t.Fatalf("expected transition from A on x to exist")
	}
	if to != "B" {
		t.Fatalf("expected transition from A on x to go to B, got %v", to)
	}

	// Test non-existing transition
	_, exists = m.GetTransition("A", 'y')
	if exists {
		t.Fatalf("expected no transition from A on y")
	}
}

func TestHasTransitionMethod(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", false)
	b.SetInitial("A")
	b.AddSymbol('x').AddSymbol('y')
	b.On("A", 'x', "B")
	// No transition from A on 'y'
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	// Test existing transition
	if !m.HasTransition("A", 'x') {
		t.Fatalf("expected transition from A on x to exist")
	}

	// Test non-existing transition
	if m.HasTransition("A", 'y') {
		t.Fatalf("expected no transition from A on y")
	}
}


