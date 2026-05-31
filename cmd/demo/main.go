package main

import (
	"fmt"
	"os"

	termlatex "github.com/floatpane/go-term-latex"
)

func main() {
	eq := `\int_0^\infty e^{-x^2}\,dx = \frac{\sqrt{\pi}}{2}`
	if len(os.Args) > 1 {
		eq = os.Args[1]
	}

	if err := termlatex.Display(os.Stdout, eq, termlatex.Options{}); err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
}
