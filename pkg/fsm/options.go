package fsm

// Options configure builder behavior.

type buildOptions struct {
	preventOverwriteTransitions bool
}

// Option mutates buildOptions when constructing a Builder.
type Option func(*buildOptions)

// WithPreventOverwriteTransitions panics if a transition is defined twice for the same (state, symbol).
func WithPreventOverwriteTransitions() Option {
	return func(o *buildOptions) { o.preventOverwriteTransitions = true }
}


