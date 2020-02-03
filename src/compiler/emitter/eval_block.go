package emitter

import . "compiler/ast"

func (fi *funcInfo) evalBlock(block *Block) {
	for _, stat := range block.States {
		fi.evalStat(stat)
	}
	if block.RetExps != nil {
		fi.evalRetStat(block.RetExps, block.LastLine)
	}
}

func (fi *funcInfo) evalRetStat(retExps []Exp, lastLine int) {
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
			fi.evalTailCallExp(fcExp, r)
			fi.freeReg()
			fi.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(retExps[n-1])
	for i, exp := range retExps {
		r := fi.allocReg()
		if i == n-1 && multRet {
			fi.evalExp(exp, r, -1)
		} else {
			fi.evalExp(exp, r, 1)
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
