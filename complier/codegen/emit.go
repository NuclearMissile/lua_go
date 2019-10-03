package codegen

import "lua_go/vm"

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
