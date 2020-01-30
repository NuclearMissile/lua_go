package emitter

import "vm"

type funcInfo struct {
	constants                  map[interface{}]int
	usedRegs, maxRegs, scopeLv int
	localVars                  []*localVarInfo
	localVarNames              map[string]*localVarInfo
	breakTable                 [][]int
}

type localVarInfo struct {
	prev              *localVarInfo
	scopeLv, regIndex int
	name              string
	isCaptured        bool
}

func (self *funcInfo) addLocalVar(name string) int {
	newVar := &localVarInfo{
		prev:       self.localVarNames[name],
		scopeLv:    self.scopeLv,
		regIndex:   self.allocReg(),
		name:       name,
		isCaptured: false,
	}
	self.localVars = append(self.localVars, newVar)
	self.localVarNames[name] = newVar
	return newVar.regIndex
}

func (self *funcInfo) slotOfLocalVar(name string) int {
	if localVar, ok := self.localVarNames[name]; ok {
		return localVar.regIndex
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
	if idx, ok := self.constants[konst]; ok {
		return idx
	}
	idx := len(self.constants)
	self.constants[konst] = idx
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
		delete(self.localVarNames, localVar.name)
	} else if localVar.prev.scopeLv == localVar.scopeLv {
		self.removeLocalVar(localVar.prev)
	} else {
		self.localVarNames[localVar.name] = localVar.prev
	}
}

func (self *funcInfo) getJmpArgA() interface{} {

}
