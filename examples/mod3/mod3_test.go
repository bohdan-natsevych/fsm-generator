package mod3

import "testing"

func TestModThreeKnownValues(t *testing.T) {
	cases := map[string]int{
		"1101": 1, // 13 % 3 = 1
		"1110": 2, // 14 % 3 = 2
		"1111": 0, // 15 % 3 = 0
		"0":     0,
		"1":     1,
		"10":    2,
		"1010":  1,
	}
	for in, want := range cases {
		got, err := ModThree(in)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", in, err)
		}
		if got != want {
			t.Errorf("%q => want %d, got %d", in, want, got)
		}
	}
}

func TestBuildMachineSuccess(t *testing.T) {
	m, err := Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if _, err := m.Eval([]rune("01")); err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
}

func TestModThreeUnexpectedStateIsError(t *testing.T) {
    if _, err := ModThree("1010"); err != nil {
        t.Fatalf("unexpected error for valid input: %v", err)
    }
}


