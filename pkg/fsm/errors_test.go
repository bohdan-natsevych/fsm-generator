package fsm

import "testing"

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

func TestValidationAppendNilError(t *testing.T) {
	ve := &ValidationErrors{}
	ve.Append(nil) // Should be handled gracefully
	if !ve.IsEmpty() {
		t.Fatalf("expected ValidationErrors to remain empty after appending nil")
	}
}


