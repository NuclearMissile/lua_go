package emitter

import . "compiler/ast"

func emitBlock(fi *funcInfo, block *Block) {
	for _, stat := range block.States {
		emitStat(fi, stat)
	}
	if block.RetExps != nil {
		emitRetStat(fi, block.RetExps, block.LastLine)
	}
}

func emitRetStat(fi *funcInfo, retExps []Exp, lastLine int) {
	n := len(retExps)
	if n == 0 {
		fi.emitReturn(lastLine, 0, 0)
		return
	}
	if n == 1 {
		if nameExp, ok := retExps[0].(*NameExp); ok {
			if r := fi.slotOfLocalVar(nameExp.Name); r >= 0 {
				fi.emitReturn(lastLine, r, 1)
				return
			}
		}
		if fcExp, ok := retExps[0].(*FuncCallExp); ok {
			r := fi.allocReg()
			emitTailCallExp(fi, fcExp, r)
			fi.freeReg()
			fi.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(retExps[n-1])
	for i, exp := range retExps {
		r := fi.allocReg()
		if i == n-1 && multRet {
			emitExp(fi, exp, r, -1)
		} else {
			emitExp(fi, exp, r, 1)
		}
	}
	fi.freeRegs(n)

	a := fi.usedRegs
	if multRet {
		fi.emitReturn(lastLine, a, -1)
	} else {
		fi.emitReturn(lastLine, a, n)
	}
}
