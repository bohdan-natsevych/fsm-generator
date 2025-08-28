package mod3

import (
	"fmt"

	"github.com/bohdan-natsevych/fsm-generator/pkg/fsm"
)

// Build constructs a modulo-3 FSM for binary input symbols '0' and '1'.
// States represent the current remainder: S0=0, S1=1, S2=2.
func Build() (*fsm.Machine[string, byte], error) {
	b := fsm.NewBuilder[string, byte](
		fsm.WithPreventOverwriteTransitions(),
		fsm.WithErrorOnUnreachableStates(),
		fsm.WithErrorWhenNoAcceptingReachable(),
	)

	// States and accepting set (all states are accepting for modulo remainder output)
	b.AddState("S0", true)
	b.AddState("S1", true)
	b.AddState("S2", true)
	b.SetInitial("S0")

	// Symbols
	b.AddSymbol('0')
	b.AddSymbol('1')

	// Transitions per provided diagram/definition
	// δ(S0,0) = S0; δ(S0,1) = S1
	b.On("S0", '0', "S0").On("S0", '1', "S1")
	// δ(S1,0) = S2; δ(S1,1) = S0
	b.On("S1", '0', "S2").On("S1", '1', "S0")
	// δ(S2,0) = S1; δ(S2,1) = S2
	b.On("S2", '0', "S1").On("S2", '1', "S2")

	return b.Build()
}

// ModThree returns the remainder in {0,1,2} for a binary string input.
func ModThree(binary string) (int, error) {
	m, err := Build()
	if err != nil {
		return 0, err
	}
	// Evaluate
	bs := []byte(binary)
	state, err := m.Eval(bs)
	if err != nil {
		return 0, err
	}
	switch state {
	case "S0":
		return 0, nil
	case "S1":
		return 1, nil
	case "S2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unexpected final state %q", state)
	}
}


