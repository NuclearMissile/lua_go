package emitter

import . "compiler/ast"

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
		fi.evalTableConstructorExp(exp, a)
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

}

func (fi *funcInfo) evalTableAccessExp(exp *TableAccessExp, a int) {

}

func (fi *funcInfo) evalNameExp(exp *NameExp, a int) {

}

func (fi *funcInfo) evalConcatExp(exp *ConcatExp, a int) {

}

func (fi *funcInfo) evalUnopExp(exp *UnopExp, a int) {

}

func (fi *funcInfo) evalBinopExp(exp *BinopExp, a int) {

}

func (fi *funcInfo) evalTableConstructorExp(exp *TableCtorExp, a int) {

}

func (fi *funcInfo) evalFuncDefExp(exp *FuncDefExp, a int) {

}

func (fi *funcInfo) evalVarargExp(exp *VarargExp, a int, n int) {

}

func (fi *funcInfo) evalTailCallExp(exp *FuncCallExp, a int) {

}
