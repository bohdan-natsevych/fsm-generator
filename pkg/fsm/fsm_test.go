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


