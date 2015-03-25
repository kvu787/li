package main

import (
	"container/list"
	"regexp"
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
	exprs := pop(stack).(*list.List)
	expr := exprs.Front().Value.(*list.List)
	return expr
}

func push(l *list.List, v interface{}) {
	l.PushFront(v)
}

func pop(l *list.List) interface{} {
	return l.Remove(l.Front())
}
