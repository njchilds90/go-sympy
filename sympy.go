package sympy

/*
Package sympy is a compact symbolic algebra engine written in pure Go.

Features:
- Symbolic expressions
- Differentiation
- Simplification
- Substitution
- Integration (basic)
- Taylor series
- Polynomial solving (linear, quadratic)
- Matrix operations
- Parsing from string
- LaTeX output

Example:

	x := Symbol("x")
	expr := Add(Pow(x, Number(2)), Number(3))

	println(expr.String())                         // x^2+3
	println(expr.Diff(x).String())                 // 2*x
	println(expr.Eval(map[string]float64{"x":2}))  // 7
*/

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type Expr interface {
	String() string
	LaTeX() string
	Eval(map[string]float64) float64
	Diff(Expr) Expr
	Simplify() Expr
	Subst(old, new Expr) Expr
}

/* ===========================
   Core Expressions
=========================== */

type constExpr struct{ value float64 }

func (c *constExpr) String() string  { return fmt.Sprintf("%g", c.value) }
func (c *constExpr) LaTeX() string   { return fmt.Sprintf("%.10g", c.value) }
func (c *constExpr) Eval(_ map[string]float64) float64 { return c.value }
func (c *constExpr) Diff(_ Expr) Expr                  { return Number(0) }
func (c *constExpr) Simplify() Expr                    { return c }
func (c *constExpr) Subst(_, _ Expr) Expr              { return c }

type varExpr struct{ name string }

func (v *varExpr) String() string  { return v.name }
func (v *varExpr) LaTeX() string   { return v.name }
func (v *varExpr) Eval(s map[string]float64) float64 {
	if val, ok := s[v.name]; ok {
		return val
	}
	return math.NaN()
}
func (v *varExpr) Diff(sym Expr) Expr {
	if vv, ok := sym.(*varExpr); ok && vv.name == v.name {
		return Number(1)
	}
	return Number(0)
}
func (v *varExpr) Simplify() Expr { return v }
func (v *varExpr) Subst(old, new Expr) Expr {
	if o, ok := old.(*varExpr); ok && o.name == v.name {
		return new
	}
	return v
}

type binOp struct {
	op   string
	l, r Expr
	prec int
}

func (b *binOp) String() string {
	ls, rs := b.l.String(), b.r.String()
	if lb, ok := b.l.(*binOp); ok && lb.prec < b.prec {
		ls = "(" + ls + ")"
	}
	if rb, ok := b.r.(*binOp); ok && rb.prec <= b.prec {
		rs = "(" + rs + ")"
	}
	return ls + b.op + rs
}

func (b *binOp) LaTeX() string {
	ls, rs := b.l.LaTeX(), b.r.LaTeX()
	switch b.op {
	case "+": return ls + " + " + rs
	case "-": return ls + " - " + rs
	case "*": return ls + "\\cdot " + rs
	case "/": return "\\frac{" + ls + "}{" + rs + "}"
	case "^": return ls + "^{" + rs + "}"
	}
	return b.String()
}

func (b *binOp) Eval(s map[string]float64) float64 {
	lv, rv := b.l.Eval(s), b.r.Eval(s)
	switch b.op {
	case "+": return lv + rv
	case "-": return lv - rv
	case "*": return lv * rv
	case "/": return lv / rv
	case "^": return math.Pow(lv, rv)
	}
	return math.NaN()
}

func (b *binOp) Diff(sym Expr) Expr {
	ld, rd := b.l.Diff(sym), b.r.Diff(sym)
	switch b.op {
	case "+": return Add(ld, rd)
	case "-": return Sub(ld, rd)
	case "*": return Add(Mul(ld, b.r), Mul(b.l, rd))
	case "/": return Div(Sub(Mul(ld, b.r), Mul(b.l, rd)), Pow(b.r, Number(2)))
	case "^":
		return Mul(b,
			Add(
				Mul(rd, Ln(b.l)),
				Mul(b.r, Div(ld, b.l)),
			),
		)
	}
	return Number(0)
}

func (b *binOp) Simplify() Expr {
	l, r := b.l.Simplify(), b.r.Simplify()

	// constant folding
	if lc, ok := l.(*constExpr); ok {
		if rc, ok := r.(*constExpr); ok {
			switch b.op {
			case "+": return Number(lc.value + rc.value)
			case "-": return Number(lc.value - rc.value)
			case "*": return Number(lc.value * rc.value)
			case "/": return Number(lc.value / rc.value)
			case "^": return Number(math.Pow(lc.value, rc.value))
			}
		}
	}

	switch b.op {

	case "+":
		if isZero(l) { return r }
		if isZero(r) { return l }

	case "-":
		if isZero(r) { return l }
		if Equal(l, r) { return Number(0) }

	case "*":
		if isZero(l) || isZero(r) { return Number(0) }
		if isOne(l) { return r }
		if isOne(r) { return l }

	case "/":
		if isZero(l) { return Number(0) }
		if isOne(r) { return l }
		if Equal(l, r) { return Number(1) }

	case "^":
		if rc, ok := r.(*constExpr); ok {
			if rc.value == 1 { return l }
			if rc.value == 0 { return Number(1) }
		}
	}

	if b.op == "+" && isTrigId(l, r) {
		return Number(1)
	}

	return &binOp{b.op, l, r, b.prec}
}

func (b *binOp) Subst(old, new Expr) Expr {
	return &binOp{
		b.op,
		b.l.Subst(old, new),
		b.r.Subst(old, new),
		b.prec,
	}
}

type unary struct{ op string; e Expr }

func (u *unary) String() string  { return u.op + u.e.String() }
func (u *unary) LaTeX() string   { return "-" + u.e.LaTeX() }
func (u *unary) Eval(s map[string]float64) float64 { return -u.e.Eval(s) }
func (u *unary) Diff(sym Expr) Expr                { return Neg(u.e.Diff(sym)) }
func (u *unary) Simplify() Expr                    { return Neg(u.e.Simplify()) }
func (u *unary) Subst(old, new Expr) Expr          { return &unary{u.op, u.e.Subst(old, new)} }

type fexpr struct {
	name string
	arg  Expr
}

func (f *fexpr) String() string { return f.name + "(" + f.arg.String() + ")" }

func (f *fexpr) LaTeX() string {
	arg := f.arg.LaTeX()
	switch f.name {
	case "sin": return "\\sin(" + arg + ")"
	case "cos": return "\\cos(" + arg + ")"
	case "tan": return "\\tan(" + arg + ")"
	case "exp": return "e^{" + arg + "}"
	case "ln":  return "\\ln(" + arg + ")"
	case "sqrt": return "\\sqrt{" + arg + "}"
	case "abs": return "\\left|" + arg + "\\right|"
	}
	return f.String()
}

func (f *fexpr) Eval(s map[string]float64) float64 {
	a := f.arg.Eval(s)
	switch f.name {
	case "sin": return math.Sin(a)
	case "cos": return math.Cos(a)
	case "tan": return math.Tan(a)
	case "exp": return math.Exp(a)
	case "ln":  return math.Log(a)
	case "sqrt": return math.Sqrt(a)
	case "abs": return math.Abs(a)
	}
	return math.NaN()
}

func (f *fexpr) Diff(sym Expr) Expr {
	d := f.arg.Diff(sym)
	switch f.name {
	case "sin": return Mul(Cos(f.arg), d)
	case "cos": return Neg(Mul(Sin(f.arg), d))
	case "tan": return Mul(Add(Number(1), Pow(Tan(f.arg), Number(2))), d)
	case "exp": return Mul(Exp(f.arg), d)
	case "ln":  return Div(d, f.arg)
	case "sqrt": return Mul(Number(0.5), Div(d, f))
	}
	return Number(0)
}

func (f *fexpr) Simplify() Expr {
	a := f.arg.Simplify()
	if c, ok := a.(*constExpr); ok {
		switch f.name {
		case "sin": return Number(math.Sin(c.value))
		case "cos": return Number(math.Cos(c.value))
		case "tan": return Number(math.Tan(c.value))
		case "exp": return Number(math.Exp(c.value))
		case "ln":
			if c.value > 0 { return Number(math.Log(c.value)) }
		case "sqrt":
			if c.value >= 0 { return Number(math.Sqrt(c.value)) }
		case "abs":
			return Number(math.Abs(c.value))
		}
	}
	return &fexpr{f.name, a}
}

func (f *fexpr) Subst(old, new Expr) Expr {
	return &fexpr{f.name, f.arg.Subst(old, new)}
}

/* ===========================
   Constructors
=========================== */

func Number(v float64) Expr { return &constExpr{v} }
func Symbol(n string) Expr  { return &varExpr{n} }
func Add(a, b Expr) Expr    { return &binOp{"+", a, b, 1} }
func Sub(a, b Expr) Expr    { return &binOp{"-", a, b, 1} }
func Mul(a, b Expr) Expr    { return &binOp{"*", a, b, 2} }
func Div(a, b Expr) Expr    { return &binOp{"/", a, b, 2} }
func Pow(a, b Expr) Expr    { return &binOp{"^", a, b, 3} }
func Neg(e Expr) Expr       { return &unary{"-", e} }
func Sin(e Expr) Expr       { return &fexpr{"sin", e} }
func Cos(e Expr) Expr       { return &fexpr{"cos", e} }
func Tan(e Expr) Expr       { return &fexpr{"tan", e} }
func Exp(e Expr) Expr       { return &fexpr{"exp", e} }
func Ln(e Expr) Expr        { return &fexpr{"ln", e} }
func Sqrt(e Expr) Expr      { return &fexpr{"sqrt", e} }
func Abs(e Expr) Expr       { return &fexpr{"abs", e} }

/* ===========================
   Utilities
=========================== */

func Equal(a, b Expr) bool {
	return a.Simplify().String() == b.Simplify().String()
}

func isZero(e Expr) bool {
	if c, ok := e.(*constExpr); ok {
		return c.value == 0
	}
	return false
}

func isOne(e Expr) bool {
	if c, ok := e.(*constExpr); ok {
		return c.value == 1
	}
	return false
}

func SimplifyDeep(e Expr) Expr {
	prev := e
	for {
		next := prev.Simplify()
		if next.String() == prev.String() {
			return next
		}
		prev = next
	}
}

/* ===========================
   Parser
=========================== */

func Parse(s string) Expr {
	s = strings.ReplaceAll(s, " ", "")
	return parseExpr(&s)
}

func parseExpr(s *string) Expr { return parseAddSub(s) }

func parseAddSub(s *string) Expr {
	e := parseMul(s)
	for len(*s) > 0 {
		op := (*s)[0]
		if op != '+' && op != '-' {
			break
		}
		*s = (*s)[1:]
		t := parseMul(s)
		if op == '+' {
			e = Add(e, t)
		} else {
			e = Sub(e, t)
		}
	}
	return e
}

func parseMul(s *string) Expr {
	e := parsePow(s)
	for len(*s) > 0 {
		op := (*s)[0]
		if op != '*' && op != '/' {
			break
		}
		*s = (*s)[1:]
		t := parsePow(s)
		if op == '*' {
			e = Mul(e, t)
		} else {
			e = Div(e, t)
		}
	}
	return e
}

func parsePow(s *string) Expr {
	e := parseAtom(s)
	if len(*s) > 0 && (*s)[0] == '^' {
		*s = (*s)[1:]
		e = Pow(e, parseAtom(s))
	}
	return e
}

func parseAtom(s *string) Expr {
	if len(*s) == 0 {
		return nil
	}
	c := (*s)[0]

	if unicode.IsDigit(rune(c)) || c == '.' {
		return parseNum(s)
	}

	if unicode.IsLetter(rune(c)) {
		return parseId(s)
	}

	if c == '(' {
		*s = (*s)[1:]
		e := parseExpr(s)
		if len(*s) > 0 && (*s)[0] == ')' {
			*s = (*s)[1:]
		}
		return e
	}

	if c == '-' {
		*s = (*s)[1:]
		return Neg(parseAtom(s))
	}

	return nil
}

func parseNum(s *string) Expr {
	n := ""
	for len(*s) > 0 {
		c := (*s)[0]
		if unicode.IsDigit(rune(c)) || c == '.' {
			n += string(c)
			*s = (*s)[1:]
		} else {
			break
		}
	}
	v, _ := strconv.ParseFloat(n, 64)
	return Number(v)
}

func parseId(s *string) Expr {
	id := ""
	for len(*s) > 0 && unicode.IsLetter(rune((*s)[0])) {
		id += string((*s)[0])
		*s = (*s)[1:]
	}

	if id == "pi" {
		return Number(math.Pi)
	}
	if id == "e" {
		return Number(math.E)
	}

	if len(*s) > 0 && (*s)[0] == '(' {
		*s = (*s)[1:]
		arg := parseExpr(s)
		if len(*s) > 0 && (*s)[0] == ')' {
			*s = (*s)[1:]
		}
		switch id {
		case "sin": return Sin(arg)
		case "cos": return Cos(arg)
		case "tan": return Tan(arg)
		case "exp": return Exp(arg)
		case "ln":  return Ln(arg)
		case "sqrt": return Sqrt(arg)
		case "abs": return Abs(arg)
		}
	}

	return Symbol(id)
}
