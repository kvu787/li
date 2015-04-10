package li

import (
	"container/list"
	"fmt"
	"testing"
)

func TestLex(t *testing.T) {
	{ // test valid source
		tokens := []string{
			"; ignore this comment", "(", "+", "(", "*", "1", "(", "/", "1",
			"zero", ")", ")", "(", "/", "2", "PI", ")",
			"; and this comment",
			"(", "-", "thirty-three", "EULERS_NUMBER", ")", ")",
		}
		src := `; ignore this comment
(+ 
  (* 
  	1 
  	(/ 1 zero)) 
  (/ 2 PI)
  ; and this comment
  (- thirty-three EULERS_NUMBER))`

		actual, err := Lex(src)
		if err != nil {
			t.Fatal(err)
		}
		if !stringSliceEquals(actual, tokens) {
			t.Fatal("Lex failed: expected != actual")
		}
	}

	{ // test invalid source
		src := `; ignore this comment
(~~~ + (* 1 (/ 1 zero)) 
  (/ 2 PI)
  ; and this comment
  (- thirty-three EULERS_NUMBER))`
		_, err := Lex(src)
		if err == nil {
			t.Fatal("Lex failed: did not receive expected error")
		}
	}
}

func TestParse(t *testing.T) {
	{ // test valid tokens
		tokens := []string{
			"; ignore this comment",
			"(", "+", "(", "*", "1", "(", "/", "1",
			"zero", ")", ")", "(", "/", "2", "PI", ")",
			"; and this comment",
			"(", "-", "thirty-three", "EULERS_NUMBER", ")", ")",
			"(", "=", "1", "2", ")",
		}
		expected := []interface{}{
			[]interface{}{"+",
				[]interface{}{"*",
					"1",
					[]interface{}{"/", "1", "zero"}},
				[]interface{}{"/", "2", "PI"},
				[]interface{}{"-", "thirty-three", "EULERS_NUMBER"}},
			[]interface{}{"=", "1", "2"},
		}
		actual, err := Parse(tokens)
		if err != nil {
			t.Fatalf("Parse failed: received unexpected error")
		}
		if !sliceEquals(actual, expected) {
			t.Fatalf("Parse failed: expected != actual")
		}
	}

	{ // test incomplete tokens
		tokens := []string{
			"(", "+", "(", "*", "1", "(", "/", "1",
			"zero", ")", ")", "(", "/", "2", "PI", ")",
			"(", "-", "thirty-three", "EULERS_NUMBER", ")",
			"(", "=", "1", "2", ")",
		}
		_, err := Parse(tokens)
		if err == nil {
			t.Fatalf("Parse failed: did not receive expected error")
		}
	}

	{ // test overcomplete tokens
		tokens := []string{
			"(", "+", "(", "*", "1", "(", "/", "1",
			"zero", ")", ")", "(", "/", "2", "PI", ")",
			"(", "-", "thirty-three", "EULERS_NUMBER", ")", ")", ")",
			"(", "=", "1", "2", ")",
		}
		_, err := Parse(tokens)
		if err == nil {
			t.Fatalf("Parse failed: did not receive expected error")
		}
	}
}

func TestExec(t *testing.T) {
	srcTable := map[string]interface{}{
		`(+ 5 2)`:                    7,
		`(+ (* 1 (/ 10 5)) (- 5 2))`: 5,
		`5`:                                               5,
		`1 2 3 4 5 42`:                                    42,
		`(define a 42) a`:                                 42,
		`(define f (lambda (a b c) (* a b c))) (f 1 2 3)`: 6,
		`(> 4 1)`:                                                     true,
		`(if (> 4 1) 16 0)`:                                           16,
		`(define l (list 1 2 3 4)) (car (cdr (cdr l)))`:               3,
		`(begin (define a 10) (define b 23) (+ a b))`:                 33,
		`(let ((a 10) (b 23)) (+ a b))`:                               33,
		`(define l (list 1 2 3 4)) (null? (cdr (cdr (cdr (cdr l)))))`: true,
		`(define l (list 1 2 3 4)) (null? (cdr (cdr (cdr l))))`:       false,
		`(= 1 1)`: true,
		`#t`:      true,
		`#f`:      false,
		`(or (= 1 2) (= 3 4))`:         false,
		`(not (= 1 1))`:                false,
		`(and (= 1 1) (= 3 (+ 1 2)))`:  true,
		`(define a (lambda () 1)) (a)`: 1,
		`(remainder 33 7)`:             5,

		`
		(define fib (lambda (n)
          (cond ((= n 0) 0)
                ((= n 1) 1)
                (else (+ (fib (- n 1)) (fib (- n 2)))))))
        (fib 13)`: 233,

		`
		(define even? (lambda (x) (= (remainder x 2) 0)))

		(define square (lambda (x) (* x x)))

		(define expt (lambda (b n)
		  (cond ((= n 0) 1)
		        ((even? n) (square (expt b (/ n 2))))
		        (else (* b (expt b (- n 1)))))))

		(expt 3 8)`: 6561,

		`
		; this is a comment
			(and
				(= 1 1)
				; and another comment
				(= 3 (+ 1 2)))`: true,

		`; SICP 1.2.6, p 51

		(define even?
		  (lambda (x) (= (remainder x 2) 0)))

		(define square
		  (lambda (x) (* x x)))

		(define expmod (lambda (base exp m)
		  (cond ((= exp 0) 1)
		        ((even? exp)
		         (remainder (square (expmod base (/ exp 2) m))
		                    m))
		        (else
		         (remainder (* base (expmod base (- exp 1) m))
		                    m)))))

		(define fermat-test (lambda (n)
		  (let
		      ((try-it
		        (lambda (a) (= (expmod a n n) a))))
		    (try-it (+ 1 (random (- n 1)))))))

		(define fast-prime?
		  (lambda (n times)
		    (cond ((= times 0) #t)
		          ((fermat-test n) (fast-prime? n (- times 1)))
		          (else #f))))

		(fast-prime? 10000 100)`: false,
	}

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println(err)
	// 		t.Fatalf("Exec failed")
	// 	}
	// }()

	for k, v := range srcTable {
		res, err := Exec(k)
		if err != nil {
			t.Fatalf(`Exec returned unexpected error: #v`, err)
		}
		if res != v {
			t.Fatalf(`Exec
	src: %s

	expected: %v
	got:      %v`, k, v, res)
		}
	}
}

func stringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sliceEquals(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		switch aVal := a[i].(type) {
		case []interface{}:
			if bVal, ok := b[i].([]interface{}); ok {
				if !sliceEquals(aVal, bVal) {
					return false
				}
			} else {
				return false
			}
		case string:
			if bVal, ok := b[i].(string); ok {
				if aVal != bVal {
					return false
				}
			} else {
				return false
			}
		default:
			panic(fmt.Sprintf("sliceEquals received bad type: %T", a[i]))
		}
	}
	return true
}

func listEquals(a, b *list.List) bool {
	if a.Len() != b.Len() {
		return false
	}
	for e1, e2 := a.Front(), b.Front(); e1 != nil; e1, e2 = e1.Next(), e2.Next() {
		switch e1.Value.(type) {
		case *list.List:
			if _, ok := e2.Value.(*list.List); ok {
				if !listEquals(e1.Value.(*list.List), e2.Value.(*list.List)) {
					return false
				}
			} else {
				return false
			}
		case string:
			if e1.Value != e2.Value {
				return false
			}
		default:
			panic("listEquals: got a type that is not string or *list.List")
		}
	}
	return true
}

func createList(elems []interface{}) *list.List {
	l := list.New()
	for _, e := range elems {
		switch val := e.(type) {
		case []interface{}:
			l.PushBack(createList(val))
		case string:
			l.PushBack(val)
		default:
			panic("createList: received type that was not string or []interface{}")
		}
	}
	return l
}

func listToString(l *list.List) string {
	if l.Len() == 0 {
		return "[]"
	} else {
		res := "["
		var e *list.Element
		for e = l.Front(); e.Next() != nil; e = e.Next() {
			if nl, ok := e.Value.(*list.List); ok {
				res += listToString(nl) + ", "
			} else {
				res += fmt.Sprintf("`%v`, ", e.Value)
			}
		}
		if nl, ok := e.Value.(*list.List); ok {
			res += listToString(nl)
		} else {
			res += fmt.Sprintf("`%v`", e.Value)
		}
		res += "]"
		return res
	}
}
