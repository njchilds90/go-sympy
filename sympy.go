// Package sympy provides a compact, deterministic symbolic math kernel in pure Go.
//
// Design Goals:
//   - Single file
//   - Zero external dependencies
//   - Deterministic simplification (stable ordering)
//   - Exact rational arithmetic
//   - AI-friendly embedding
//   - Rule-based symbolic manipulation
//
// Limitations:
//   - No advanced polynomial factoring
//   - No symbolic matrix inversion
//   - No symbolic limits beyond substitution
//   - Simplification is rule-based, not canonical
//   - Integration is pattern-based (no Risch algorithm)
//   - No expression parser (AST built programmatically)
package sympy

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
)

/* =======================
   Rational
======================= */

// Rational wraps big.Rat to provide exact rational arithmetic.
type Rational struct{ *big.Rat }

// NewInt returns a new Rational from an int64.
func NewInt(n int64) Rational { return Rational{big.NewRat(n, 1)} }

// NewFrac returns a new Rational representing a/b.
func NewFrac(a, b int64) Rational { return Rational{big.NewRat(a, b)} }

// Zero returns the rational number 0.
func Zero() Rational { return NewInt(0) }

// One returns the rational number 1.
func One() Rational { return NewInt(1) }

// Add returns r + o.
func (r Rational) Add(o Rational) Rational { return Rational{new(big.Rat).Add(r.Rat, o.Rat)} }

// Sub returns r - o.
func (r Rational) Sub(o Rational) Rational { return Rational{new(big.Rat).Sub(r.Rat, o.Rat)} }

// Mul returns r * o.
func (r Rational) Mul(o Rational) Rational { return Rational{new(big.Rat).Mul(r.Rat, o.Rat)} }

// Div returns r / o.
func (r Rational) Div(o Rational) Rational { return Rational{new(big.Rat).Quo(r.Rat, o.Rat)} }

// Neg returns -r.
func (r Rational) Neg() Rational { return Rational{new(big.Rat).Neg(r.Rat)} }

// IsZero reports whether r == 0.
func (r Rational) IsZero() bool { return r.Sign() == 0 }

// String returns the canonical rational string form.
func (r Rational) String() string { return r.Rat.RatString() }

/* =======================
   Expr Core
======================= */

// Expr represents a symbolic expression node.
type Expr interface {
	Simplify() Expr
	String() string
	Sub(varName string, value Expr) Expr
}

// Simplify simplifies an expression using rule-based reduction.
func Simplify(e Expr) Expr { return e.Simplify() }

// String returns the deterministic string form of a simplified expression.
func String(e Expr) string { return e.Simplify().String() }

/* ---------- Num ---------- */

// Num represents a rational numeric literal.
type Num struct{ V Rational }

// N creates an integer literal expression.
func N(n int64) Expr { return Num{NewInt(n)} }

// F creates a rational literal expression a/b.
func F(a, b int64) Expr { return Num{NewFrac(a, b)} }

// Simplify returns the numeric literal unchanged.
func (n Num) Simplify() Expr { return n }

// String returns the rational string form.
func (n Num) String() string { return n.V.String() }

// Sub performs substitution (no-op for numbers).
func (n Num) Sub(string, Expr) Expr { return n }

/* ---------- Sym ---------- */

// Sym represents a symbolic variable.
type Sym struct{ Name string }

// S creates a new symbol expression.
func S(name string) Expr { return Sym{Name: name} }

// Simplify returns the symbol unchanged.
func (s Sym) Simplify() Expr { return s }

// String returns the symbol name.
func (s Sym) String() string { return s.Name }

// Sub substitutes the symbol if it matches varName.
func (s Sym) Sub(v string, val Expr) Expr {
	if s.Name == v {
		return val
	}
	return s
}

/* ---------- Add ---------- */

// Add represents a sum of terms.
type Add struct{ Terms []Expr }

// AddOf constructs and simplifies an addition.
func AddOf(terms ...Expr) Expr { return Add{terms}.Simplify() }

// Simplify flattens nested additions, combines numeric terms,
// and enforces deterministic ordering.
func (a Add) Simplify() Expr {
	var flat []Expr
	sum := Zero()

	for _, t := range a.Terms {
		t = t.Simplify()
		switch v := t.(type) {
		case Add:
			flat = append(flat, v.Terms...)
		case Num:
			sum = sum.Add(v.V)
		default:
			flat = append(flat, t)
		}
	}

	if !sum.IsZero() {
		flat = append(flat, Num{sum})
	}

	if len(flat) == 0 {
		return Num{Zero()}
	}
	if len(flat) == 1 {
		return flat[0]
	}

	sort.Slice(flat, func(i, j int) bool {
		return flat[i].String() < flat[j].String()
	})

	return Add{flat}
}

// String returns a deterministic string representation of the sum.
func (a Add) String() string {
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = t.String()
	}
	return strings.Join(parts, " + ")
}

// Sub substitutes within each term.
func (a Add) Sub(v string, val Expr) Expr {
	var out []Expr
	for _, t := range a.Terms {
		out = append(out, t.Sub(v, val))
	}
	return AddOf(out...)
}

/* ---------- Mul ---------- */

// Mul represents a product of factors.
type Mul struct{ Factors []Expr }

// MulOf constructs and simplifies a multiplication.
func MulOf(factors ...Expr) Expr { return Mul{factors}.Simplify() }

// Simplify flattens nested multiplications, combines numeric factors,
// and enforces deterministic ordering.
func (m Mul) Simplify() Expr {
	var flat []Expr
	prod := One()

	for _, f := range m.Factors {
		f = f.Simplify()
		switch v := f.(type) {
		case Mul:
			flat = append(flat, v.Factors...)
		case Num:
			prod = prod.Mul(v.V)
		default:
			flat = append(flat, f)
		}
	}

	if prod.IsZero() {
		return Num{Zero()}
	}
	if prod.Cmp(big.NewRat(1, 1)) != 0 {
		flat = append(flat, Num{prod})
	}

	if len(flat) == 0 {
		return Num{prod}
	}
	if len(flat) == 1 {
		return flat[0]
	}

	sort.Slice(flat, func(i, j int) bool {
		return flat[i].String() < flat[j].String()
	})

	return Mul{flat}
}

// String returns a deterministic string representation of the product.
func (m Mul) String() string {
	parts := make([]string, len(m.Factors))
	for i, f := range m.Factors {
		parts[i] = f.String()
	}
	return strings.Join(parts, "*")
}

// Sub substitutes within each factor.
func (m Mul) Sub(v string, val Expr) Expr {
	var out []Expr
	for _, f := range m.Factors {
		out = append(out, f.Sub(v, val))
	}
	return MulOf(out...)
}

/* ---------- Pow ---------- */

// Pow represents exponentiation Base^Exp.
type Pow struct{ Base, Exp Expr }

// PowOf constructs and simplifies a power expression.
func PowOf(b, e Expr) Expr { return Pow{b, e}.Simplify() }

// Simplify applies simple exponent rules (x^0 = 1, x^1 = x).
func (p Pow) Simplify() Expr {
	b := p.Base.Simplify()
	e := p.Exp.Simplify()

	if en, ok := e.(Num); ok {
		if en.V.IsZero() {
			return Num{One()}
		}
		if en.V.Cmp(big.NewRat(1, 1)) == 0 {
			return b
		}
	}
	return Pow{b, e}
}

// String returns a string representation of the power.
func (p Pow) String() string {
	return fmt.Sprintf("(%s)^%s", p.Base, p.Exp)
}

// Sub substitutes within base and exponent.
func (p Pow) Sub(v string, val Expr) Expr {
	return PowOf(p.Base.Sub(v, val), p.Exp.Sub(v, val))
}

/* =======================
   Polynomial Utilities
======================= */

// Degree returns the degree of expression e with respect to variable v.
// Non-polynomial parts are treated conservatively.
func Degree(e Expr, v string) int {
	switch t := e.(type) {
	case Num:
		return 0
	case Sym:
		if t.Name == v {
			return 1
		}
		return 0
	case Add:
		max := 0
		for _, term := range t.Terms {
			if d := Degree(term, v); d > max {
				max = d
			}
		}
		return max
	case Mul:
		sum := 0
		for _, f := range t.Factors {
			sum += Degree(f, v)
		}
		return sum
	case Pow:
		if base, ok := t.Base.(Sym); ok && base.Name == v {
			if exp, ok := t.Exp.(Num); ok {
				i, _ := exp.V.Int64()
				return int(i)
			}
		}
	}
	return 0
}

// PolyCoeffs extracts polynomial coefficients as map[degree]Rational.
func PolyCoeffs(e Expr, v string) map[int]Rational {
	coeffs := map[int]Rational{}

	var collect func(Expr)
	collect = func(ex Expr) {
		switch t := ex.(type) {
		case Add:
			for _, term := range t.Terms {
				collect(term)
			}
		case Mul:
			deg := Degree(t, v)
			c := One()
			for _, f := range t.Factors {
				if n, ok := f.(Num); ok {
					c = c.Mul(n.V)
				}
			}
			coeffs[deg] = coeffs[deg].Add(c)
		case Pow:
			deg := Degree(t, v)
			coeffs[deg] = coeffs[deg].Add(One())
		case Sym:
			if t.Name == v {
				coeffs[1] = coeffs[1].Add(One())
			}
		case Num:
			coeffs[0] = coeffs[0].Add(t.V)
		}
	}

	collect(e.Simplify())
	return coeffs
}

/* =======================
   Solvers
======================= */

// SolveLinear solves ax + b = 0 and returns the exact rational root.
func SolveLinear(a, b Rational) Rational {
	return b.Neg().Div(a)
}

// SolveQuadratic solves ax^2 + bx + c = 0 numerically (float64).
func SolveQuadratic(a, b, c float64) []float64 {
	d := b*b - 4*a*c
	if d < 0 {
		return nil
	}
	s := math.Sqrt(d)
	return []float64{
		(-b + s) / (2 * a),
		(-b - s) / (2 * a),
	}
}

/* =======================
   Integration
======================= */

// Integrate performs rule-based integration with respect to variable v.
// Supports:
//   - Constant rule
//   - Power rule
//   - Sum rule
//   - Constant multiple rule
func Integrate(e Expr, v string) Expr {
	switch t := e.(type) {
	case Num:
		return MulOf(t, S(v))
	case Sym:
		if t.Name == v {
			return MulOf(F(1, 2), PowOf(t, N(2)))
		}
		return MulOf(t, S(v))
	case Add:
		var parts []Expr
		for _, term := range t.Terms {
			parts = append(parts, Integrate(term, v))
		}
		return AddOf(parts...)
	case Mul:
		if len(t.Factors) == 2 {
			if c, ok := t.Factors[0].(Num); ok {
				return MulOf(c, Integrate(t.Factors[1], v))
			}
		}
	case Pow:
		if base, ok := t.Base.(Sym); ok && base.Name == v {
			if exp, ok := t.Exp.(Num); ok {
				n := exp.V
				newExp := n.Add(One())
				return MulOf(
					Num{One().Div(newExp)},
					PowOf(base, Num{newExp}),
				)
			}
		}
	}
	return nil
}
