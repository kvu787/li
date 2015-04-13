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
interpreter. Expressions are read from standard input. The result of each
expression is printed to standard output until EOF is encountered or an
error occurs.`

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "-i" {
		repl()
	} else {
		readStdin()
	}

}

func readStdin() {
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
					fmt.Fprintf(os.Stderr, "Error when reading from stdin: %v\n", err)
					os.Exit(1)
				}
			}
			linesc <- s
		}
	}()

	// lex
	go func() {
		for line := range linesc {
			tokens, err := Lex(line)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
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
				if err == ErrIncompleteExpression {
					continue
				} else {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				exprsc <- exprs[0]
				currentTokens = []string{}
			}
		}
		close(exprsc)
	}()

	// eval
	for expr := range exprsc {
		v, err := Eval(expr, env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(">>>", v)
	}

	os.Exit(0)
}
