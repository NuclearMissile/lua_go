package emitter

import . "compiler/ast"

func emitStat(fi *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		emitFuncCallStat(fi, stat)
	case *BreakStat:
		emitBreakStat(fi, stat)
	case *DoStat:
		emitDoStat(fi, stat)
	case *WhileStat:
		emitWhileStat(fi, stat)
	case *RepeatStat:
		emitRepeatStat(fi, stat)
	case *IfStat:
		emitIfStat(fi, stat)
	case *ForNumStat:
		emitForNumStat(fi, stat)
	case *ForInStat:
		emitForInStat(fi, stat)
	case *AssignStat:
		emitAssignStat(fi, stat)
	case *LocalVarDeclStat:
		emitLocalVarDeclStat(fi, stat)
	case *LocalFuncDefStat:
		emitLocalFuncDefStat(fi, stat)
	case *LabelStat:
		emitLabelStat(fi, stat)
	case *GotoStat:
		emitGotoStat(fi, stat)
	}
}

func emitGotoStat(fi *funcInfo, stat *GotoStat) {

}

func emitLabelStat(fi *funcInfo, stat *LabelStat) {

}

func emitLocalFuncDefStat(fi *funcInfo, stat *LocalFuncDefStat) {

}

func emitLocalVarDeclStat(fi *funcInfo, stat *LocalVarDeclStat) {

}

func emitAssignStat(fi *funcInfo, stat *AssignStat) {

}

func emitForInStat(fi *funcInfo, stat *ForInStat) {

}

func emitForNumStat(fi *funcInfo, stat *ForNumStat) {

}

func emitIfStat(fi *funcInfo, stat *IfStat) {

}

func emitRepeatStat(fi *funcInfo, stat *RepeatStat) {

}

func emitWhileStat(fi *funcInfo, stat *WhileStat) {

}

func emitDoStat(fi *funcInfo, stat *DoStat) {

}

func emitBreakStat(fi *funcInfo, stat *BreakStat) {

}

func emitFuncCallStat(fi *funcInfo, stat *FuncCallStat) {

}
