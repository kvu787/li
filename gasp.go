package main

import (
	"fmt"
	"regexp"
)

var src = `(define (union-set s1 s2)
  (cond ((and (null? s1) (null? s2)) '())
        ((null? s1) s2)
        ((null? s2) s1)
        (else
          (let (
                (e1 (car s1))
                (e2 (car s2))
                (rest (union-set (cdr s1) (cdr s2))))
            (cond ((< e1 e2) (cons e1 (cons e2 rest)))
                  ((= e1 e2) (cons e1 rest))
                  ((> e1 e2) (cons e2 (cons e1 rest))))))))`

var src2 = `(+ (* 1 2 ) (/ 2 chicken) (- he-llo the_re))`

func main() {
	fmt.Printf("%q\n", lex(src2))
}

func lex(src string) []string {
	parens := `[(]|[)]`
	numbers := `\d+`
	operators := `\+|\-|\*|/`
	identifiers := `(\w|\-)+`
	re := regexp.MustCompile(parens + "|" + numbers + "|" + operators + "|" + identifiers)
	matches := re.FindAllString(src, -1)
	return matches
}
