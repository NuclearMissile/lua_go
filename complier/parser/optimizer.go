package parser

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
	"lua_go/number"
)

func optimizeUnaryOp(exp *UnopExp) Exp {
	switch exp.Op {
	case TOKEN_OP_UNM:
		switch x := exp.Exp.(type) {
		case *IntegerExp:
			x.Val = -x.Val
			return x
		case *FloatExp:
			x.Val = -x.Val
			return x
		}
	case TOKEN_OP_NOT:
		switch exp.Exp.(type) {
		case *NilExp, *FalseExp:
			return &TrueExp{Line: exp.Line}
		case *TrueExp, *IntegerExp, *StringExp, *FloatExp:
			return &FalseExp{Line: exp.Line}
		}
	case TOKEN_OP_BNOT:
		switch x := exp.Exp.(type) {
		case *IntegerExp:
			x.Val = ^x.Val
			return x
		case *FloatExp:
			if i, ok := number.FloatToInteger(x.Val); ok {
				return &IntegerExp{
					Line: exp.Line,
					Val:  ^i,
				}
			}
		}
	}
	return exp
}
