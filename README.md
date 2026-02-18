# go-sympy

Minimal symbolic math library for Go (single-file, no deps). Inspired by SymPy.

Supports symbolic expressions, arithmetic, trig/exp/log/sqrt functions, precedence-aware printing, LaTeX, differentiation (general rules), basic simplification, string parsing, numerical evaluation, simple expansion, basic integration, linear/quadratic solving.

**Early prototype** — developed Feb 18, 2026.

## Installation

```bash
go get github.com/njchilds90/go-sympy

-----------------------------------------------------

Features

Variables & constants
+, −, ×, /, ^, unary −
sin, cos, exp, ln, sqrt
Precedence-aware printing
LaTeX output
Differentiation (product/chain/general power)
Basic simplification (constants, identities, trig)
String parsing
Numerical evaluation
Basic expansion (binomial + distribution)
Basic indefinite integration
Linear + quadratic solving

-----------------------------------------------------

package main

import (
	"fmt"
	"github.com/njchilds90/go-sympy"
)

func main() {
	x := sympy.Symbol("x")

	expr := sympy.Add(sympy.Pow(x, sympy.Number(2)), sympy.Mul(sympy.Number(3), x))
	fmt.Println("Expr:", expr.String())          // x^2 + 3*x
	fmt.Println("LaTeX:", expr.LaTeX())           // x^{2} + 3x
	fmt.Println("Diff:", expr.Diff(x).String())   // 2*x + 3

	trig := sympy.Add(sympy.Pow(sympy.Sin(x), sympy.Number(2)), sympy.Pow(sympy.Cos(x), sympy.Number(2)))
	fmt.Println("Trig simp:", trig.Simplify().String()) // 1

	parsed := sympy.Parse("x^2 + sin(x)^2 + cos(x)^2 - 5")
	fmt.Println("Parsed simp:", parsed.Simplify().String()) // x^2 - 4

	fmt.Printf("Eval x=2: %.1f\n", expr.Eval(map[string]float64{"x": 2})) // 10.0

	quad := sympy.Sub(sympy.Pow(x, sympy.Number(2)), sympy.Add(sympy.Mul(sympy.Number(5), x), sympy.Number(6)))
	sols := sympy.Solve(quad, x)
	fmt.Println("Roots:", sols[0].String(), sols[1].String()) // e.g. 3 2
}
