package fsm

// Options configure builder behavior such as determinism enforcement.

type buildOptions struct {
	enforceDeterministic        bool
	preventOverwriteTransitions bool
}

// Option mutates buildOptions when constructing a Builder.
type Option func(*buildOptions)

// WithDeterministic enforces that for each (state, symbol) there is at most one transition.
func WithDeterministic() Option {
	return func(o *buildOptions) { o.enforceDeterministic = true }
}

// WithPreventOverwriteTransitions panics if a transition is defined twice for the same (state, symbol).
func WithPreventOverwriteTransitions() Option {
	return func(o *buildOptions) { o.preventOverwriteTransitions = true }
}


