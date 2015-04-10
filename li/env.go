package li

import (
	"math/rand"
	"time"
)

func createDefaultEnv() map[string]interface{} {
	env := copyEnv(defaultEnv)
	return env
}

func copyEnv(src map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k := range src {
		copy[k] = src[k]
	}
	return copy
}

func createIntBinaryproc(bf func(a, b int) interface{}) proc {
	return proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) interface{} {
			return bf(env["a"].(int), env["b"].(int))
		},
	}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var defaultEnv = map[string]interface{}{
	// return random integer in [0, n)
	"random": proc{
		[]string{"i"},
		func(env map[string]interface{}) interface{} {
			return rng.Intn(env["i"].(int))
		},
	},

	"cons": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) interface{} {
			return [2]interface{}{env["a"], env["b"]}
		},
	},

	"car": proc{
		[]string{"a"},
		func(env map[string]interface{}) interface{} {
			return env["a"].([2]interface{})[0]
		},
	},

	"cdr": proc{
		[]string{"a"},
		func(env map[string]interface{}) interface{} {
			return env["a"].([2]interface{})[1]
		},
	},

	"null?": proc{
		[]string{"a"},
		func(env map[string]interface{}) interface{} {
			a := env["a"].([2]interface{})
			return (a[0] == nil) && (a[1] == nil)
		},
	},

	"list": variadicProc{
		"elements",
		func(env map[string]interface{}) interface{} {
			elements := env["elements"].([]interface{})
			result := [2]interface{}{nil, nil}
			for i := len(elements) - 1; i >= 0; i-- {
				result = [2]interface{}{elements[i], result}
			}
			return result
		},
	},

	"+": variadicProc{
		"nums",
		func(env map[string]interface{}) interface{} {
			result := 0
			nums := env["nums"].([]interface{})
			for _, num := range nums {
				result += num.(int)
			}
			return result
		},
	},

	"*": variadicProc{
		"nums",
		func(env map[string]interface{}) interface{} {
			result := 1
			nums := env["nums"].([]interface{})
			for _, num := range nums {
				result *= num.(int)
			}
			return result
		},
	},

	"not": proc{
		[]string{"a"},
		func(env map[string]interface{}) interface{} {
			a := env["a"].(bool)
			return !a
		},
	},

	"and": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) interface{} {
			a := env["a"].(bool)
			b := env["b"].(bool)
			return a && b
		},
	},

	"or": proc{
		[]string{"a", "b"},
		func(env map[string]interface{}) interface{} {
			a := env["a"].(bool)
			b := env["b"].(bool)
			return a || b
		},
	},

	"-":         createIntBinaryproc(func(a, b int) interface{} { return a - b }),
	"/":         createIntBinaryproc(func(a, b int) interface{} { return a / b }),
	">":         createIntBinaryproc(func(a, b int) interface{} { return a > b }),
	">=":        createIntBinaryproc(func(a, b int) interface{} { return a >= b }),
	"<":         createIntBinaryproc(func(a, b int) interface{} { return a < b }),
	"<=":        createIntBinaryproc(func(a, b int) interface{} { return a <= b }),
	"=":         createIntBinaryproc(func(a, b int) interface{} { return a == b }),
	"remainder": createIntBinaryproc(func(a, b int) interface{} { return a % b }),

	// modifies given env
	"define": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
		name := args[0].(string)
		value := args[1]
		env[name] = Eval(value, env)
		return nil
	}),

	"lambda": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
		if _, ok := args[0].([]interface{}); ok {
			// normal args
			params := args[0].([]interface{})
			stringParams := make([]string, len(params))
			for i := range params {
				stringParams[i] = params[i].(string)
			}
			body := func(env map[string]interface{}) interface{} { return Eval(args[1], env) }
			return proc{
				stringParams,
				body,
			}
		} else {
			// variadic args
			param := args[0].(string)
			body := func(env map[string]interface{}) interface{} { return Eval(args[1], env) }
			return variadicProc{
				param,
				body,
			}
		}
	}),

	"if": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
		condition := args[0]
		conseq := args[1]
		alt := args[2]
		if Eval(condition, env).(bool) {
			return Eval(conseq, env)
		} else {
			return Eval(alt, env)
		}
	}),

	"cond": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
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

	"begin": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
		beginEnv := copyEnv(env)
		var retval interface{} = nil
		for _, arg := range args {
			retval = Eval(arg, beginEnv)
		}
		return retval
	}),

	"let": specialForm(func(args []interface{}, env map[string]interface{}) interface{} {
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
