package sympy

import "fmt"

// Expr is the interface for all symbolic expressions.
type Expr interface {
    String() string
}

// constExpr represents a constant value.
type constExpr struct {
    value float64
}

func (c *constExpr) String() string {
    return fmt.Sprintf("%v", c.value)
}

// varExpr represents a symbolic variable.
type varExpr struct {
    name string
}

func (v *varExpr) String() string {
    return v.name
}

// addExpr represents an addition operation.
type addExpr struct {
    left, right Expr
}

func (a *addExpr) String() string {
    return fmt.Sprintf("( %s + %s )", a.left.String(), a.right.String())
}

// mulExpr represents a multiplication operation.
type mulExpr struct {
    left, right Expr
}

func (m *mulExpr) String() string {
    return fmt.Sprintf("( %s * %s )", m.left.String(), m.right.String())
}

// Number creates a constant expression.
func Number(value float64) Expr {
    return &constExpr{value}
}

// Symbol creates a variable expression.
func Symbol(name string) Expr {
    return &varExpr{name}
}

// Add creates an addition expression.
func Add(left, right Expr) Expr {
    return &addExpr{left, right}
}

// Mul creates a multiplication expression.
func Mul(left, right Expr) Expr {
    return &mulExpr{left, right}
}

// Diff computes the derivative of the expression with respect to the given symbol.
func Diff(e Expr, sym Expr) Expr {
    v, ok := sym.(*varExpr)
    if !ok {
        return nil // Invalid symbol
    }

    switch expr := e.(type) {
    case *constExpr:
        return Number(0)
    case *varExpr:
        if expr.name == v.name {
            return Number(1)
        }
        return Number(0)
    case *addExpr:
        return Add(Diff(expr.left, sym), Diff(expr.right, sym))
    case *mulExpr:
        // Product rule: u'v + uv'
        return Add(Mul(Diff(expr.left, sym), expr.right), Mul(expr.left, Diff(expr.right, sym)))
    default:
        return nil // Unsupported type
    }
}
