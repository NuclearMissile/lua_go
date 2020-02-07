package emitter

import (
	. "compiler/ast"
	"compiler/lexer"
	"vm"
)

// type of operands
const (
	ArgConst = 1 // const index
	ArgReg   = 2 // register index
	ArgUpval = 4 // upvalue index
	ArgRK    = ArgReg | ArgConst
	ArgRU    = ArgReg | ArgUpval
)

func (fi *funcInfo) evalExp(node Exp, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(exp.Line, a, n)
	case *FalseExp:
		fi.emitLoadBool(exp.Line, a, 0, 0)
	case *TrueExp:
		fi.emitLoadBool(exp.Line, a, 1, 0)
	case *IntegerExp:
		fi.emitLoadK(exp.Line, a, fi.indexOfConstant(exp.Val))
	case *FloatExp:
		fi.emitLoadK(exp.Line, a, fi.indexOfConstant(exp.Val))
	case *StringExp:
		fi.emitLoadK(exp.Line, a, fi.indexOfConstant(exp.Str))
	case *ParensExp:
		fi.evalExp(exp.Exp, a, 1)
	case *VarargExp:
		fi.evalVarargExp(exp, a, n)
	case *FuncDefExp:
		fi.evalFuncDefExp(exp, a)
	case *TableCtorExp:
		fi.evalTableCtorExp(exp, a)
	case *UnopExp:
		fi.evalUnopExp(exp, a)
	case *BinopExp:
		fi.evalBinopExp(exp, a)
	case *ConcatExp:
		fi.evalConcatExp(exp, a)
	case *NameExp:
		fi.evalNameExp(exp, a)
	case *TableAccessExp:
		fi.evalTableAccessExp(exp, a)
	case *FuncCallExp:
		fi.evalFuncCallExp(exp, a, n)
	}
}

func (fi *funcInfo) evalFuncCallExp(exp *FuncCallExp, a int, n int) {
	nArgs := fi.prepFuncCall(exp, a)
	fi.emitCall(exp.Line, a, nArgs, n)
}

func (fi *funcInfo) evalTableAccessExp(exp *TableAccessExp, a int) {
	regsBefore := fi.usedRegs
	b, kindB := fi.expToOpArg(exp.PrefixExp, ArgRU)
	c, _ := fi.expToOpArg(exp.KeyExp, ArgRK)
	fi.usedRegs = regsBefore

	if kindB == ArgUpval {
		fi.emitGetTabUp(exp.LastLine, a, b, c)
	} else {
		fi.emitGetTable(exp.LastLine, a, b, c)
	}
}

func (fi *funcInfo) evalNameExp(exp *NameExp, a int) {
	if r := fi.slotOfLocalVar(exp.Name); r >= 0 {
		fi.emitMove(exp.Line, a, r)
	} else if idx := fi.indexOfUpval(exp.Name); idx >= 0 {
		fi.emitGetUpval(exp.Line, a, idx)
	} else {
		fi.evalTableAccessExp(&TableAccessExp{
			LastLine: exp.Line,
			PrefixExp: &NameExp{
				Line: exp.Line,
				Name: "_ENV",
			},
			KeyExp: &StringExp{
				Line: exp.Line,
				Str:  exp.Name,
			},
		}, a)
	}
}

func (fi *funcInfo) evalConcatExp(exp *ConcatExp, a int) {
	for _, subExp := range exp.Exps {
		a := fi.allocReg()
		fi.evalExp(subExp, a, 1)
	}
	c := fi.usedRegs - 1
	b := c - len(exp.Exps) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(exp.Line, vm.OP_CONCAT, a, b, c)
}

func (fi *funcInfo) evalUnopExp(exp *UnopExp, a int) {
	regsBefore := fi.usedRegs
	b, _ := fi.expToOpArg(exp.Exp, ArgReg)
	fi.emitUnaryOp(exp.Line, exp.Op, a, b)
	fi.usedRegs = regsBefore
}

func (fi *funcInfo) evalBinopExp(exp *BinopExp, a int) {
	switch exp.Op {
	case lexer.TOKEN_OP_AND, lexer.TOKEN_OP_OR:
		regsBefore := fi.usedRegs
		b, _ := fi.expToOpArg(exp.Exp1, ArgReg)
		fi.usedRegs = regsBefore
		if exp.Op == lexer.TOKEN_OP_AND {
			fi.emitTestSet(exp.Line, a, b, 0)
		} else {
			fi.emitTestSet(exp.Line, a, b, 1)
		}
		pcJmp := fi.emitJmp(exp.Line, 0, 0)
		fi.usedRegs = regsBefore
		fi.emitMove(exp.Line, a, b)
		fi.setSBx(pcJmp, fi.pc()-pcJmp)
	default:
		regsBefore := fi.usedRegs
		b, _ := fi.expToOpArg(exp.Exp1, ArgRK)
		c, _ := fi.expToOpArg(exp.Exp2, ArgRK)
		fi.emitBinaryOp(exp.Line, exp.Op, a, b, c)
		fi.usedRegs = regsBefore
	}
}

func (fi *funcInfo) evalTableCtorExp(exp *TableCtorExp, a int) {
	nArr := 0
	for _, keyExp := range exp.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}
	nExps := len(exp.KeyExps)
	multRet := nExps > 0 && isVarargOrFuncCall(exp.ValExps[nExps-1])
	fi.emitNewTable(exp.Line, a, nArr, nExps-nArr)
	arrIdx := 0
	for i, keyExp := range exp.KeyExps {
		valExp := exp.ValExps[i]
		if keyExp == nil {
			arrIdx++
			tmp := fi.allocReg()
			if i == nExps-1 && multRet {
				fi.evalExp(valExp, tmp, -1)
			} else {
				fi.evalExp(valExp, tmp, 1)
			}
			if arrIdx%50 == 0 || arrIdx == nArr {
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				fi.freeRegs(n)
				line := lastLineOf(valExp)
				c := (arrIdx-1)/50 + 1
				if i == nExps-1 && multRet {
					fi.emitSetList(line, a, 0, c)
				} else {
					fi.emitSetList(line, a, n, c)
				}
			}
		}

		b := fi.allocReg()
		fi.evalExp(keyExp, b, 1)
		c := fi.allocReg()
		fi.evalExp(valExp, c, 1)
		fi.freeRegs(2)

		line := lastLineOf(valExp)
		fi.emitSetTable(line, a, b, c)
	}
}

func (fi *funcInfo) evalFuncDefExp(exp *FuncDefExp, a int) {
	subFunc := newFuncInfo(fi, exp)
	fi.subFuncs = append(fi.subFuncs, subFunc)

	for _, param := range exp.Params {
		subFunc.addLocVar(param, 0)
	}

	subFunc.evalBlock(exp.Block)
	subFunc.exitScope(subFunc.pc() + 2)
	subFunc.emitReturn(exp.LastLine, 0, 0)
	fi.emitClosure(exp.LastLine, a, len(fi.subFuncs)-1)
}

func (fi *funcInfo) evalVarargExp(exp *VarargExp, a int, n int) {
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	fi.emitVararg(exp.Line, a, n)
}

func (fi *funcInfo) evalTailCallExp(exp *FuncCallExp, a int) {
	nArgs := fi.prepFuncCall(exp, a)
	fi.emitTailCall(exp.Line, a, nArgs)
}

func (fi *funcInfo) prepFuncCall(exp *FuncCallExp, a int) int {
	nArgs := len(exp.Args)
	lastArgIsVarargOrFuncCall := false

	fi.evalExp(exp.PrefixExp, a, 1)
	if exp.NameExp != nil {
		fi.allocReg()
		c, k := fi.expToOpArg(exp.NameExp, ArgRK)
		fi.emitSelf(exp.Line, a, a, c)
		if k == ArgReg {
			fi.freeReg()
		}
	}

	for i, arg := range exp.Args {
		tmp := fi.allocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			fi.evalExp(arg, tmp, -1)
		} else {
			fi.evalExp(arg, tmp, 1)
		}
	}
	fi.freeRegs(nArgs)
	if exp.NameExp != nil {
		fi.freeReg()
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}
	return nArgs
}
