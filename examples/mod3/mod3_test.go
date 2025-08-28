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


