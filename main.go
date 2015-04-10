package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var help string = `usage: li [-h]

Li evaulates Scheme (Lisp) expressions.

Scheme expressions are read from standard input and evaluated.
The value of the final expression is printed to standard output.`

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println(help)
		os.Exit(0)
	}

	// read from stdin
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// execute src
	expr, err := Exec(string(src))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// print evaluated src
	fmt.Println(expr)
	os.Exit(0)
}
