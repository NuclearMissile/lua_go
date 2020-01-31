package emitter

import . "compiler/ast"

func emitExp(fi *funcInfo, node Exp, a, n int) {
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
		emitExp(fi, exp.Exp, a, 1)
	case *VarargExp:
		emitVarargExp(fi, exp, a, n)
	case *FuncDefExp:
		emitFuncDefExp(fi, exp, a)
	case *TableCtorExp:
		emitTableConstructorExp(fi, exp, a)
	case *UnopExp:
		emitUnopExp(fi, exp, a)
	case *BinopExp:
		emitBinopExp(fi, exp, a)
	case *ConcatExp:
		emitConcatExp(fi, exp, a)
	case *NameExp:
		emitNameExp(fi, exp, a)
	case *TableAccessExp:
		emitTableAccessExp(fi, exp, a)
	case *FuncCallExp:
		emitFuncCallExp(fi, exp, a, n)
	}
}

func emitFuncCallExp(fi *funcInfo, exp *FuncCallExp, a int, n int) {

}

func emitTableAccessExp(fi *funcInfo, exp *TableAccessExp, a int) {

}

func emitNameExp(fi *funcInfo, exp *NameExp, a int) {

}

func emitConcatExp(fi *funcInfo, exp *ConcatExp, a int) {

}

func emitUnopExp(fi *funcInfo, exp *UnopExp, a int) {

}

func emitBinopExp(fi *funcInfo, exp *BinopExp, a int) {

}

func emitTableConstructorExp(fi *funcInfo, exp *TableCtorExp, a int) {

}

func emitFuncDefExp(fi *funcInfo, exp *FuncDefExp, a int) {

}

func emitVarargExp(fi *funcInfo, exp *VarargExp, a int, n int) {

}

func emitTailCallExp(fi *funcInfo, exp *FuncCallExp, a int) {

}
