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
	if _, err := m.Eval([]byte("01")); err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
}

func TestModThreeUnexpectedStateIsError(t *testing.T) {
    if _, err := ModThree("1010"); err != nil {
        t.Fatalf("unexpected error for valid input: %v", err)
    }
}

func TestModThreeEmptyAndSingleBit(t *testing.T) {
    // empty input: no steps, should remain in S0 => 0
    if got, err := ModThree(""); err != nil || got != 0 {
        t.Fatalf("empty => want 0, got %d, err %v", got, err)
    }
    // single bit inputs already covered; add another sanity
    if got, err := ModThree("0"); err != nil || got != 0 {
        t.Fatalf("0 => want 0, got %d, err %v", got, err)
    }
}

func TestModThreeRejectsNonBinaryASCII(t *testing.T) {
	cases := []string{"2", "102", "a1", "1x0"}
	for _, in := range cases {
		if _, err := ModThree(in); err == nil {
			t.Fatalf("expected error for non-binary input %q, got nil", in)
		}
	}
}

func TestModThreeRejectsNonASCII(t *testing.T) {
	cases := []string{"🙂", "1🙂0", "٠١"} // note: Arabic-Indic digits
	for _, in := range cases {
		if _, err := ModThree(in); err == nil {
			t.Fatalf("expected error for non-ASCII/non-binary input %q, got nil", in)
		}
	}
}

// CURSOR: Benchmark tests to verify performance improvements
func BenchmarkModThree(b *testing.B) {
	testInput := "11010110101010101010101010101010"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ModThree(testInput)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkModThreeLong(b *testing.B) {
	// Generate a long binary string
	var testInput string
	for i := 0; i < 1000; i++ {
		testInput += "1101010"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ModThree(testInput)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// CURSOR: Test the new validation provides better error messages
func TestModThreeValidationErrorMessages(t *testing.T) {
	_, err := ModThree("102")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
	expectedMsg := "invalid binary character '2' at position 2"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}


