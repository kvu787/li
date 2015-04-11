; Use Horner's rule and list accumulation to evaluate polynomials

; From "Structure and Interpretation of Computer Programs, Second Edition"
; By Harold Abelson, Gerald Jay Sussman with Julie Sussman
; Page 119
; Section 2.2.3

(define accumulate
  (lambda (op initial sequence)
    (if (null? sequence)
      initial
      (op (car sequence)
          (accumulate op initial (cdr sequence))))))

(define horner-eval
  (lambda (x coefficient-sequence)
    (accumulate (lambda (this-coeff higher-terms)
                        (+ this-coeff (* x higher-terms)))
                0
                coefficient-sequence)))

(horner-eval 2 (list 1 3 0 5 0 1)) ; => 79