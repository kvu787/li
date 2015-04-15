# li

Li evaulates Scheme (Lisp) expressions.

## Installation

```
go get github.com/kvu787/li
```

## Usage

```
li [-h | -i]
```

Li evaulates Scheme (Lisp) expressions.

If no flags are specified, expressions are read from standard input and
evaluated.
The value of the final expression is printed to standard output.

The `-i` flag launches an interactive read-evaluate-print-loop (REPL)
interpreter. Expressions are read from standard input. The result of each
expression is printed to standard output until EOF is encountered or an
error occurs.

## Examples

### Running Scheme programs

The `examples` directory contains several Scheme programs that `li` can
evaulate. You can evaluate a term of the Fibonacci sequence by piping
`fib.scm` into `li`:

```
cat examples/fib.scm | li
```

### Running the REPL

`li` runs in interactive mode when given the `-i` option. Below is the
transcript of a sample session that evaluates Fibonacci terms:

```
$ li -i
(define fib (lambda (n)
  (cond ((= n 0) 0)
        ((= n 1) 1)
        (else (+ (fib (- n 1)) (fib (- n 2)))))))
>>> <nil>
(fib 0)
>>> 0
(fib 2)
>>> 1
(fib 10)
>>> 55
(fib 6)
>>> 8
```