(define even? (lambda (x) (= (remainder x 2) 0)))

(define square (lambda (x) (* x x)))

(define expt (lambda (b n)
  (cond ((= n 0) 1)
        ((even? n) (square (expt b (/ n 2))))
        (else (* b (expt b (- n 1)))))))

(expt 3 8)
