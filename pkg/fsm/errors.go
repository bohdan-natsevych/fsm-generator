package fsm

import (
	"errors"
	"fmt"
)

type BuildError struct {
	message string
}

func (e *BuildError) Error() string { return e.message }

func newBuildError(format string, args ...any) error {
	return &BuildError{message: fmt.Sprintf(format, args...)}
}

type ValidationErrors struct {
	errors []error
}

func (ve *ValidationErrors) Error() string {
	if len(ve.errors) == 0 {
		return "no validation errors"
	}
	if len(ve.errors) == 1 {
		return ve.errors[0].Error()
	}
	msg := "validation errors:"
	for _, err := range ve.errors {
		msg += "\n - " + err.Error()
	}
	return msg
}

func (ve *ValidationErrors) Append(err error) {
	if err == nil {
		return
	}
	ve.errors = append(ve.errors, err)
}

func (ve *ValidationErrors) IsEmpty() bool { return len(ve.errors) == 0 }

func (ve *ValidationErrors) AsError() error {
	if ve.IsEmpty() {
		return nil
	}
	return ve
}

type TransitionError struct {
	From   any
	Symbol any
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("no transition from %v on %v", e.From, e.Symbol)
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}


