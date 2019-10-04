package codegen

import (
	. "lua_go/complier/ast"
)

// MUX
func cgStat(fi *funcInfo, stat Stat) {
	switch stat := stat.(type) {
	case *FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *BreakStat:
		cgBreakStat(fi, stat)
	case *DoStat:
		cgDoStat(fi, stat)
	case *RepeatStat:
		cgRepeatStat(fi, stat)
	case *WhileStat:
		cgWhileStat(fi, stat)
	case *IfStat:
		cgIfStat(fi, stat)
	case *ForNumStat:
		cgForNumStat(fi, stat)
	case *ForInStat:
		cgForInStat(fi, stat)
	case *AssignStat:
		cgAssignStat(fi, stat)
	case *LocalVarDeclStat:
		cgLocalVarStat(fi, stat)
	case *LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *LabelStat:
		cgLabelStat(fi, stat)
	case *GotoStat:
		cgGotoStat(fi, stat)
	}
}

func cgWhileStat(fi *funcInfo, node *WhileStat) {
	pcBeforeExp := fi.pc()
	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ArgReg)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	pcJmpToEnd := fi.emitJmp(line, 0, 0)
	fi.enterScope(true)
	cgBlock(fi, node.Block)
	fi.emitJmp(node.Block.LastLine, 0, pcBeforeExp-fi.pc()-1)
	fi.exitScope(fi.pc())
	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

func cgDoStat(fi *funcInfo, node *DoStat) {
	fi.enterScope(false)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.exitScope(fi.pc() + 1)
}

func cgBreakStat(fi *funcInfo, node *BreakStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addBreakJmp(pc)
}

func cgLocalFuncDefStat(fi *funcInfo, node *LocalFuncDefStat) {
	r := fi.addLocalVar(node.Name)
	cgFuncDefExp(fi, node.Exp, r)
}

func cgFuncCallStat(fi *funcInfo, node Stat) {

}
