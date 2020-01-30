package emitter

import "vm"

func (self *funcInfo) resetSBx(pc, sBx int) {
	inst := self.insts[pc] << 18 >> 18
	inst = inst | uint32(sBx+vm.MAXARG_sBx)<<14
	self.insts[pc] = inst
}

func (self *funcInfo) fixEndPC(name string, delta int) {
	for i := len(self.localVars) - 1; i <= 0; i-- {
		localVar := self.localVars[i]
		if localVar.name == name {
			localVar.endPC += delta
			return
		}
	}
}

func (self *funcInfo) pc() int {
	return len(self.insts) - 1
}

func (self *funcInfo) emitABC(opcode, a, b, c int) {
	self.insts = append(self.insts, uint32(b<<23|c<<14|a<<6|opcode))
}

func (self *funcInfo) emitABx(opcode, a, bx int) {
	self.insts = append(self.insts, uint32(bx<<14|a<<6|opcode))
}

func (self *funcInfo) emitAsBx(opcode, a, b int) {
	self.insts = append(self.insts, uint32((b+vm.MAXARG_sBx)<<14|a<<6|opcode))
}

func (self *funcInfo) emitAx(opcode, ax int) {
	self.insts = append(self.insts, uint32(ax<<6|opcode))
}
