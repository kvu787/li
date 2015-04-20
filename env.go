package main

import (
	"fmt"
	"math/rand"
	"time"
)

func copyEnv(src map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k := range src {
		copy[k] = src[k]
	}
	return copy
}

func createTypeError(name string, expectedType string, actual interface{}) error {
	return fmt.Errorf(
		"Eval: procedure '%s' expected argument type '%s', but got '%T'",
		name, expectedType, actual)
}

func createArgLenError(name string, expected int, args []interface{}) error {
	return fmt.Errorf(
		"Eval: procedure '%v' expected %d arguments, but got %d arguments",
		name, expected, len(args))
}

func createIntBinaryProc(name string, bf func(a, b int) interface{}) proc {
	return proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) (interface{}, error) {
			a, ok := env["a"].(int)
			if !ok {
				return nil, createTypeError(name, "int", env["a"])
			}
			b, ok := env["b"].(int)
			if !ok {
				return nil, createTypeError(name, "int", env["b"])
			}
			return bf(a, b), nil
		},
	}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var defaultEnv = map[string]interface{}{
	// return random integer in [0, n)
	"random": proc{
		[]string{"n"},
		func(env map[string]interface{}) (interface{}, error) {
			if n, ok := env["n"].(int); ok {
				return rng.Intn(n), nil
			} else {
				return nil, createTypeError("random", "int", env["n"])
			}
		},
	},

	"cons": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) (interface{}, error) {
			return [2]interface{}{env["a"], env["b"]}, nil
		},
	},

	"car": proc{
		[]string{"a"},
		func(env map[string]interface{}) (interface{}, error) {
			if a, ok := env["a"].([2]interface{}); ok {
				return a[0], nil
			} else {
				return nil, createTypeError("car", "[2]interface{}", env["a"])
			}
		},
	},

	"cdr": proc{
		[]string{"a"},
		func(env map[string]interface{}) (interface{}, error) {
			if a, ok := env["a"].([2]interface{}); ok {
				return a[1], nil
			} else {
				return nil, createTypeError("cdr", "[2]interface{}", env["a"])
			}
		},
	},

	"null?": proc{
		[]string{"a"},
		func(env map[string]interface{}) (interface{}, error) {
			if a, ok := env["a"].([2]interface{}); ok {
				return a == [2]interface{}{nil, nil}, nil
			} else {
				return nil, createTypeError("null?", "[2]interface{}", env["a"])
			}
		},
	},

	"list": variadicProc{
		"elements",
		func(env map[string]interface{}) (interface{}, error) {
			elements := env["elements"].([]interface{})
			result := [2]interface{}{nil, nil}
			for i := len(elements) - 1; i >= 0; i-- {
				result = [2]interface{}{elements[i], result}
			}
			return result, nil
		},
	},

	"+": variadicProc{
		"nums",
		func(env map[string]interface{}) (interface{}, error) {
			result := 0
			nums := env["nums"].([]interface{})
			for _, inum := range nums {
				if num, ok := inum.(int); ok {
					result += num
				} else {
					return nil, createTypeError("+", "int", inum)
				}
			}
			return result, nil
		},
	},

	"*": variadicProc{
		"nums",
		func(env map[string]interface{}) (interface{}, error) {
			result := 1
			nums := env["nums"].([]interface{})
			for _, inum := range nums {
				if num, ok := inum.(int); ok {
					result *= num
				} else {
					return nil, createTypeError("*", "int", inum)
				}
			}
			return result, nil
		},
	},

	"not": proc{
		[]string{"a"},
		func(env map[string]interface{}) (interface{}, error) {
			if a, ok := env["a"].(bool); ok {
				return !a, nil
			} else {
				return nil, createTypeError("not", "bool", env["a"])
			}
		},
	},

	"and": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) (interface{}, error) {
			a, ok := env["a"].(bool)
			if !ok {
				return nil, createTypeError("and", "bool", env["a"])
			}
			b, ok := env["b"].(bool)
			if !ok {
				return nil, createTypeError("and", "bool", env["b"])
			}
			return a && b, nil
		},
	},

	"or": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) (interface{}, error) {
			a, ok := env["a"].(bool)
			if !ok {
				return nil, createTypeError("or", "bool", env["a"])
			}
			b, ok := env["b"].(bool)
			if !ok {
				return nil, createTypeError("or", "bool", env["b"])
			}
			return a || b, nil
		},
	},

	"-":         createIntBinaryProc("-", func(a, b int) interface{} { return a - b }),
	"/":         createIntBinaryProc("/", func(a, b int) interface{} { return a / b }),
	">":         createIntBinaryProc(">", func(a, b int) interface{} { return a > b }),
	">=":        createIntBinaryProc(">=", func(a, b int) interface{} { return a >= b }),
	"<":         createIntBinaryProc("<", func(a, b int) interface{} { return a < b }),
	"<=":        createIntBinaryProc("<=", func(a, b int) interface{} { return a <= b }),
	"=":         createIntBinaryProc("=", func(a, b int) interface{} { return a == b }),
	"remainder": createIntBinaryProc("remainder", func(a, b int) interface{} { return a % b }),

	// modifies given env
	"define": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, createArgLenError("define", 2, args)
		}

		name := args[0].(string)
		value := args[1]
		val, err := Eval(value, env)
		if err != nil {
			return nil, err
		}
		env[name] = val
		return nil, nil
	}),

	"lambda": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, createArgLenError("lambda", 2, args)
		}

		if params, ok := args[0].([]interface{}); ok {
			stringParams := make([]string, len(params))
			for i := range params {
				if param, ok := params[i].(string); ok {
					stringParams[i] = param
				} else {
					return nil, fmt.Errorf("Eval: procedure 'lambda' expected 'string' parameter names, got '%T'", params[i])
				}
			}
			return proc{
				stringParams,
				func(env map[string]interface{}) (interface{}, error) {
					return Eval(args[1], env)
				},
			}, nil
		} else if param, ok := args[0].(string); ok { // variadic args
			return variadicProc{
				param,
				func(env map[string]interface{}) (interface{}, error) {
					return Eval(args[1], env)
				},
			}, nil
		} else {
			return nil, fmt.Errorf("Eval: procedure 'lambda' expected 'list' or 'string' type for first argument, got '%T'", args[0])
		}
	}),

	"if": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, createArgLenError("if", 3, args)
		}

		condition := args[0]
		conseq := args[1]
		alt := args[2]
		conditionVal, err := Eval(condition, env)
		if err != nil {
			return nil, err
		}
		conditionBool, ok := conditionVal.(bool)
		if !ok {
			return nil, fmt.Errorf("Eval: procedure 'if' expected 'bool' type for condition, got '%T'", conditionVal)
		}
		if conditionBool {
			return Eval(conseq, env)
		} else {
			return Eval(alt, env)
		}
	}),

	"cond": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("Eval: procedure 'cond' expected at least 1 argument, got 0")
		}

		// check n - 1 branches
		for _, arg := range args[:len(args)-1] {
			branch, ok := arg.([]interface{})
			if !ok {
				return nil, fmt.Errorf("Eval: procedure 'cond' expected 'list' type for argument, got '%T'", arg)
			}

			if len(branch) != 2 {
				return nil, fmt.Errorf("Eval: procedure 'cond' expected 2 items in a branch, got %d", len(branch))
			}
			condition := branch[0]
			body := branch[1]

			conditionVal, err := Eval(condition, env)
			if err != nil {
				return nil, err
			}
			conditionBool, ok := conditionVal.(bool)
			if !ok {
				return nil, fmt.Errorf("Eval: procedure 'cond' expected 'bool' type for condition, got '%T'", conditionVal)
			}
			if conditionBool {
				return Eval(body, env)
			}
		}

		// check last branch for 'else'
		branch, ok := args[len(args)-1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("Eval: procedure 'cond' expected 'list' type for argument, got '%T'", args[len(args)-1])
		}
		if len(branch) != 2 {
			return nil, fmt.Errorf("Eval: procedure 'cond' expected 2 items in a branch, got %d", len(branch))
		}
		condition := branch[0]
		body := branch[1]
		if scond, ok := condition.(string); ok && (scond == "else") {
			return Eval(body, env)
		} else {
			conditionVal, err := Eval(condition, env)
			if err != nil {
				return nil, err
			}
			conditionBool, ok := conditionVal.(bool)
			if !ok {
				return nil, fmt.Errorf("Eval: procedure 'cond' expected 'bool' type for condition, got '%T'", conditionVal)
			}
			if conditionBool {
				return Eval(body, env)
			}
		}
		return nil, fmt.Errorf("Eval: no branch matched in 'cond' procedure")
	}),

	"begin": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		beginEnv := copyEnv(env)
		var retval interface{} = nil
		for _, arg := range args {
			var err error
			retval, err = Eval(arg, beginEnv)
			if err != nil {
				return nil, err
			}
		}
		return retval, nil
	}),

	"let": specialForm(func(args []interface{}, env map[string]interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, createArgLenError("let", 2, args)
		}

		letEnv := copyEnv(env)
		defs, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("Eval: procedure 'let' expected type '[]interface{}' for arg 1, got '%T'", args[0])
		}
		for _, def := range defs {
			pair, ok := def.([]interface{})
			if !ok {
				return nil, fmt.Errorf("Eval: procedure 'let' expected a definition of type 'list', but got '%T'", def)
			}
			if len(pair) != 2 {
				return nil, fmt.Errorf("Eval: procedure 'let' expected a definition list of length 2, but got %d", len(pair))
			}
			name, ok := pair[0].(string)
			if !ok {
				return nil, fmt.Errorf("Eval: procedure 'let' expected a definition name of type 'string', but got '%T'", pair[0])
			}
			value := pair[1]
			v, err := Eval(value, env)
			if err != nil {
				return nil, err
			}
			letEnv[name] = v
		}
		return Eval(args[1], letEnv)
	}),
}
