package li

import (
	"fmt"
	"github.com/kvu787/li/stack"
	"regexp"
	"strconv"
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
		for j, re := range regexes {
			loc := re.FindStringIndex(src[i:])
			if loc != nil && loc[0] == 0 {

				// skip whitespace regex
				if j != len(reStrings)-1 {
					tokens = append(tokens, src[i:][loc[0]:loc[1]])
				}

				reMatched = true
				i += (loc[1] - loc[0])
				break
			}
		}

		// error if no regex can match current input
		if !reMatched {
			return nil, fmt.Errorf("Lex: unrecognized token: %q", src[i:])
		}
	}

	return tokens, nil
}

func Parse(tokens []string) ([]interface{}, error) {
	stk := stack.New()
	stk.Push(stack.New())
	for _, token := range tokens {
		if token == "(" {
			stk.Push(stack.New())
		} else if token == ")" {
			childExpr := stk.Pop().(stack.Stack)
			if stk.Len() == 0 {
				return nil, fmt.Errorf("Parse: overcomplete expression")
			}
			parentExpr := stk.Pop().(stack.Stack)
			parentExpr.Push(childExpr)
			stk.Push(parentExpr)
		} else if token[0] == ';' { // ignore comments
			continue
		} else {
			expr := stk.Pop().(stack.Stack)
			expr.Push(token)
			stk.Push(expr)
		}
	}
	if stk.Len() > 1 {
		return nil, fmt.Errorf("Parse: incomplete expression")
	}
	return stk.Pop().(stack.Stack).ToSlice(), nil
}

func Eval(expr interface{}, env map[string]interface{}) interface{} {
	switch expr.(type) {
	case []interface{}:
		// must be a function application
		lst := expr.([]interface{})
		function := Eval(lst[0], env)
		switch function.(type) {
		case specialForm:
			return function.(specialForm)(lst[1:], env)
		case proc:
			proc := function.(proc)
			args := lst[1:]
			if len(args) != len(proc.params) {
				panic("Eval: wrong number of params")
			}
			procEnv := copyEnv(env)
			evaluatedArgs := make([]interface{}, len(args))
			for i := range args {
				evaluatedArgs[i] = Eval(args[i], env)
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
				evaluatedArgs[i] = Eval(args[i], env)
			}
			procEnv[vproc.param] = evaluatedArgs
			return vproc.body(procEnv)
		default:
			panic("Eval: expected special form or procedure")
		}
	case string:
		// must be either literal or a binding
		s := expr.(string)
		if s == "#t" {
			// true literal
			return true
		} else if s == "#f" {
			// false literal
			return false
		} else if i, err := strconv.Atoi(s); err == nil {
			// integer literal
			return i
		} else {
			// identifier
			val, ok := env[s]
			if !ok {
				panic(fmt.Sprintf("Eval: identifier not found: %s", s))
			} else {
				return val
			}
		}
	default:
		panic(fmt.Sprintf(`Eval: received invalid expression
	type: %T
	value: %v`,
			expr, expr))
	}
}

func Exec(src string) (interface{}, error) {
	env := createDefaultEnv()

	var err error
	tokens, err := Lex(src)
	if err != nil {
		return nil, err
	}

	exprs, err := Parse(tokens)
	if err != nil {
		return nil, err
	}

	var retval interface{}
	for _, expr := range exprs {
		retval = Eval(expr, env)
	}

	return retval, nil
}
