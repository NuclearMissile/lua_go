package emitter

import (
	. "compiler/lexer"
	. "vm"
)

type instBuf struct {
	insts    []uint32
	lineNums []uint32
}

func (ib *instBuf) pc() int {
	return len(ib.insts) - 1
}

func (ib *instBuf) fixSbx(pc, sBx int) {
	if sBx > 0 && sBx > MAXARG_sBx || sBx < 0 && -sBx > MAXARG_sBx {
		panic("control structure too long")
	}

	i := ib.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+MAXARG_sBx)<<14 // reset sBx
	ib.insts[pc] = i
}

func (ib *instBuf) emitABC(line, opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	ib.insts = append(ib.insts, uint32(i))
	ib.lineNums = append(ib.lineNums, uint32(line))
}

func (ib *instBuf) emitABx(line, opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	ib.insts = append(ib.insts, uint32(i))
	ib.lineNums = append(ib.lineNums, uint32(line))
}

func (ib *instBuf) emitAsBx(line, opcode, a, sBx int) {
	i := (sBx+MAXARG_sBx)<<14 | a<<6 | opcode
	ib.insts = append(ib.insts, uint32(i))
	ib.lineNums = append(ib.lineNums, uint32(line))
}

func (ib *instBuf) emitAx(line, opcode, ax int) {
	i := ax<<6 | opcode
	ib.insts = append(ib.insts, uint32(i))
	ib.lineNums = append(ib.lineNums, uint32(line))
}

// r[a] = r[b]
func (ib *instBuf) emitMove(line, a, b int) {
	ib.emitABC(line, OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (ib *instBuf) emitLoadNil(line, a, n int) {
	ib.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (ib *instBuf) emitLoadBool(line, a, b, c int) {
	ib.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (ib *instBuf) emitLoadK(line, a int, idx int) {
	if idx < (1 << 18) {
		ib.emitABx(line, OP_LOADK, a, idx)
	} else {
		ib.emitABx(line, OP_LOADKX, a, 0)
		ib.emitAx(line, OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (ib *instBuf) emitVararg(line, a, n int) {
	ib.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (ib *instBuf) emitClosure(line, a, bx int) {
	ib.emitABx(line, OP_CLOSURE, a, bx)
}

// r[a] = {}
func (ib *instBuf) emitNewTable(line, a, nArr, nRec int) {
	ib.emitABC(line, OP_NEWTABLE,
		a, Int2FloatByte(nArr), Int2FloatByte(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (ib *instBuf) emitSetList(line, a, b, c int) {
	ib.emitABC(line, OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (ib *instBuf) emitGetTable(line, a, b, c int) {
	ib.emitABC(line, OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (ib *instBuf) emitSetTable(line, a, b, c int) {
	ib.emitABC(line, OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (ib *instBuf) emitGetUpval(line, a, b int) {
	ib.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (ib *instBuf) emitSetUpval(line, a, b int) {
	ib.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (ib *instBuf) emitGetTabUp(line, a, b, c int) {
	ib.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (ib *instBuf) emitSetTabUp(line, a, b, c int) {
	ib.emitABC(line, OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (ib *instBuf) emitCall(line, a, nArgs, nRet int) {
	ib.emitABC(line, OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (ib *instBuf) emitTailCall(line, a, nArgs int) {
	ib.emitABC(line, OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (ib *instBuf) emitReturn(line, a, n int) {
	ib.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (ib *instBuf) emitSelf(line, a, b, c int) {
	ib.emitABC(line, OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (ib *instBuf) emitJmp(line, a, sBx int) int {
	ib.emitAsBx(line, OP_JMP, a, sBx)
	return len(ib.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (ib *instBuf) emitTest(line, a, c int) {
	ib.emitABC(line, OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (ib *instBuf) emitTestSet(line, a, b, c int) {
	ib.emitABC(line, OP_TESTSET, a, b, c)
}

func (ib *instBuf) emitForPrep(line, a, sBx int) int {
	ib.emitAsBx(line, OP_FORPREP, a, sBx)
	return len(ib.insts) - 1
}

func (ib *instBuf) emitForLoop(line, a, sBx int) int {
	ib.emitAsBx(line, OP_FORLOOP, a, sBx)
	return len(ib.insts) - 1
}

func (ib *instBuf) emitTForCall(line, a, c int) {
	ib.emitABC(line, OP_TFORCALL, a, 0, c)
}

func (ib *instBuf) emitTForLoop(line, a, sBx int) {
	ib.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (ib *instBuf) emitUnaryOp(line, op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		ib.emitABC(line, OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		ib.emitABC(line, OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		ib.emitABC(line, OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		ib.emitABC(line, OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (ib *instBuf) emitBinaryOp(line, op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		ib.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case TOKEN_OP_EQ:
			ib.emitABC(line, OP_EQ, 1, b, c)
		case TOKEN_OP_NE:
			ib.emitABC(line, OP_EQ, 0, b, c)
		case TOKEN_OP_LT:
			ib.emitABC(line, OP_LT, 1, b, c)
		case TOKEN_OP_GT:
			ib.emitABC(line, OP_LT, 1, c, b)
		case TOKEN_OP_LE:
			ib.emitABC(line, OP_LE, 1, b, c)
		case TOKEN_OP_GE:
			ib.emitABC(line, OP_LE, 1, c, b)
		}
		ib.emitJmp(line, 0, 1)
		ib.emitLoadBool(line, a, 0, 1)
		ib.emitLoadBool(line, a, 1, 0)
	}
}
