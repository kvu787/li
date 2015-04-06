package main

import (
	"container/list"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

var specialForms map[string]Proc

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

func Parse(tokens []string) *list.List {
	stack := list.New()
	push(stack, list.New())
	for _, token := range tokens {
		if token == "(" {
			push(stack, list.New())
		} else if token == ")" {
			childExpr := pop(stack).(*list.List)
			parentExpr := pop(stack).(*list.List)
			parentExpr.PushBack(childExpr)
			push(stack, parentExpr)
		} else if token[0] == ';' { // ignore comments
			continue
		} else {
			expr := pop(stack).(*list.List)
			expr.PushBack(token)
			push(stack, expr)
		}
	}
	return pop(stack).(*list.List)
}

func init() {
	specialForms = map[string]Proc{
		"define": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			name := expr.Front().Next().Value.(string)
			value := expr.Front().Next().Next().Value
			env[name] = Eval(value, env)
			return nil
		}),
		"lambda": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			params := expr.Front().Next().Value.(*list.List)
			body := expr.Front().Next().Next().Value
			return Proc(func(args *list.List, env map[string]interface{}) interface{} {
				// set arguments and evaluate
				procEnv := copyEnv(env)
				e1 := params.Front()
				e2 := args.Front()
				for e1 != nil {
					procEnv[e1.Value.(string)] = e2.Value
					e1 = e1.Next()
					e2 = e2.Next()
				}
				return Eval(body, procEnv)
			})
		}),
		"if": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			condition := expr.Front().Next().Value
			conseq := expr.Front().Next().Next().Value
			alt := expr.Front().Next().Next().Next().Value
			if Eval(condition, env).(bool) {
				return Eval(conseq, env)
			} else {
				return Eval(alt, env)
			}
		}),
		"cond": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			var e *list.Element

			// check n - 1 branches
			for e = expr.Front().Next(); e.Next() != nil; e = e.Next() {
				branch := e.Value.(*list.List)
				conditionExpr := branch.Front().Value
				if Eval(conditionExpr, env).(bool) {
					return Eval(branch.Front().Next().Value, env)
				}
			}

			// check last branch for 'else'
			branch := e.Value.(*list.List)
			if scond, ok := branch.Front().Value.(string); ok && (scond == "else") {
				return Eval(branch.Front().Next().Value, env)
			} else {
				conditionExpr := branch.Front().Value
				if Eval(conditionExpr, env).(bool) {
					return Eval(branch.Front().Next().Value, env)
				}
			}
			panic("Eval: no branch matched in 'cond' expression")
		}),
		"begin": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			beginEnv := copyEnv(env)
			var retval interface{} = nil
			for e := expr.Front().Next(); e != nil; e = e.Next() {
				retval = Eval(e.Value, beginEnv)
			}
			return retval
		}),
		"let": Proc(func(expr *list.List, env map[string]interface{}) interface{} {
			letEnv := copyEnv(env)
			defList := expr.Front().Next().Value.(*list.List)
			for e := defList.Front(); e != nil; e = e.Next() {
				pair := e.Value.(*list.List)
				name := pair.Front().Value.(string)
				value := pair.Front().Next().Value
				letEnv[name] = Eval(value, env)
			}
			return Eval(expr.Front().Next().Next().Value, letEnv)
		}),
	}
}

func Eval(expr interface{}, env map[string]interface{}) interface{} {
	switch expr.(type) {
	case *list.List:
		l := expr.(*list.List)
		if l.Len() > 1 {
			// must be a function application
			function := l.Front().Value.(string)
			if sf, ok := specialForms[function]; ok {
				// check if special form
				return sf(l, env)
			} else {
				// eval as regular procedure
				proc, ok := Eval(function, env).(Proc)
				if !ok {
					panic(fmt.Sprintf(`Eval: expected procedure, received token %s`, function))
				}
				args := list.New()
				for e := l.Front().Next(); e != nil; e = e.Next() {
					args.PushBack(Eval(e.Value, env))
				}
				return proc(args, env)
			}
		} else {
			// might be a function application
			name := l.Front().Value.(string)
			if val, ok := Eval(name, env).(Proc); ok {
				// apply if proc
				return val(list.New(), env)
			} else {
				// return val if not proc
				return val
			}
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
	tokens, err := Lex(src)
	if err != nil {
		return nil, err
	}
	exprs := Parse(tokens)
	var retval interface{}
	for e := exprs.Front(); e != nil; e = e.Next() {
		retval = Eval(e.Value, env)
	}
	return retval, nil
}

type Proc func(args *list.List, env map[string]interface{}) interface{}

func push(l *list.List, v interface{}) {
	l.PushFront(v)
}

func pop(l *list.List) interface{} {
	return l.Remove(l.Front())
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
	createIntBinaryProc := func(bf func(a, b int) interface{}) Proc {
		return Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.(int)
			b := args.Front().Next().Value.(int)
			return bf(a, b)
		})
	}

	return map[string]interface{}{
		// return random integer in [0, n)
		"random": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value
			return rng.Intn(a.(int))
		}),
		"cons": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value
			b := args.Front().Next().Value
			return [2]interface{}{a, b}
		}),
		"car": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.([2]interface{})
			return a[0]
		}),
		"cdr": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.([2]interface{})
			return a[1]
		}),
		"null?": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.([2]interface{})
			return (a[0] == nil) && (a[1] == nil)
		}),
		"list": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			result := [2]interface{}{nil, nil}
			for e := args.Back(); e != nil; e = e.Prev() {
				result = [2]interface{}{e.Value, result}
			}
			return result
		}),
		"+": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			res := 0
			for e := args.Front(); e != nil; e = e.Next() {
				res += e.Value.(int)
			}
			return res
		}),
		"*": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			res := 1
			for e := args.Front(); e != nil; e = e.Next() {
				res *= e.Value.(int)
			}
			return res
		}),
		"-":         createIntBinaryProc(func(a, b int) interface{} { return a - b }),
		"/":         createIntBinaryProc(func(a, b int) interface{} { return a / b }),
		">":         createIntBinaryProc(func(a, b int) interface{} { return a > b }),
		">=":        createIntBinaryProc(func(a, b int) interface{} { return a >= b }),
		"<":         createIntBinaryProc(func(a, b int) interface{} { return a < b }),
		"<=":        createIntBinaryProc(func(a, b int) interface{} { return a <= b }),
		"=":         createIntBinaryProc(func(a, b int) interface{} { return a == b }),
		"remainder": createIntBinaryProc(func(a, b int) interface{} { return a % b }),
		"not": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			return !(args.Front().Value.(bool))
		}),
		"and": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.(bool)
			b := args.Front().Next().Value.(bool)
			return a && b
		}),
		"or": Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.(bool)
			b := args.Front().Next().Value.(bool)
			return a || b
		}),
	}
}
