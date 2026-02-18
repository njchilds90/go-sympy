# go-sympy

Minimal symbolic math library for Go — single file, zero dependencies. Inspired by SymPy.

Supports symbolic expressions, arithmetic (+ − × / ^ negation), trig/exp/log/sqrt functions, precedence-aware printing, LaTeX output, differentiation (general rules), basic simplification, string parsing, numerical evaluation, simple expansion, basic integration, and linear/quadratic solving.

**Early prototype** — all development on Feb 18, 2026.

## Installation

```bash
go get github.com/njchilds90/go-sympy
How to Use the API

Create variables: x := sympy.Symbol("x")
Create constants: sympy.Number(5)
Build expressions using these constructors:
Add(a, b) → a + b
Sub(a, b) → a - b
Mul(a, b) → a * b
Div(a, b) → a / b (uses ^-1 internally)
Pow(a, b) → a ^ b
Neg(e) → -e
Sin(e), Cos(e), Exp(e), Ln(e), Sqrt(e)

Parse string: sympy.Parse("x^2 + 3*sin(x)")
Differentiate: expr.Diff(x)
Simplify: expr.Simplify()
Evaluate numerically: expr.Eval(map[string]float64{"x": 2.0})
Print: expr.String() or expr.LaTeX()
Expand: sympy.Expand(expr)
Integrate: sympy.Integrate(expr, x)
Solve (equation = 0): sympy.Solve(equation, x) → returns []Expr (roots)

Usage Examples
Gopackage main

import (
	"fmt"
	"github.com/njchilds90/go-sympy"
)

func main() {
	x := sympy.Symbol("x")

	// Expression: x² + 3x
	expr := sympy.Add(sympy.Pow(x, sympy.Number(2)), sympy.Mul(sympy.Number(3), x))
	fmt.Println("Expression:", expr.String())          // x^2 + 3*x
	fmt.Println("LaTeX:", expr.LaTeX())                 // x^{2} + 3 \cdot x
	fmt.Println("Derivative:", expr.Diff(x).String())   // 2*x + 3

	// Trig identity simplification
	trig := sympy.Add(sympy.Pow(sympy.Sin(x), sympy.Number(2)), sympy.Pow(sympy.Cos(x), sympy.Number(2)))
	fmt.Println("sin² + cos² →", trig.Simplify().String()) // 1

	// Parsing from string
	p := sympy.Parse("x^2 + sin(x)^2 + cos(x)^2 - 5")
	fmt.Println("Parsed simplified:", p.Simplify().String()) // x^2 - 4

	// Numerical evaluation
	fmt.Printf("Eval at x=2: %.1f\n", expr.Eval(map[string]float64{"x": 2})) // 10.0

	// Quadratic solve: x² - 5x + 6 = 0
	quad := sympy.Sub(sympy.Pow(x, sympy.Number(2)), sympy.Add(sympy.Mul(sympy.Number(5), x), sympy.Number(6)))
	sols := sympy.Solve(quad, x)
	fmt.Println("Roots:", sols[0].String(), sols[1].String()) // e.g. 3 2
}
Features

Variables & constants
Operations: +, −, ×, / (via ^-1), ^, unary −
Functions: sin, cos, exp, ln, sqrt
Precedence-aware printing & LaTeX
Differentiation (general rules)
Basic simplification (constants, identities, trig)
String parsing
Numerical evaluation
Basic expansion & integration
Linear + quadratic solving

Limitations
Early prototype — missing deep simplification, full integration, higher-degree solving, multi-var partials, series, matrices, limits, exact rationals. No unit tests yet.
Contributing
Fork → branch → PR. Tests, more rules, docs welcome.
