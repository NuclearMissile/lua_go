package codegen

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
	. "lua_go/vm"
)

type funcInfo struct {
	constants         map[interface{}]int
	usedRegs, maxRegs int
	scopeLV           int
	locVars           []*localVarInfo
	locNames          map[string]*localVarInfo
	breaks            [][]int
	insts             []uint32
	parent            *funcInfo
	upvalues          map[string]upvalInfo
	subFuncs          []*funcInfo
	numParams         int
	isVararg          bool
}

type upvalInfo struct {
	localVarSlot, upvalIndex, index int
}

type localVarInfo struct {
	prev     *localVarInfo
	name     string
	scopeLV  int
	slot     int
	captured bool
}

var arithAndBitwiseBinops = map[int]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		constants: map[interface{}]int{},
		locVars:   make([]*localVarInfo, 0, 8),
		locNames:  map[string]*localVarInfo{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 16),
		parent:    parent,
		upvalues:  map[string]upvalInfo{},
		subFuncs:  []*funcInfo{},
		isVararg:  fd.IsVararg,
		numParams: len(fd.ParaList),
	}
}

func (self *funcInfo) addBreakJmp(pc int) {
	for i := self.scopeLV; i >= 0; i-- {
		if self.breaks[i] != nil {
			self.breaks[i] = append(self.breaks[i], pc)
			return
		}
	}
	panic("<break> not in a loop")
}

func (self *funcInfo) enterScope(breakable bool) {
	self.scopeLV++
	if breakable {
		self.breaks = append(self.breaks, []int{})
	} else {
		self.breaks = append(self.breaks, nil)
	}
}

func (self *funcInfo) exitScope() {
	pendingBreakJmps := self.breaks[len(self.breaks)-1]
	self.breaks = self.breaks[:len(self.breaks)-1]
	a := self.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := self.pc() - pc
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP
		self.insts[pc] = uint32(i)
	}

	self.scopeLV--
	for _, locVar := range self.locVars {
		if locVar.scopeLV > self.scopeLV {
			self.removeLocalVar(locVar)
		}
	}
}

func (self *funcInfo) removeLocalVar(locVar *localVarInfo) {
	self.freeReg()
	if locVar.prev == nil {
		delete(self.locNames, locVar.name)
	} else if locVar.prev.scopeLV == locVar.scopeLV {
		self.removeLocalVar(locVar.prev)
	} else {
		self.locNames[locVar.name] = locVar.prev
	}
}

func (self *funcInfo) slotOfLocalVar(name string) int {
	if localVar, found := self.locNames[name]; found {
		return localVar.slot
	}
	return -1
}

func (self *funcInfo) addLocalVar(name string) int {
	newVar := &localVarInfo{
		prev:    self.locNames[name],
		name:    name,
		scopeLV: self.scopeLV,
		slot:    self.allocReg(),
	}
	self.locVars = append(self.locVars, newVar)
	self.locNames[name] = newVar
	return newVar.slot
}

func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("too many registers to alloc")
	}
	if self.usedRegs > self.maxRegs {
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1
}

func (self *funcInfo) freeReg() {
	self.usedRegs--
}

func (self *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		self.allocReg()
	}
	return self.usedRegs - n
}

func (self *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := self.constants[k]; found {
		return idx
	}
	idx := len(self.constants)
	self.constants[k] = idx
	return idx
}

func (self *funcInfo) indexOfUpval(name string) int {
	if upval, found := self.upvalues[name]; found {
		return upval.index
	}
	if self.parent != nil {
		if locVar, found := self.parent.locNames[name]; found {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{
				localVarSlot: locVar.slot,
				upvalIndex:   -1,
				index:        idx,
			}
			locVar.captured = true
			return idx
		}
		if uvIdx := (self).indexOfUpval(name); uvIdx > 0 {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{
				localVarSlot: -1,
				upvalIndex:   uvIdx,
				index:        idx,
			}
			return idx
		}
	}
	return -1
}

func (self *funcInfo) fixSbx(pc, sBx int) {
	i := self.insts[pc]
	i = i << 18 >> 18
	i = i | uint32(sBx+MAXARG_sBx)<<14
	self.insts[pc] = i
}

func (self *funcInfo) pc() int {
	return len(self.insts) - 1
}

func (self *funcInfo) getJmpArgA() int {

}
