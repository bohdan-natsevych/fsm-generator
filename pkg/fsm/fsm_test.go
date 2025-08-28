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

func TestTransitionErrorWhenMissing(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", true)
	b.AddSymbol('x')
	b.SetInitial("A")
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	_, err = m.Eval([]rune{'x'})
	if err == nil {
		t.Fatalf("expected transition error due to missing transition")
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

func TestRunnerStepSequence(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", true)
	b.SetInitial("A")
	b.AddSymbol('x')
	b.On("A", 'x', "B").On("B", 'x', "A")
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	r := m.Start()
	if r.State() != "A" {
		t.Fatalf("expected initial A, got %v", r.State())
	}
	if err := r.Step('x'); err != nil {
		t.Fatalf("unexpected step error: %v", err)
	}
	if r.State() != "B" {
		t.Fatalf("expected B after one step, got %v", r.State())
	}
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

func TestTransitionErrorMessageContainsDetails(t *testing.T) {
    b := NewBuilder[string, rune]()
    b.SetInitial("A")
    b.AddState("A", true)
    b.AddSymbol('y')
    m, err := b.Build()
    if err != nil {
        t.Fatalf("unexpected build error: %v", err)
    }
    _, err = m.Eval([]rune{'x'})
    if err == nil {
        t.Fatalf("expected error")
    }
    msg := err.Error()
    if msg == "" || msg == "<nil>" {
        t.Fatalf("expected non-empty error message")
    }
    if msg == "no validation errors" { // sanity check not mixing error types
        t.Fatalf("unexpected validation error string")
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

func TestStepMissingRowReturnsTransitionError(t *testing.T) {
	b := NewBuilder[string, rune]()
	b.AddState("A", true).AddState("B", false)
	b.SetInitial("A")
	b.AddSymbol('x')
	// No transition from A, only from B
	b.On("B", 'x', "A")
	m, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	r := m.Start()
	if err := r.Step('x'); err == nil {
		t.Fatalf("expected transition error when stepping from state with no row")
	}
}

func TestValidationErrorsFormatting(t *testing.T) {
	ve := &ValidationErrors{}
	if got := ve.Error(); got != "no validation errors" {
		t.Fatalf("empty ValidationErrors message mismatch: %q", got)
	}
	ve.Append(newBuildError("a"))
	if got := ve.Error(); got != "a" {
		t.Fatalf("single ValidationError message mismatch: %q", got)
	}
	ve.Append(newBuildError("b"))
	msg := ve.Error()
	if msg == "a" || msg == "no validation errors" {
		t.Fatalf("expected multi-error message, got %q", msg)
	}
}

//

func TestValidationAsErrorAndIsEmpty(t *testing.T) {
    ve := &ValidationErrors{}
    if !ve.IsEmpty() {
        t.Fatalf("expected empty at start")
    }
    if ve.AsError() != nil {
        t.Fatalf("expected nil error for empty ValidationErrors")
    }
    ve.Append(newBuildError("x"))
    if ve.IsEmpty() {
        t.Fatalf("expected non-empty after append")
    }
    if ve.AsError() == nil {
        t.Fatalf("expected non-nil error after append")
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

func TestStepRowExistsButSymbolMissing(t *testing.T) {
    b := NewBuilder[string, rune]()
    b.AddState("A", true).AddState("B", false)
    b.SetInitial("A")
    b.AddSymbol('x').AddSymbol('y')
    b.On("A", 'y', "B") // row exists for A, but no transition on 'x'
    m, err := b.Build()
    if err != nil {
        t.Fatalf("unexpected build error: %v", err)
    }
    r := m.Start()
    if err := r.Step('x'); err == nil {
        t.Fatalf("expected transition error when missing symbol in existing row")
    }
}

func TestBuildTransitionFromUnknownState(t *testing.T) {
    b := NewBuilder[string, rune]()
    b.AddState("A", true)
    b.SetInitial("A")
    b.AddSymbol('x')
    // Inject transition from unknown state "X"
    if b.transitions["X"] == nil {
        b.transitions["X"] = make(map[rune]string)
    }
    b.transitions["X"]['x'] = "A"
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
    if b.transitions["A"] == nil {
        b.transitions["A"] = make(map[rune]string)
    }
    b.transitions["A"]['z'] = "A"
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
    if b.transitions["A"] == nil {
        b.transitions["A"] = make(map[rune]string)
    }
    b.transitions["A"]['x'] = "Z"
    if _, err := b.Build(); err == nil {
        t.Fatalf("expected error for transition to unknown state")
    }
}

func TestTransitionErrorMessage(t *testing.T) {
    b := NewBuilder[string, rune]()
    b.AddState("A", true)
    b.SetInitial("A")
    b.AddSymbol('x')
    m, err := b.Build()
    if err != nil {
        t.Fatalf("unexpected build error: %v", err)
    }
    if _, err := m.Eval([]rune{'x'}); err == nil {
        t.Fatalf("expected transition error")
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


