package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/bohdan-natsevych/fsm-generator/examples/mod3"
)

// CLI to compute mod3 remainder of binary strings using the FSM.
func main() {
	var input string
	flag.StringVar(&input, "in", "", "binary string to evaluate (default: read from stdin)")
	flag.Parse()

	if input == "" {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "read error:", err)
			os.Exit(2)
		}
	}

	rem, err := mod3.ModThree(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(rem)
}


