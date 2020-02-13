package emitter

import . "compiler/ast"

func evalBlock(fi *funcInfo, node *Block) {
	for _, stat := range node.Stats {
		evalStat(fi, stat)
	}

	if node.RetExps != nil {
		evalRetStat(fi, node.RetExps, node.LastLine)
	}
}

func evalRetStat(fi *funcInfo, exps []Exp, lastLine int) {
	nExps := len(exps)
	if nExps == 0 {
		fi.emitReturn(lastLine, 0, 0)
		return
	}

	if nExps == 1 {
		if nameExp, ok := exps[0].(*NameExp); ok {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				fi.emitReturn(lastLine, r, 1)
				return
			}
		}
		if fcExp, ok := exps[0].(*FuncCallExp); ok {
			r := fi.allocReg()
			cgTailCallExp(fi, fcExp, r)
			fi.freeReg()
			fi.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(exps[nExps-1])
	for i, exp := range exps {
		r := fi.allocReg()
		if i == nExps-1 && multRet {
			evalExp(fi, exp, r, -1)
		} else {
			evalExp(fi, exp, r, 1)
		}
	}
	fi.freeRegs(nExps)

	a := fi.usedRegs // correct?
	if multRet {
		fi.emitReturn(lastLine, a, -1)
	} else {
		fi.emitReturn(lastLine, a, nExps)
	}
}

func evalStat(fi *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		evalFuncCallStat(fi, stat)
	case *BreakStat:
		evalBreakStat(fi, stat)
	case *DoStat:
		evalDoStat(fi, stat)
	case *WhileStat:
		evalWhileStat(fi, stat)
	case *RepeatStat:
		evalRepeatStat(fi, stat)
	case *IfStat:
		evalIfStat(fi, stat)
	case *ForNumStat:
		evalForNumStat(fi, stat)
	case *ForInStat:
		evalForInStat(fi, stat)
	case *AssignStat:
		evalAssignStat(fi, stat)
	case *LocalVarDeclStat:
		evalLocalVarDeclStat(fi, stat)
	case *LocalFuncDefStat:
		evalLocalFuncDefStat(fi, stat)
	case *LabelStat:
		evalLabelStat(fi, stat)
	case *GotoStat:
		evalGotoStat(fi, stat)
	}
}

func evalLabelStat(fi *funcInfo, node *LabelStat) {
	fi.addLabel(node.Name, node.Line)
}

func evalGotoStat(fi *funcInfo, node *GotoStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addGoto(pc, fi.scopeLv, node.Name)
}

func evalLocalFuncDefStat(fi *funcInfo, node *LocalFuncDefStat) {
	r := fi.addLocVar(node.Name, fi.pc()+2)
	evalFuncDefExp(fi, node.Exp, r)
}

func evalFuncCallStat(fi *funcInfo, node *FuncCallStat) {
	r := fi.allocReg()
	evalFuncCallExp(fi, node, r, 0)
	fi.freeReg()
}

func evalBreakStat(fi *funcInfo, node *BreakStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addBreakJmp(pc)
}

func evalDoStat(fi *funcInfo, node *DoStat) {
	fi.enterScope(false)
	evalBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.exitScope(fi.pc() + 1)
}

/*
           ______________
          /  false? jmp  |
         /               |
while exp do block end <-'
      ^           \
      |___________/
           jmp
*/
func evalWhileStat(fi *funcInfo, node *WhileStat) {
	pcBeforeExp := fi.pc()

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ArgReg)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	pcJmpToEnd := fi.emitJmp(line, 0, 0)

	fi.enterScope(true)
	evalBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.emitJmp(node.Block.LastLine, 0, pcBeforeExp-fi.pc()-1)
	fi.exitScope(fi.pc())

	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

/*
        ______________
       |  false? jmp  |
       V              /
repeat block until exp
*/
func evalRepeatStat(fi *funcInfo, node *RepeatStat) {
	fi.enterScope(true)

	pcBeforeBlock := fi.pc()
	evalBlock(fi, node.Block)

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ArgReg)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	fi.emitJmp(line, fi.getJmpArgA(), pcBeforeBlock-fi.pc()-1)
	fi.closeOpenUpvals(line)

	fi.exitScope(fi.pc() + 1)
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func evalIfStat(fi *funcInfo, node *IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
		}

		oldRegs := fi.usedRegs
		a, _ := expToOpArg(fi, exp, ArgReg)
		fi.usedRegs = oldRegs

		line := lastLineOf(exp)
		fi.emitTest(line, a, 0)
		pcJmpToNextExp = fi.emitJmp(line, 0, 0)

		block := node.Blocks[i]
		fi.enterScope(false)
		evalBlock(fi, block)
		fi.closeOpenUpvals(block.LastLine)
		fi.exitScope(fi.pc() + 1)
		if i < len(node.Exps)-1 {
			pcJmpToEnds[i] = fi.emitJmp(block.LastLine, 0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

func evalForNumStat(fi *funcInfo, node *ForNumStat) {
	forIndexVar := "(for index)"
	forLimitVar := "(for limit)"
	forStepVar := "(for step)"

	fi.enterScope(true)

	evalLocalVarDeclStat(fi, &LocalVarDeclStat{
		Names: []string{forIndexVar, forLimitVar, forStepVar},
		Exps:  []Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName, fi.pc()+2)

	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(node.LineOfDo, a, 0)
	evalBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	pcForLoop := fi.emitForLoop(node.LineOfFor, a, 0)

	fi.fixSbx(pcForPrep, pcForLoop-pcForPrep-1)
	fi.fixSbx(pcForLoop, pcForPrep-pcForLoop)

	fi.exitScope(fi.pc())
	fi.fixEndPC(forIndexVar, 1)
	fi.fixEndPC(forLimitVar, 1)
	fi.fixEndPC(forStepVar, 1)
}

func evalForInStat(fi *funcInfo, node *ForInStat) {
	forGeneratorVar := "(for generator)"
	forStateVar := "(for state)"
	forControlVar := "(for control)"

	fi.enterScope(true)

	evalLocalVarDeclStat(fi, &LocalVarDeclStat{
		//LastLine: 0,
		Names: []string{forGeneratorVar, forStateVar, forControlVar},
		Exps:  node.Exps,
	})
	for _, name := range node.Names {
		fi.addLocVar(name, fi.pc()+2)
	}

	pcJmpToTFC := fi.emitJmp(node.LineOfDo, 0, 0)
	evalBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)

	line := lineOf(node.Exps[0])
	rGenerator := fi.slotOfLocVar(forGeneratorVar)
	fi.emitTForCall(line, rGenerator, len(node.Names))
	fi.emitTForLoop(line, rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope(fi.pc() - 1)
	fi.fixEndPC(forGeneratorVar, 2)
	fi.fixEndPC(forStateVar, 2)
	fi.fixEndPC(forControlVar, 2)
}

func evalLocalVarDeclStat(fi *funcInfo, node *LocalVarDeclStat) {
	exps := removeTailNils(node.Exps)
	nExps := len(exps)
	nNames := len(node.Names)

	oldRegs := fi.usedRegs
	if nExps == nNames {
		for _, exp := range exps {
			a := fi.allocReg()
			evalExp(fi, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				evalExp(fi, exp, a, 0)
			} else {
				evalExp(fi, exp, a, 1)
			}
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				evalExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				evalExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	fi.usedRegs = oldRegs
	startPC := fi.pc() + 1
	for _, name := range node.Names {
		fi.addLocVar(name, startPC)
	}
}

func evalAssignStat(fi *funcInfo, node *AssignStat) {
	exps := removeTailNils(node.Exps)
	nExps := len(exps)
	nVars := len(node.Vars)

	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)
	oldRegs := fi.usedRegs

	for i, exp := range node.Vars {
		if taExp, ok := exp.(*TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			evalExp(fi, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			evalExp(fi, taExp.KeyExp, kRegs[i], 1)
		} else {
			name := exp.(*NameExp).Name
			if fi.slotOfLocVar(name) < 0 && fi.indexOfUpval(name) < 0 {
				// global var
				kRegs[i] = -1
				if fi.indexOfConstant(name) > 0xFF {
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
				evalExp(fi, exp, a, 0)
			} else {
				evalExp(fi, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				evalExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				evalExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	lastLine := node.LastLine
	for i, exp := range node.Vars {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(lastLine, a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(lastLine, vRegs[i], b)
			} else if a := fi.slotOfLocVar("_ENV"); a >= 0 {
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTable(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTable(lastLine, a, kRegs[i], vRegs[i])
				}
			} else { // global var
				a := fi.indexOfUpval("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTabUp(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTabUp(lastLine, a, kRegs[i], vRegs[i])
				}
			}
		} else {
			fi.emitSetTable(lastLine, tRegs[i], kRegs[i], vRegs[i])
		}
	}

	// todo
	fi.usedRegs = oldRegs
}
