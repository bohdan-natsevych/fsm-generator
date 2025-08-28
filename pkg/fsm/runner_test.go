package fsm

import "testing"

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


