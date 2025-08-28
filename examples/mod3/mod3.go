package mod3

import (
	"fmt"
	"sync"

	"github.com/bohdan-natsevych/fsm-generator/pkg/fsm"
)

var (
	// Singleton pattern for better performance - avoid rebuilding FSM on each call
	machine     *fsm.Machine[string, byte]
	machineOnce sync.Once
	machineErr  error
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

// getMachine returns the singleton modulo-3 FSM instance, building it once.
func getMachine() (*fsm.Machine[string, byte], error) {
	machineOnce.Do(func() {
		machine, machineErr = Build()
	})
	return machine, machineErr
}

// ModThree returns the remainder in {0,1,2} for a binary string input.
// The function validates that input contains only binary digits.
func ModThree(binary string) (int, error) {
	// Input validation for better error messages
	if binary == "" {
		return 0, nil // Empty string represents 0, so remainder is 0
	}
	
	// Validate binary input
	for i, char := range binary {
		if char != '0' && char != '1' {
			return 0, fmt.Errorf("invalid binary character '%c' at position %d", char, i)
		}
	}
	
	m, err := getMachine()
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


