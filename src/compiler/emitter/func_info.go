package emitter

import (
	"compiler/ast"
	"vm"
)

type funcInfo struct {
	constantMap                           map[interface{}]int
	usedRegs, maxRegs, scopeLv, numParams int
	localVars                             []*localVarInfo
	localVarMap                           map[string]*localVarInfo
	breakTable                            [][]int
	parent                                *funcInfo
	upvalMap                              map[string]upvalInfo
	insts                                 []uint32
	subFuncs                              []*funcInfo
	isVararg                              bool
}

func newFuncInfo(parent *funcInfo, funcDefExp *ast.FuncDefExp) *funcInfo {
	return &funcInfo{
		constantMap: map[interface{}]int{},
		numParams:   len(funcDefExp.ParList),
		localVars:   make([]*localVarInfo, 0, 8),
		localVarMap: map[string]*localVarInfo{},
		breakTable:  make([][]int, 0, 1),
		parent:      parent,
		upvalMap:    map[string]upvalInfo{},
		insts:       make([]uint32, 0, 8),
		subFuncs:    []*funcInfo{},
		isVararg:    funcDefExp.IsVararg,
	}
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

func (self *funcInfo) indexOfUpval(name string) int {
	if upval, ok := self.upvalMap[name]; ok {
		return upval.index
	}
	if self.parent != nil {
		if localVar, ok := self.parent.localVarMap[name]; ok {
			idx := len(self.upvalMap)
			self.upvalMap[name] = upvalInfo{
				locValSlot: localVar.slot,
				upvalIndex: -1,
				index:      idx,
			}
			return idx
		}
		if upvalIdx, ok := self.upvalMap[name]; ok {
			idx := len(self.upvalMap)
			self.upvalMap[name] = upvalInfo{
				locValSlot: -1,
				upvalIndex: upvalIdx.index,
				index:      idx,
			}
			return idx
		}
	}
	return -1
}

func (self *funcInfo) addLocalVar(name string) int {
	newVar := &localVarInfo{
		prev:       self.localVarMap[name],
		scopeLv:    self.scopeLv,
		slot:       self.allocReg(),
		name:       name,
		isCaptured: false,
	}
	self.localVars = append(self.localVars, newVar)
	self.localVarMap[name] = newVar
	return newVar.slot
}

func (self *funcInfo) slotOfLocalVar(name string) int {
	if localVar, ok := self.localVarMap[name]; ok {
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

func (self *funcInfo) exitScope() {
	tempBreakJmps := self.breakTable[len(self.breakTable)-1]
	self.breakTable = self.breakTable[:len(self.breakTable)-1]
	a := self.getJmpArgA()
	for _, pc := range tempBreakJmps {
		sBx := self.pc() - pc
		i := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
		self.insts[pc] = uint32(i)
	}
	self.scopeLv--
	for _, localVar := range self.localVars {
		if self.scopeLv < localVar.scopeLv {
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

func (self *funcInfo) indexOfConstant(konst interface{}) int {
	if idx, ok := self.constantMap[konst]; ok {
		return idx
	}
	idx := len(self.constantMap)
	self.constantMap[konst] = idx
	return idx
}

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

func (self *funcInfo) removeLocalVar(localVar *localVarInfo) {
	self.freeReg()
	if localVar.prev == nil {
		delete(self.localVarMap, localVar.name)
	} else if localVar.prev.scopeLv == localVar.scopeLv {
		self.removeLocalVar(localVar.prev)
	} else {
		self.localVarMap[localVar.name] = localVar.prev
	}
}

func (self *funcInfo) getJmpArgA() int {
	hasLocalVarCaptured := false
	minSlotOfLocalVars := self.maxRegs
	for _, locVar := range self.localVarMap {
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
