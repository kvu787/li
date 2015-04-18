package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrUnrecognizedToken      = errors.New("Lex: unrecognized token")
	ErrIncompleteExpression   = errors.New("Parse: incomplete expression")
	ErrOvercompleteExpression = errors.New("Parse: overcomplete expression")
)

func Lex(src string) ([]string, error) {
	// declare regexp strings
	reStrings := []string{
		`(#t)|(#f)`,                  // boolean literals
		`[(]|[)]`,                    // parens
		`[123456789]\d*`,             // integer literals
		`\+|\-|\*|/|<|>|(<=)|(>=)|=`, // operators
		`(\w|\-|\?)+`,                // identifiers
		`;.*`,                        // single-line comments
		`((?s)[[:space:]]+)`,         // whitespace
	}

	// compile strings to regexp objects
	regexes := make([]*regexp.Regexp, len(reStrings))
	for i, v := range reStrings {
		regexes[i] = regexp.MustCompile(v)
	}

	// let i be the current index of input
	// advance and tokenize input until end of input or error
	tokens := []string{}
	for i := 0; i < len(src); {

		// check if any regex matches input
		reMatched := false
		for _, re := range regexes {
			loc := re.FindStringIndex(src[i:])
			if loc != nil && loc[0] == 0 {
				tokens = append(tokens, src[i:][loc[0]:loc[1]])
				reMatched = true
				i += (loc[1] - loc[0])
				break
			}
		}

		// error if no regex can match current input
		if !reMatched {
			return nil, ErrUnrecognizedToken
		}
	}

	return tokens, nil
}

// Remove whitespace and comment tokens
func Preprocess(tokens []string) []string {
	preprocessedTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if strings.ContainsRune("\t\n\v\f\r ", rune(token[0])) {
			continue
		}
		if token[0] == ';' {
			continue
		}
		preprocessedTokens = append(preprocessedTokens, token)
	}
	return preprocessedTokens
}

func Parse(tokens []string) ([]interface{}, error) {
	stk := NewStack()
	stk.Push(NewStack())
	for _, token := range tokens {
		if token == "(" {
			stk.Push(NewStack())
		} else if token == ")" {
			childExpr := stk.Pop().(Stack)
			if stk.Len() == 0 {
				return nil, ErrOvercompleteExpression
			}
			parentExpr := stk.Pop().(Stack)
			parentExpr.Push(childExpr)
			stk.Push(parentExpr)
		} else {
			expr := stk.Pop().(Stack)
			expr.Push(token)
			stk.Push(expr)
		}
	}
	if stk.Len() > 1 {
		return nil, ErrIncompleteExpression
	}
	return stk.Pop().(Stack).ToSlice(), nil
}

func Eval(expr interface{}, env map[string]interface{}) (interface{}, error) {
	switch expr.(type) {
	case []interface{}:
		// must be a function application
		lst := expr.([]interface{})
		if len(lst) == 0 {
			return nil, fmt.Errorf("Eval: cannot evaluate empty expression ()")
		}
		function, err := Eval(lst[0], env)
		if err != nil {
			return nil, err
		}
		switch function.(type) {
		case specialForm:
			return function.(specialForm)(lst[1:], env)
		case proc:
			proc := function.(proc)
			args := lst[1:]
			if len(args) != len(proc.params) {
				return nil, fmt.Errorf("Eval: wrong number of params")
			}
			procEnv := copyEnv(env)
			evaluatedArgs := make([]interface{}, len(args))
			for i := range args {
				evaluatedArg, err := Eval(args[i], env)
				evaluatedArgs[i] = evaluatedArg
				if err != nil {
					return nil, err
				}
			}
			for i := range evaluatedArgs {
				procEnv[proc.params[i]] = evaluatedArgs[i]
			}
			return proc.body(procEnv)
		case variadicProc:
			vproc := function.(variadicProc)
			args := lst[1:]
			procEnv := copyEnv(env)
			evaluatedArgs := make([]interface{}, len(args))
			for i := range args {
				evaluatedArg, err := Eval(args[i], env)
				evaluatedArgs[i] = evaluatedArg
				if err != nil {
					return nil, err
				}
			}
			procEnv[vproc.param] = evaluatedArgs
			return vproc.body(procEnv)
		default:
			return nil, fmt.Errorf(
				"Eval: expected special form or procedure but received type '%T'",
				function)
		}
	case string:
		// must be either literal or a binding
		s := expr.(string)
		if s == "#t" {
			// true literal
			return true, nil
		} else if s == "#f" {
			// false literal
			return false, nil
		} else if i, err := strconv.Atoi(s); err == nil {
			// integer literal
			return i, nil
		} else {
			// identifier
			val, ok := env[s]
			if !ok {
				return nil, fmt.Errorf("Eval: identifier not found: '%s'", s)
			} else {
				return val, nil
			}
		}
	default:
		return nil, fmt.Errorf(`Eval: received invalid expression
	type: '%T'
	value: '%v'`,
			expr, expr)
	}
}

func Exec(src string) (interface{}, error) {
	// initialize execution environment
	env := copyEnv(defaultEnv)

	var err error

	// lex source into tokens
	tokens, err := Lex(src)
	if err != nil {
		return nil, err
	}

	// remove comments and whitespace
	preprocessedTokens := Preprocess(tokens)

	// parse into AST
	exprs, err := Parse(preprocessedTokens)
	if err != nil {
		return nil, err
	}

	// evaluate expressions
	var retval interface{}
	for _, expr := range exprs {
		retval, err = Eval(expr, env)
		if err != nil {
			return nil, err
		}
	}

	return retval, nil
}
