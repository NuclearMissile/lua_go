package emitter

import (
	"compiler/ast"
	"vm"
)

type funcInfo struct {
	constants                  map[interface{}]int
	usedRegs, maxRegs, scopeLv int
	line, lastLine, numParams  int
	locVars                    []*localVarInfo
	locVarNames                map[string]*localVarInfo
	breakTable                 [][]int
	parent                     *funcInfo
	upvalues                   map[string]upvalInfo
	insts                      []uint32
	subFuncs                   []*funcInfo
	isVararg                   bool
	lineNums                   []uint32
}

type localVarInfo struct {
	prev                          *localVarInfo
	scopeLv, slot, startPC, endPC int
	name                          string
	isCaptured                    bool
}

type upvalInfo struct {
	locValSlot, upvalIndex, index int
}

type labelInfo struct {
}

type gotoInfo struct {
}

func newFuncInfo(parent *funcInfo, funcDefExp *ast.FuncDefExp) *funcInfo {
	return &funcInfo{
		constants:   map[interface{}]int{},
		locVars:     make([]*localVarInfo, 0, 8),
		locVarNames: map[string]*localVarInfo{},
		breakTable:  make([][]int, 1),
		parent:      parent,
		upvalues:    map[string]upvalInfo{},
		insts:       make([]uint32, 0, 8),
		subFuncs:    make([]*funcInfo, 0, 8),
		isVararg:    funcDefExp.IsVararg,
		lineNums:    make([]uint32, 0, 8),
		line:        funcDefExp.Line,
		lastLine:    funcDefExp.LastLine,
		numParams:   len(funcDefExp.ParList),
	}
}

// Constants
func (self *funcInfo) indexOfConstant(konst interface{}) int {
	if idx, ok := self.constants[konst]; ok {
		return idx
	}
	idx := len(self.constants)
	self.constants[konst] = idx
	return idx
}

// Registers
func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("too many regs are needed.")
	}
	if self.usedRegs > self.maxRegs {
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1
}

func (self *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		self.allocReg()
	}
	return self.usedRegs - n
}

func (self *funcInfo) freeReg() {
	self.usedRegs--
}

func (self *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		self.freeReg()
	}
}

// Upvalues
func (self *funcInfo) indexOfUpval(name string) int {
	if upval, ok := self.upvalues[name]; ok {
		return upval.index
	}
	if self.parent != nil {
		if localVar, ok := self.parent.locVarNames[name]; ok {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{
				locValSlot: localVar.slot,
				upvalIndex: -1,
				index:      idx,
			}
			return idx
		}
		if upvalIdx, ok := self.upvalues[name]; ok {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{
				locValSlot: -1,
				upvalIndex: upvalIdx.index,
				index:      idx,
			}
			return idx
		}
	}
	return -1
}

func (self *funcInfo) closeOpenUpvals(line int) {
	a := self.getJmpArgA()
	if a > 0 {
		self.emitJmp(line, a, 0)
	}
}

func (self *funcInfo) getJmpArgA() int {
	hasLocalVarCaptured := false
	minSlotOfLocalVars := self.maxRegs
	for _, locVar := range self.locVarNames {
		if locVar.scopeLv == self.scopeLv {
			for v := locVar; v != nil && v.scopeLv == self.scopeLv; v = v.prev {
				if v.isCaptured {
					hasLocalVarCaptured = true
				}
				if v.slot < minSlotOfLocalVars && v.name[0] != '(' {
					minSlotOfLocalVars = v.slot
				}
			}
		}
	}
	if hasLocalVarCaptured {
		return minSlotOfLocalVars + 1
	} else {
		return 0
	}
}

// scope
func (self *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &localVarInfo{
		prev:    self.locVarNames[name],
		scopeLv: self.scopeLv,
		slot:    self.allocReg(),
		name:    name,
		startPC: startPC,
		endPC:   0,
	}
	self.locVars = append(self.locVars, newVar)
	self.locVarNames[name] = newVar
	return newVar.slot
}

func (self *funcInfo) slotOfLocalVar(name string) int {
	if localVar, ok := self.locVarNames[name]; ok {
		return localVar.slot
	}
	return -1
}

func (self *funcInfo) enterScope(breakable bool) {
	self.scopeLv++
	if breakable {
		self.breakTable = append(self.breakTable, []int{})
	} else {
		self.breakTable = append(self.breakTable, nil)
	}
}

func (self *funcInfo) exitScope(endPC int) {
	tempBreakJmps := self.breakTable[len(self.breakTable)-1]
	self.breakTable = self.breakTable[:len(self.breakTable)-1]
	a := self.getJmpArgA()
	for _, pc := range tempBreakJmps {
		sBx := self.pc() - pc
		i := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
		self.insts[pc] = uint32(i)
	}
	self.scopeLv--
	for _, localVar := range self.locVars {
		if self.scopeLv < localVar.scopeLv {
			localVar.endPC = endPC
			self.removeLocalVar(localVar)
		}
	}
}

func (self *funcInfo) addBreakJmp(pc int) {
	for i := self.scopeLv; i >= 0; i-- {
		if self.breakTable[i] != nil {
			self.breakTable[i] = append(self.breakTable[i], pc)
			return
		}
	}
	panic("<break> not inside a loop")
}

func (self *funcInfo) removeLocalVar(localVar *localVarInfo) {
	self.freeReg()
	if localVar.prev == nil {
		delete(self.locVarNames, localVar.name)
	} else if localVar.prev.scopeLv == localVar.scopeLv {
		self.removeLocalVar(localVar.prev)
	} else {
		self.locVarNames[localVar.name] = localVar.prev
	}
}
