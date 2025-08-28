package fsm

// Options configure builder behavior.

type buildOptions struct {
	preventOverwriteTransitions bool
	requireTotalTransitions      bool
	requireAtLeastOneAccepting   bool
	errorOnUnreachableStates     bool
	errorWhenNoAcceptingReachable bool
}

// Option mutates buildOptions when constructing a Builder.
type Option func(*buildOptions)

// WithPreventOverwriteTransitions panics if a transition is defined twice for the same (state, symbol).
func WithPreventOverwriteTransitions() Option {
	return func(o *buildOptions) { o.preventOverwriteTransitions = true }
}

// WithRequireTotalTransitions enforces that δ is total: every (state, symbol) has a transition.
func WithRequireTotalTransitions() Option {
	return func(o *buildOptions) { o.requireTotalTransitions = true }
}

// WithRequireAtLeastOneAccepting enforces F is non-empty.
func WithRequireAtLeastOneAccepting() Option {
	return func(o *buildOptions) { o.requireAtLeastOneAccepting = true }
}

// WithErrorOnUnreachableStates fails build if any state is unreachable from q0.
func WithErrorOnUnreachableStates() Option {
	return func(o *buildOptions) { o.errorOnUnreachableStates = true }
}

// WithErrorWhenNoAcceptingReachable fails build if no accepting state is reachable from q0.
func WithErrorWhenNoAcceptingReachable() Option {
	return func(o *buildOptions) { o.errorWhenNoAcceptingReachable = true }
}


