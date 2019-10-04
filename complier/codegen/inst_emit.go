package codegen

import (
	. "lua_go/complier/lexer"
	. "lua_go/vm"
)

func (self *funcInfo) emitABC(line, opcode, a, b, c int) {
	self.insts = append(self.insts, uint32(b<<23|c<<14|a<<6|opcode))
	self.lineNums = append(self.lineNums, uint32(line))
}

func (self *funcInfo) emitABx(line, opcode, a, bx int) {
	self.insts = append(self.insts, uint32(bx<<14|a<<6|opcode))
	self.lineNums = append(self.lineNums, uint32(line))
}

func (self *funcInfo) emitAsBx(line, opcode, a, b int) {
	self.insts = append(self.insts, uint32((b+vm.MAXARG_sBx)<<14|a<<6|opcode))
	self.lineNums = append(self.lineNums, uint32(line))
}

func (self *funcInfo) emitAx(line, opcode, ax int) {
	self.insts = append(self.insts, uint32(ax<<6|opcode))
	self.lineNums = append(self.lineNums, uint32(line))
}

// r[a] = r[b]
func (self *funcInfo) emitMove(line, a, b int) {
	self.emitABC(line, OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (self *funcInfo) emitLoadNil(line, a, n int) {
	self.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (self *funcInfo) emitLoadBool(line, a, b, c int) {
	self.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (self *funcInfo) emitLoadK(line, a int, idx int) {
	if idx < (1 << 18) {
		self.emitABx(line, OP_LOADK, a, idx)
	} else {
		self.emitABx(line, OP_LOADKX, a, 0)
		self.emitAx(line, OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (self *funcInfo) emitVararg(line, a, n int) {
	self.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (self *funcInfo) emitClosure(line, a, bx int) {
	self.emitABx(line, OP_CLOSURE, a, bx)
}

// r[a] = {}
func (self *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	self.emitABC(line, OP_NEWTABLE, a, Int2fb(nArr), Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (self *funcInfo) emitSetList(line, a, b, c int) {
	self.emitABC(line, OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (self *funcInfo) emitGetTable(line, a, b, c int) {
	self.emitABC(line, OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTable(line, a, b, c int) {
	self.emitABC(line, OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (self *funcInfo) emitGetUpval(line, a, b int) {
	self.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (self *funcInfo) emitSetUpval(line, a, b int) {
	self.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (self *funcInfo) emitGetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (self *funcInfo) emitCall(line, a, nArgs, nRet int) {
	self.emitABC(line, OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (self *funcInfo) emitTailCall(line, a, nArgs int) {
	self.emitABC(line, OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (self *funcInfo) emitReturn(line, a, n int) {
	self.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (self *funcInfo) emitSelf(line, a, b, c int) {
	self.emitABC(line, OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (self *funcInfo) emitJmp(line, a, sBx int) int {
	self.emitAsBx(line, OP_JMP, a, sBx)
	return len(self.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (self *funcInfo) emitTest(line, a, c int) {
	self.emitABC(line, OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (self *funcInfo) emitTestSet(line, a, b, c int) {
	self.emitABC(line, OP_TESTSET, a, b, c)
}

func (self *funcInfo) emitForPrep(line, a, sBx int) int {
	self.emitAsBx(line, OP_FORPREP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitForLoop(line, a, sBx int) int {
	self.emitAsBx(line, OP_FORLOOP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitTForCall(line, a, c int) {
	self.emitABC(line, OP_TFORCALL, a, 0, c)
}

func (self *funcInfo) emitTForLoop(line, a, sBx int) {
	self.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (self *funcInfo) emitUnaryOp(line, op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		self.emitABC(line, OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		self.emitABC(line, OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		self.emitABC(line, OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		self.emitABC(line, OP_UNM, a, b, 0)
	}
}
