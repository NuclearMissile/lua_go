package emitter

import . "compiler/ast"

func (fi *funcInfo) evalStat(node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		fi.evalFuncCallStat(stat)
	case *BreakStat:
		fi.evalBreakStat(stat)
	case *DoStat:
		fi.evalDoStat(stat)
	case *WhileStat:
		fi.evalWhileStat(stat)
	case *RepeatStat:
		fi.evalRepeatStat(stat)
	case *IfStat:
		fi.evalIfStat(stat)
	case *ForNumStat:
		fi.evalForNumStat(stat)
	case *ForInStat:
		fi.evalForInStat(stat)
	case *AssignStat:
		fi.evalAssignStat(stat)
	case *LocalVarDeclStat:
		fi.evalLocalVarDeclStat(stat)
	case *LocalFuncDefStat:
		fi.evalLocalFuncDefStat(stat)
	case *LabelStat:
		fi.evalLabelStat(stat)
	case *GotoStat:
		fi.evalGotoStat(stat)
	}
}

func (fi *funcInfo) evalGotoStat(stat *GotoStat) {
	pc := fi.emitJmp(stat.Line, 0, 0)
	fi.addGoto(pc, fi.scopeLv, stat.Name)
}

func (fi *funcInfo) evalLabelStat(stat *LabelStat) {
	fi.addLabel(stat.Name, stat.Line)
}

func (fi *funcInfo) evalLocalFuncDefStat(stat *LocalFuncDefStat) {
	r := fi.addLocVar(stat.Name, fi.pc()+2)
	fi.evalFuncDefExp(stat.Exp, r)
}

func (fi *funcInfo) evalLocalVarDeclStat(stat *LocalVarDeclStat) {
	nExps, nNames := len(stat.Exps), len(stat.Names)
	exps := removeTailNils(stat.Exps)

	regsBefore := fi.usedRegs
	if nExps >= nNames {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nNames && i == nExps-1 && isVarargOrFuncCall(exp) {
				fi.evalExp(exp, a, 0)
			} else {
				fi.evalExp(exp, a, 1)
			}
		}
	} else if nExps < nNames {
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				fi.evalExp(exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				fi.evalExp(exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(stat.LastLine, a, n)
		}
	}

	fi.usedRegs = regsBefore
	startPC := fi.pc() + 1
	for _, name := range stat.Names {
		fi.addLocVar(name, startPC)
	}
}

func (fi *funcInfo) evalAssignStat(stat *AssignStat) {
	exps := removeTailNils(stat.Exps)
	nExps, nVars := len(stat.Exps), len(stat.Vars)

	tRegs, kRegs, vRegs := make([]int, nVars), make([]int, nVars), make([]int, nVars)
	regsBefore := fi.usedRegs

	for i, exp := range stat.Vars {
		if taExp, ok := exp.(*TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			fi.evalExp(taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			fi.evalExp(taExp.KeyExp, kRegs[i], 1)
		} else {
			name := exp.(*NameExp).Name
			if fi.slotOfLocalVar(name) < 0 && fi.indexOfUpval(name) < 0 {
				kRegs[i] = -1
				if fi.indexOfUpval(name) > 0xff {
					kRegs[i] = fi.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				fi.evalExp(exp, a, 0)
			} else {
				fi.evalExp(exp, a, 1)
			}
		}
	} else {
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				fi.evalExp(exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				fi.evalExp(exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(stat.LastLine, a, n)
		}
	}

	lastLine := stat.LastLine
	for i, exp := range stat.Vars {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocalVar(varName); a >= 0 {
				fi.emitMove(lastLine, a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(lastLine, vRegs[i], b)
			} else if a := fi.slotOfLocalVar("_ENV"); a >= 0 {
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTable(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTable(lastLine, a, kRegs[i], vRegs[i])
				}
			} else {
				a := fi.indexOfUpval("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfUpval(varName)
					fi.emitSetTabUp(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTabUp(lastLine, a, kRegs[i], vRegs[i])
				}
			}
		} else {
			fi.emitSetTable(lastLine, tRegs[i], kRegs[i], vRegs[i])
		}
	}
	fi.usedRegs = regsBefore
}

func (fi *funcInfo) evalForInStat(stat *ForInStat) {
	gen, state, ctrl := "(FOR_GEN)", "(FOR_STATE)", "(FOR_CTRL)"
	fi.enterScope(true)
	fi.evalLocalVarDeclStat(&LocalVarDeclStat{
		LastLine: stat.LineOfFor,
		Names:    []string{gen, state, ctrl},
		Exps:     stat.Exps,
	})

	for _, name := range stat.Names {
		fi.addLocVar(name, fi.pc()+2)
	}

	pcToTFC := fi.emitJmp(stat.LineOfDo, 0, 0)
	fi.evalBlock(stat.Block)
	fi.closeOpenUpvals(stat.Block.LastLine)
	fi.setSBx(pcToTFC, fi.pc()-pcToTFC)

	line := lineOf(stat.Exps[0])
	rGenerator := fi.slotOfLocalVar(gen)
	fi.emitTForCall(line, rGenerator, len(stat.Names))
	fi.emitTForLoop(line, rGenerator+2, pcToTFC-fi.pc()-1)

	fi.exitScope(fi.pc() - 1)
	fi.setEndPC(gen, 2)
	fi.setEndPC(state, 2)
	fi.setEndPC(ctrl, 2)
}

func (fi *funcInfo) evalForNumStat(stat *ForNumStat) {
	index, limit, step := "(FOR_INDEX)", "(FOR_LIMIT)", "(FOR_STEP)"
	fi.enterScope(true)
	fi.evalLocalVarDeclStat(&LocalVarDeclStat{
		LastLine: stat.LineOfFor,
		Names:    []string{index, limit, step},
		Exps:     []Exp{stat.InitExp, stat.LimitExp, stat.StepExp},
	})
	fi.addLocVar(stat.VarName, fi.pc()+2)

	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(stat.LineOfDo, a, 0)
	fi.evalBlock(stat.Block)
	fi.closeOpenUpvals(stat.Block.LastLine)
	pcForLoop := fi.emitForLoop(stat.LineOfFor, a, 0)

	fi.setSBx(pcForPrep, pcForLoop-pcForPrep-1)
	fi.setSBx(pcForLoop, pcForPrep-pcForLoop)

	fi.exitScope(fi.pc())
	fi.setEndPC(index, 1)
	fi.setEndPC(limit, 1)
	fi.setEndPC(step, 1)
}

func (fi *funcInfo) evalIfStat(stat *IfStat) {
	pcToEnds := make([]int, len(stat.Exps))
	pcToNextExp := -1

	for i, exp := range stat.Exps {
		if pcToNextExp >= 0 {
			fi.setSBx(pcToNextExp, fi.pc()-pcToNextExp)
		}

		regsBefore := fi.usedRegs
		a, _ := fi.expToOpArg(exp, ArgReg)
		fi.usedRegs = regsBefore

		line := lastLineOf(exp)
		fi.emitTest(line, a, 0)
		pcToNextExp = fi.emitJmp(line, 0, 0)

		currBlock := stat.Blocks[i]
		fi.enterScope(false)
		fi.evalBlock(currBlock)
		fi.closeOpenUpvals(currBlock.LastLine)
		fi.exitScope(fi.pc() + 1)

		if i < len(stat.Exps)-1 {
			pcToEnds[i] = fi.emitJmp(currBlock.LastLine, 0, 0)
		} else {
			pcToEnds[i] = pcToNextExp
		}
	}

	for _, pc := range pcToEnds {
		fi.setSBx(pc, fi.pc()-pc)
	}
}

func (fi *funcInfo) evalRepeatStat(stat *RepeatStat) {
	fi.enterScope(true)
	pcBefore := fi.pc()
	fi.evalBlock(stat.Block)
	regsBefore := fi.usedRegs
	a, _ := fi.expToOpArg(stat.Exp, ArgReg)
	fi.usedRegs = regsBefore

	line := lastLineOf(stat.Exp)
	fi.emitTest(line, a, 0)
	fi.emitJmp(line, fi.getJmpArgA(), pcBefore-fi.pc()-1)
	fi.closeOpenUpvals(line)
	fi.exitScope(fi.pc() + 1)
}

func (fi *funcInfo) evalWhileStat(stat *WhileStat) {
	pcBefore := fi.pc()
	regsBefore := fi.usedRegs
	a, _ := fi.expToOpArg(stat.Exp, ArgReg)
	fi.usedRegs = regsBefore

	line := lastLineOf(stat.Exp)
	fi.emitTest(line, a, 0)
	pcToEnd := fi.emitJmp(line, 0, 0)

	fi.enterScope(true)
	fi.evalBlock(stat.Block)
	fi.closeOpenUpvals(stat.Block.LastLine)
	fi.emitJmp(stat.Block.LastLine, 0, pcBefore-fi.pc()-1)
	fi.exitScope(fi.pc())
	fi.setSBx(pcToEnd, fi.pc()-pcToEnd)
}

func (fi *funcInfo) evalDoStat(stat *DoStat) {
	fi.enterScope(false)
	fi.evalBlock(stat.Block)
	fi.closeOpenUpvals(stat.Block.LastLine)
	fi.exitScope(fi.pc() + 1)
}

func (fi *funcInfo) evalBreakStat(stat *BreakStat) {
	pc := fi.emitJmp(stat.Line, 0, 0)
	fi.addBreakJmp(pc)
}

func (fi *funcInfo) evalFuncCallStat(stat *FuncCallStat) {
	r := fi.allocReg()
	fi.evalFuncCallExp(stat, r, 0)
	fi.freeReg()
}
