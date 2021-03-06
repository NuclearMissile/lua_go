package vm

import "api"

const (
	MAXARG_Bx  = 1<<18 - 1      // 2^18 - 1 = 262143
	MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071
)

/*
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6| iABC
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6| iABx
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6| iAsBx
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6| iAx
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/

type Instruction uint32

func (self Instruction) Execute(vm api.LuaVM) {
	action := opcodes[self.OpCode()].action
	if action != nil {
		action(self, vm)
	} else {
		panic(self.OpName())
	}
}

func (self Instruction) OpCode() int {
	return int(self & 0x3F)
}

func (self Instruction) ABC() (a, b, c int) {
	a = int(self >> 6 & 0xFF)
	b = int(self >> 23 & 0x1FF)
	c = int(self >> 14 & 0x1FF)
	return
}

func (self Instruction) ABx() (a, bx int) {
	a = int(self >> 6 & 0xFF)
	bx = int(self >> 14)
	return
}

func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx
}

func (self Instruction) Ax() int {
	return int(self >> 6)
}

func (self Instruction) OpName() string {
	return opcodes[self.OpCode()].name
}

func (self Instruction) OpMode() byte {
	return opcodes[self.OpCode()].opMode
}

func (self Instruction) BMode() byte {
	return opcodes[self.OpCode()].argBMode
}

func (self Instruction) CMode() byte {
	return opcodes[self.OpCode()].argCMode
}
