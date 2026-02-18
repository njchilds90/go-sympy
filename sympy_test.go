package sympy

import "testing"

func TestDiff(t *testing.T) {
	x := Symbol("x")
	tests := []struct {
		expr Expr
		want string
	}{
		{Pow(x, Number(2)), "2*x"},
		{Mul(x, Number(3)), "3"},
		{Sin(x), "cos(x)"},
		{Add(Pow(x, Number(2)), Mul(Number(3), x)), "2*x + 3"},
	}
	for _, tt := range tests {
		got := tt.expr.Diff(x).Simplify().String()
		if got != tt.want {
			t.Errorf("Diff(%s) = %s, want %s", tt.expr, got, tt.want)
		}
	}
}

func TestSimplify(t *testing.T) {
	x := Symbol("x")
	trig := Add(Pow(Sin(x), Number(2)), Pow(Cos(x), Number(2)))
	if trig.Simplify().String() != "1" {
		t.Error("sin² + cos² != 1")
	}
}

func TestParse(t *testing.T) {
	p := Parse("x^2 + 3*x")
	if p.String() != "x^2 + 3*x" {
		t.Errorf("Parse got %s", p.String())
	}
}
