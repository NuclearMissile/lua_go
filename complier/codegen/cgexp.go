package codegen

import "lua_go/complier/ast"

const (
	ArgConst = 1 // const index
	ArgReg   = 2 // register index
	ArgUpval = 4 // upvalue index
	ArgRK    = ArgReg | ArgConst
	ArgRU    = ArgReg | ArgUpval
)

func expToOpArg(fi *funcInfo, node ast.Exp, argKinds int) (arg, argKind int) {
	if argKinds&ArgConst > 0 {
		idx := -1
		switch x := node.(type) {
		case *ast.NilExp:
			idx = fi.indexOfConstant(nil)
		case *ast.FalseExp:
			idx = fi.indexOfConstant(false)
		case *ast.TrueExp:
			idx = fi.indexOfConstant(true)
		case *ast.IntegerExp:
			idx = fi.indexOfConstant(x.Val)
		case *ast.FloatExp:
			idx = fi.indexOfConstant(x.Val)
		case *ast.StringExp:
			idx = fi.indexOfConstant(x.Str)
		}
		if idx >= 0 && idx <= 0xFF {
			return 0x100 + idx, ArgConst
		}
	}
	if nameExp, ok := node.(*ast.NameExp); ok {
		if argKinds&ArgReg > 0 {
			if r := fi.slotOfLocalVar(nameExp.Name); r >= 0 {
				return r, ArgReg
			}
		}
		if argKinds&ArgUpval > 0 {
			if idx := fi.indexOfUpval(nameExp.Name); idx >= 0 {
				return idx, ArgUpval
			}
		}
	}
	a := fi.allocReg()
	cgExp(fi, node, a, 1)
	return a, ArgReg
}
