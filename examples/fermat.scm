; Use Fermat's little theorem to develop a fast, probabilistic algorithm
; for checking integer primality.

; From "Structure and Interpretation of Computer Programs, Second Edition"
; By Harold Abelson, Gerald Jay Sussman with Julie Sussman
; Page 51
; Section 1.2.6

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

(fast-prime? 999995 10) ; => false
