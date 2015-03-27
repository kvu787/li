package main

import (
	"container/list"
	"regexp"
	"strconv"
)

func main() {
}

func Lex(src string) []string {
	parens := `[(]|[)]`
	numbers := `\d+`
	operators := `\+|\-|\*|/`
	identifiers := `(\w|\-)+`
	re := regexp.MustCompile(
		parens +
			"|" + numbers +
			"|" + operators +
			"|" + identifiers)
	matches := re.FindAllString(src, -1)
	return matches
}

func Parse(tokens []string) interface{} {
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
	exprs := pop(stack).(*list.List)
	expr := exprs.Front().Value
	return expr
}

func Eval(expr interface{}) interface{} {
	switch expr.(type) {
	case *list.List:
		l := expr.(*list.List)
		switch l.Len() {
		case 0:
			// empty expression
			panic("empty expression")
		case 1:
			// cannot be a function, so eval only element
			panic("function application with no arguments")
		default:
			// must be a function
			operator := l.Front().Value.(string)
			a := Eval(l.Front().Next().Value).(int)
			b := Eval(l.Front().Next().Next().Value).(int)
			switch operator {
			case "+":
				return a + b
			case "-":
				return a - b
			case "*":
				return a * b
			case "/":
				return a / b
			default:
				panic("unrecognized operator")
			}
		}
	case string:
		s := expr.(string)
		i, _ := strconv.Atoi(s)
		return i
	default:
		panic("unrecognized expression")
	}
}

func push(l *list.List, v interface{}) {
	l.PushFront(v)
}

func pop(l *list.List) interface{} {
	return l.Remove(l.Front())
}
