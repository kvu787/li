package main

import (
	"container/list"
	"regexp"
	"strconv"
)

func main() {
}

func Lex(src string) []string {
	bools := `(#t)|(#f)`
	parens := `[(]|[)]`
	numbers := `\d+`
	operators := `\+|\-|\*|/|<|>|(<=)|(>=)`
	identifiers := `(\w|\-)+`
	re := regexp.MustCompile(
		bools +
			"|" + parens +
			"|" + numbers +
			"|" + operators +
			"|" + identifiers)
	matches := re.FindAllString(src, -1)
	return matches
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
		} else {
			expr := pop(stack).(*list.List)
			expr.PushBack(token)
			push(stack, expr)
		}
	}
	return pop(stack).(*list.List)
}

func Eval(expr interface{}, env map[string]interface{}) interface{} {
	switch expr.(type) {
	case *list.List:
		// must be a function application
		l := expr.(*list.List)
		function := l.Front().Value.(string)
		switch function {
		case "define":
			// modify the current environment
			name := l.Front().Next().Value.(string)
			value := l.Front().Next().Next().Value
			env[name] = Eval(value, env)
			return nil
		case "lambda":
			params := l.Front().Next().Value.(*list.List)
			body := l.Front().Next().Next().Value.(*list.List)
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
		case "if":
			conditionExpr := l.Front().Next().Value
			conseqExpr := l.Front().Next().Next().Value
			altExpr := l.Front().Next().Next().Next().Value
			if Eval(conditionExpr, env).(bool) {
				return Eval(conseqExpr, env)
			} else {
				return Eval(altExpr, env)
			}
		default:
			proc := Eval(function, env).(Proc)
			args := list.New()
			for e := l.Front().Next(); e != nil; e = e.Next() {
				args.PushBack(Eval(e.Value, env))
			}
			return proc(args, env)
		}
	case string:
		// must be either literal or a binding
		s := expr.(string)
		if s == "#t" {
			return true
		} else if s == "#f" {
			return false
		} else if i, err := strconv.Atoi(s); err == nil {
			return i
		} else {
			return env[s]
		}
	}
	panic("bad")
}

func Exec(src string) interface{} {
	env := CreateDefaultEnv()
	tokens := Lex(src)
	exprs := Parse(tokens)
	var retval interface{}
	for e := exprs.Front(); e != nil; e = e.Next() {
		retval = Eval(e.Value, env)
	}
	return retval
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

func CreateDefaultEnv() map[string]interface{} {
	createIntBinaryProc := func(bf func(a, b int) interface{}) Proc {
		return Proc(func(args *list.List, _ map[string]interface{}) interface{} {
			a := args.Front().Value.(int)
			b := args.Front().Next().Value.(int)
			return bf(a, b)
		})
	}

	return map[string]interface{}{
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
		"-":  createIntBinaryProc(func(a, b int) interface{} { return a - b }),
		"/":  createIntBinaryProc(func(a, b int) interface{} { return a / b }),
		">":  createIntBinaryProc(func(a, b int) interface{} { return a > b }),
		">=": createIntBinaryProc(func(a, b int) interface{} { return a >= b }),
		"<":  createIntBinaryProc(func(a, b int) interface{} { return a < b }),
		"<=": createIntBinaryProc(func(a, b int) interface{} { return a <= b }),
	}
}
