package fsm

import (
	"testing"
)

func TestBuildRequiresInitialAndSymbolsStates(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("S0", true)
	// No initial, no symbols
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error when building without initial or symbols")
	}
}

func TestAcceptingMustBeSubsetOfStates(t *testing.T) {
	b := NewBuilder[string, rune]()

	b.AddState("Known", true)
	b.SetInitial("Known")
	b.AddSymbol('x')
	// Corrupt internal state for test purposes
	delete(b.states, "Known")
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error due to accepting state not in states")
	}
}

func TestPreventOverwriteTransitionsPanics(t *testing.T) {
	b := NewBuilder[string, rune](WithPreventOverwriteTransitions())
	b.AddState("A", true).AddState("B", true)
	b.AddSymbol('x')
	b.SetInitial("A")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate transition definition")
		}
	}()

	b.On("A", 'x', "B")
	b.On("A", 'x', "A") // duplicate should panic
}

func TestAddStateIdempotentAndAccepting(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("S", false)
	b.AddState("S", true)  // becomes accepting
	b.AddState("S", false) // should not unset accepting
	b.SetInitial("S")
	b.AddSymbol('x')
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if !m.Accepting("S") {
		t.Fatalf("expected S to be accepting after toggling")
	}
}

func TestInitialImplicitRegistration(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.SetInitial("I") // not added via AddState
	b.AddSymbol('x')
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if m.Accepting("I") { // never marked accepting
		t.Fatalf("initial should not be accepting unless specified")
	}
}

func TestRequireTotalTransitions(t *testing.T) {
	b := NewBuilder[string, rune](WithRequireTotalTransitions())
	b.SetInitial("S0")
	b.AddState("S0", true).AddState("S1", false)
	b.AddSymbol('0').AddSymbol('1')
	b.On("S0", '0', "S1") // missing S0 on '1' and entire row for S1
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error due to missing total transitions")
	}
}

func TestRequireAtLeastOneAccepting(t *testing.T) {
	b := NewBuilder[string, rune](WithRequireAtLeastOneAccepting())
	b.SetInitial("S0")
	b.AddState("S0", false)
	b.AddSymbol('x')
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error requiring at least one accepting state")
	}
}

func TestErrorOnUnreachableStates(t *testing.T) {
	b := NewBuilder[string, rune](WithErrorOnUnreachableStates())
	b.SetInitial("A")
	b.AddState("A", true).AddState("B", false)
	b.AddSymbol('x')
	// No transitions to B, so B is unreachable
	b.On("A", 'x', "A")
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error for unreachable state B")
	}
}

func TestErrorWhenNoAcceptingReachable(t *testing.T) {
	b := NewBuilder[string, rune](WithErrorWhenNoAcceptingReachable())
	b.SetInitial("A")
	b.AddState("A", false).AddState("B", true)
	b.AddSymbol('x')
	// Transitions form a self-loop on A; B is accepting but unreachable
	b.On("A", 'x', "A")
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error when no accepting state is reachable")
	}
}

func TestOnImplicitlyRegisters(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.SetInitial("A")
	b.On("A", 'x', "B") // registers B and 'x'
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	r := m.Start()
	if err := r.Step('x'); err != nil {
		t.Fatalf("unexpected step error: %v", err)
	}
}

func TestBuildTransitionFromUnknownState(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	// Inject transition from unknown state "X"
	key := TransitionKey[string, rune]{From: "X", Symbol: 'x'}
	b.transitions[key] = "A"
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error for transition from unknown state")
	}
}

func TestBuildTransitionUsesUnknownSymbol(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	// Inject transition on unknown symbol 'z'
	key := TransitionKey[string, rune]{From: "A", Symbol: 'z'}
	b.transitions[key] = "A"
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error for transition with unknown symbol")
	}
}

func TestBuildTransitionToUnknownState(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	// Inject transition to unknown state "Z"
	key := TransitionKey[string, rune]{From: "A", Symbol: 'x'}
	b.transitions[key] = "Z"
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error for transition to unknown state")
	}
}

func TestOverwriteTransitionWhenAllowed(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	b.On("A", 'x', "B")
	b.On("A", 'x', "A") // overwrite allowed
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	r := m.Start()
	if err := r.Step('x'); err != nil {
		t.Fatalf("unexpected step error: %v", err)
	}
	if r.State() != "A" {
		t.Fatalf("expected overwritten transition to go to A, got %v", r.State())
	}
}

func TestStartReturnsInitialAndAccepting(t *testing.T) {
    b := NewBuilder[string, rune]()
    b.AddState("Init", true).AddState("Other", false)
    b.SetInitial("Init")
    b.AddSymbol('x')
    b.On("Init", 'x', "Other")
    m, err := b.Build()
    if err != nil {
        t.Fatalf("unexpected build error: %v", err)
    }
    r := m.Start()
    if r.State() != "Init" {
        t.Fatalf("expected initial state 'Init', got %v", r.State())
    }
    if !m.Accepting("Init") {
        t.Fatalf("expected 'Init' to be accepting")
    }
    if m.Accepting("Unknown") {
        t.Fatalf("did not expect unknown state to be accepting")
    }
}


