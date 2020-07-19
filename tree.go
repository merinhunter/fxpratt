package fxpratt

import (
	"fmt"
	"fxlex"
	"math"
	"os"
)

type Expr struct {
	tok    fxlex.Token
	ERight *Expr
	ELeft  *Expr
}

func NewExpr(tok fxlex.Token) (expr *Expr) {
	return &Expr{tok: tok}
}

func (e *Expr) String() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("\t%p EXPR[%s](%d) L->%p R->%p", e, e.tok.GetType(), e.tok.GetValue(), e.ELeft, e.ERight)
}

const DebugExpr = false

func (e *Expr) Eval() float64 {
	if DebugExpr {
		fmt.Fprintf(os.Stderr, "%s\n", e)
	}
	rV := 0.0
	lV := 0.0
	if e == nil {
		return 0
	}
	if e.ERight != nil {
		rV = e.ERight.Eval()
	}
	if e.ELeft != nil {
		lV = e.ELeft.Eval()
	}
	tok := e.tok
	switch tok.GetTokType() {
	case fxlex.TokMinus:
		return lV - rV
	case fxlex.TokPlus:
		return lV + rV
	case fxlex.TokTimes:
		return lV * rV
	case fxlex.TokDivide:
		return lV / rV
	case fxlex.TokRem:
		return math.Mod(lV, rV)
	case fxlex.TokPow:
		return math.Pow(lV, rV)
	case fxlex.TokGT:
		if lV > rV {
			return 1.0
		}
		return 0.0
	case fxlex.TokLT:
		if lV < rV {
			return 1.0
		}
		return 0.0
	case fxlex.TokGTE:
		if lV >= rV {
			return 1.0
		}
		return 0.0
	case fxlex.TokLTE:
		if lV <= rV {
			return 1.0
		}
		return 0.0
	case fxlex.TokOr:
		if (lV != 0) || (rV != 0) {
			return 1.0
		}
		return 0.0
	case fxlex.TokAnd:
		if (lV != 0) && (rV != 0) {
			return 1.0
		}
		return 0.0
	case fxlex.TokNeg:
		if !(rV != 0) {
			return 1.0
		}
		return 0.0
	case fxlex.TokXor:
		if (lV != 0) != (rV != 0) {
			return 1.0
		}
		return 0.0
	case fxlex.TokIntLit:
		return float64(tok.GetValue())
	case fxlex.TokBoolLit:
		return float64(tok.GetValue())
	default:
		panic("Bad subtree")
	}
}
