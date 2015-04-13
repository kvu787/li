package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var help string = `usage: li [-h | -i]

Li evaulates Scheme (Lisp) expressions.

If no flags are specified, expressions are read from standard input and
evaluated.
The value of the final expression is printed to standard output.

The -i flag launches an interactive read-evaluate-print-loop (REPL)
interpreter.`

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "-i" {
		repl()
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

func repl() {
	linesc := make(chan string)
	tokensc := make(chan string)
	exprsc := make(chan interface{})
	rd := bufio.NewReader(os.Stdin)
	env := copyEnv(defaultEnv)

	// read stdin
	go func() {
		for {
			s, err := rd.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					close(linesc)
					return
				} else {
					fmt.Fprintf(os.Stderr, "Unexpected read err encountered: %v", err)
					os.Exit(1)
				}
			}
			linesc <- s
		}
	}()

	// lex
	go func() {
		for line := range linesc {
			tokens, _ := Lex(line)
			tokens = Preprocess(tokens)
			for _, token := range tokens {
				tokensc <- token
			}
		}
		close(tokensc)
	}()

	// parse
	go func() {
		currentTokens := []string{}
		for token := range tokensc {
			currentTokens = append(currentTokens, token)
			exprs, err := Parse(currentTokens)
			if err != nil {
				continue
			} else {
				exprsc <- exprs[0]
				currentTokens = []string{}
			}
		}
		close(exprsc)
	}()

	// eval
	for expr := range exprsc {
		v, _ := Eval(expr, env)
		fmt.Println(">>>", v)
	}
}
