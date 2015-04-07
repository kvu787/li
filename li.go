package main

import (
	"fmt"
	"li/stack"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

func main() {

}

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
		case SpecialForm:
			return function.(SpecialForm)(lst[1:], env)
		case Proc:
			proc := function.(Proc)
			args := lst[1:]
			procEnv := copyEnv(env)
			evaluatedArgs := make([]interface{}, len(args))
			for i := range args {
				evaluatedArgs[i] = Eval(args[i], env)
			}
			if proc.isVariadic {
				procEnv[proc.params[0]] = evaluatedArgs
			} else {
				for i := range evaluatedArgs {
					procEnv[proc.params[i]] = evaluatedArgs[i]
				}
			}
			return proc.body(procEnv)
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
	env := CreateDefaultEnv()

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

type SpecialForm func(args []interface{}, env map[string]interface{}) interface{}
type Proc struct {
	params     []string
	isVariadic bool
	body       func(env map[string]interface{}) interface{}
}

func copyEnv(src map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k := range src {
		copy[k] = src[k]
	}
	return copy
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func CreateDefaultEnv() map[string]interface{} {
	// return nil
	createIntBinaryProc := func(bf func(a, b int) interface{}) Proc {
		return Proc{
			[]string{"a", "b"},
			false,
			func(env map[string]interface{}) interface{} {
				return bf(env["a"].(int), env["b"].(int))
			},
		}
	}

	return map[string]interface{}{
		// return random integer in [0, n)
		"random": Proc{
			[]string{"i"},
			false,
			func(env map[string]interface{}) interface{} {
				return rng.Intn(env["i"].(int))
			},
		},
		"cons": Proc{
			[]string{"a", "b"},
			false,
			func(env map[string]interface{}) interface{} {
				return [2]interface{}{env["a"], env["b"]}
			},
		},

		"car": Proc{
			[]string{"a"},
			false,
			func(env map[string]interface{}) interface{} {
				return env["a"].([2]interface{})[0]
			},
		},

		"cdr": Proc{
			[]string{"a"},
			false,
			func(env map[string]interface{}) interface{} {
				return env["a"].([2]interface{})[1]
			},
		},

		"null?": Proc{
			[]string{"a"},
			false,
			func(env map[string]interface{}) interface{} {
				a := env["a"].([2]interface{})
				return (a[0] == nil) && (a[1] == nil)
			},
		},

		"list": Proc{
			[]string{"elements"},
			true,
			func(env map[string]interface{}) interface{} {
				elements := env["elements"].([]interface{})
				result := [2]interface{}{nil, nil}
				for i := len(elements) - 1; i >= 0; i-- {
					result = [2]interface{}{elements[i], result}
				}
				return result
			},
		},
		"+": Proc{
			[]string{"nums"},
			true,
			func(env map[string]interface{}) interface{} {
				result := 0
				nums := env["nums"].([]interface{})
				for _, num := range nums {
					result += num.(int)
				}
				return result
			},
		},
		"*": Proc{
			[]string{"nums"},
			true,
			func(env map[string]interface{}) interface{} {
				result := 1
				nums := env["nums"].([]interface{})
				for _, num := range nums {
					result *= num.(int)
				}
				return result
			},
		},
		"not": Proc{
			[]string{"a"},
			false,
			func(env map[string]interface{}) interface{} {
				a := env["a"].(bool)
				return !a
			},
		},
		"and": Proc{
			[]string{"a", "b"},
			false,
			func(env map[string]interface{}) interface{} {
				a := env["a"].(bool)
				b := env["b"].(bool)
				return a && b
			},
		},
		"or": Proc{
			[]string{"a", "b"},
			false,
			func(env map[string]interface{}) interface{} {
				a := env["a"].(bool)
				b := env["b"].(bool)
				return a || b
			},
		},
		"-":         createIntBinaryProc(func(a, b int) interface{} { return a - b }),
		"/":         createIntBinaryProc(func(a, b int) interface{} { return a / b }),
		">":         createIntBinaryProc(func(a, b int) interface{} { return a > b }),
		">=":        createIntBinaryProc(func(a, b int) interface{} { return a >= b }),
		"<":         createIntBinaryProc(func(a, b int) interface{} { return a < b }),
		"<=":        createIntBinaryProc(func(a, b int) interface{} { return a <= b }),
		"=":         createIntBinaryProc(func(a, b int) interface{} { return a == b }),
		"remainder": createIntBinaryProc(func(a, b int) interface{} { return a % b }),

		// modifies given env
		"define": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			name := args[0].(string)
			value := args[1]
			env[name] = Eval(value, env)
			return nil
		}),
		"lambda": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			if _, ok := args[0].([]interface{}); ok {
				// normal args
				params := args[0].([]interface{})
				stringParams := make([]string, len(params))
				for i := range params {
					stringParams[i] = params[i].(string)
				}
				body := func(env map[string]interface{}) interface{} { return Eval(args[1], env) }
				return Proc{
					stringParams,
					false,
					body,
				}
			} else {
				// variadic args
				param := args[0].(string)
				body := func(env map[string]interface{}) interface{} { return Eval(args[1], env) }
				return Proc{
					[]string{param},
					true,
					body,
				}
			}
		}),
		"if": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			condition := args[0]
			conseq := args[1]
			alt := args[2]
			if Eval(condition, env).(bool) {
				return Eval(conseq, env)
			} else {
				return Eval(alt, env)
			}
		}),
		"cond": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			// check n - 1 branches
			for _, arg := range args[:len(args)-1] {
				branch := arg.([]interface{})
				condition := branch[0]
				body := branch[1]
				if Eval(condition, env).(bool) {
					return Eval(body, env)
				}
			}

			// check last branch for 'else'
			branch := args[len(args)-1].([]interface{})
			condition := branch[0]
			body := branch[1]
			if scond, ok := condition.(string); ok && (scond == "else") {
				return Eval(body, env)
			} else {
				if Eval(condition, env).(bool) {
					return Eval(body, env)
				}
			}
			panic("Eval: no branch matched in 'cond' expression")
		}),
		"begin": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			beginEnv := copyEnv(env)
			var retval interface{} = nil
			for _, arg := range args {
				retval = Eval(arg, beginEnv)
			}
			return retval
		}),
		"let": SpecialForm(func(args []interface{}, env map[string]interface{}) interface{} {
			letEnv := copyEnv(env)
			defs := args[0].([]interface{})
			for _, def := range defs {
				pair := def.([]interface{})
				name := pair[0].(string)
				value := pair[1]
				letEnv[name] = Eval(value, env)
			}
			return Eval(args[1], letEnv)
		}),
	}
}
