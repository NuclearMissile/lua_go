package vm

const MAXARG_Bx = 1<<18 - 1       // 2^18 - 1 = 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071

/*
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6| ABC
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6| ABx
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6|
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/

type Instruction uint32

func (self Instruction) Opcode() int {
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
