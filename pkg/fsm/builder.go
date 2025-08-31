package fsm

import (
	"fmt"
)

// Builder incrementally constructs a Machine.
type Builder[S comparable, Sym comparable] struct {
	states       map[S]struct{}
	symbols      map[Sym]struct{}
	initialSet   bool
	initialState S
	accepting    map[S]struct{}
	transitions  map[TransitionKey[S, Sym]]S
	options      buildOptions
}

// NewBuilder creates a new FSM builder.
func NewBuilder[S comparable, Sym comparable](opts ...Option) *Builder[S, Sym] {
	b := &Builder[S, Sym]{
		states:      make(map[S]struct{}),
		symbols:     make(map[Sym]struct{}),
		accepting:   make(map[S]struct{}),
		transitions: make(map[TransitionKey[S, Sym]]S),
	}
	for _, o := range opts {
		o(&b.options)
	}
	return b
}

// AddState registers a state. If isAccepting is true, it is added to the accepting set.
func (b *Builder[S, Sym]) AddState(state S, isAccepting bool) *Builder[S, Sym] {
	b.states[state] = struct{}{}
	if isAccepting {
		b.accepting[state] = struct{}{}
	}
	return b
}

// SetInitial sets the initial state. The state is implicitly registered.
func (b *Builder[S, Sym]) SetInitial(state S) *Builder[S, Sym] {
	b.initialSet = true
	b.initialState = state
	b.states[state] = struct{}{}
	return b
}

// AddSymbol registers an input symbol.
func (b *Builder[S, Sym]) AddSymbol(sym Sym) *Builder[S, Sym] {
	b.symbols[sym] = struct{}{}
	return b
}

// On adds a transition: from --sym--> to. States and symbol are implicitly registered.
func (b *Builder[S, Sym]) On(from S, sym Sym, to S) *Builder[S, Sym] {
	b.states[from] = struct{}{}
	b.states[to] = struct{}{}
	b.symbols[sym] = struct{}{}
	
	key := TransitionKey[S, Sym]{From: from, Symbol: sym}
	if _, exists := b.transitions[key]; exists && b.options.preventOverwriteTransitions {
		panic(fmt.Sprintf("transition already defined for (%v,%v)", from, sym))
	}
	b.transitions[key] = to
	return b
}

// Optional checks are extracted to helpers to keep Build concise.
func (b *Builder[S, Sym]) checkRequireTotalTransitions(verr *ValidationErrors) {
	if !b.options.requireTotalTransitions {
		return
	}
	for s := range b.states {
		for sym := range b.symbols {
			key := TransitionKey[S, Sym]{From: s, Symbol: sym}
			if _, ok := b.transitions[key]; !ok {
				verr.Append(newBuildError("missing transition from %v on %v", s, sym))
			}
		}
	}
}

func (b *Builder[S, Sym]) checkRequireAtLeastOneAccepting(verr *ValidationErrors) {
	if b.options.requireAtLeastOneAccepting && len(b.accepting) == 0 {
		verr.Append(newBuildError("at least one accepting state required"))
	}
}

func (b *Builder[S, Sym]) checkReachability(verr *ValidationErrors) {
	if !b.initialSet || !(b.options.errorOnUnreachableStates || b.options.errorWhenNoAcceptingReachable) {
		return
	}
	reached := make(map[S]struct{})
	queue := []S{b.initialState}
	reached[b.initialState] = struct{}{}

	for i := 0; i < len(queue); i++ {
		cur := queue[i]
		for key, to := range b.transitions {
			if key.From == cur {
				if _, ok := reached[to]; !ok {
					reached[to] = struct{}{}
					queue = append(queue, to)
				}
			}
		}
	}
	if b.options.errorOnUnreachableStates {
		for s := range b.states {
			if _, ok := reached[s]; !ok {
				verr.Append(newBuildError("unreachable state %v", s))
			}
		}
	}
	if b.options.errorWhenNoAcceptingReachable {
		any := false
		for s := range b.accepting {
			if _, ok := reached[s]; ok {
				any = true
				break
			}
		}
		if !any {
			verr.Append(newBuildError("no accepting state reachable from initial"))
		}
	}
}

// Build validates and returns an immutable Machine.
func (b *Builder[S, Sym]) Build() (*Machine[S, Sym], error) {
	verr := &ValidationErrors{}
	if !b.initialSet {
		verr.Append(newBuildError("initial state must be set"))
	}
	if len(b.states) == 0 {
		verr.Append(newBuildError("at least one state is required"))
	}
	if len(b.symbols) == 0 {
		verr.Append(newBuildError("at least one input symbol is required"))
	}

	// Ensure F ⊆ Q: every accepting state must be a registered state
	for s := range b.accepting {
		if _, ok := b.states[s]; !ok {
			verr.Append(newBuildError("accepting state unknown %v", s))
		}
	}

	// Ensure all transitions reference known states/symbols.
	for key, to := range b.transitions {
		if _, ok := b.states[key.From]; !ok {
			verr.Append(newBuildError("transition from unknown state %v", key.From))
		}
		if _, ok := b.symbols[key.Symbol]; !ok {
			verr.Append(newBuildError("transition uses unknown symbol %v", key.Symbol))
		}
		if _, ok := b.states[to]; !ok {
			verr.Append(newBuildError("transition to unknown state %v", to))
		}
	}

	// Optional checks controlled by flags
	b.checkRequireTotalTransitions(verr)
	b.checkRequireAtLeastOneAccepting(verr)
	b.checkReachability(verr)

	if err := verr.AsError(); err != nil {
		return nil, err
	}

	// Copy into immutable machine.
	acc := make(map[S]struct{}, len(b.accepting))
	for s := range b.accepting {
		acc[s] = struct{}{}
	}
	trans := make(map[TransitionKey[S, Sym]]S, len(b.transitions))
	for key, to := range b.transitions {
		trans[key] = to
	}
	return &Machine[S, Sym]{
		initialState: b.initialState,
		accepting:    acc,
		transitions:  trans,
	}, nil
}


