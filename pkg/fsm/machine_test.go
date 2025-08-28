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


