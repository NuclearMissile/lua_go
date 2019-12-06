package vm

import "api"

const LFIELD_PER_FLUSH = 50

func newTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	vm.CreateTable(FloatByte2Int(b), FloatByte2Int(c))
	vm.Replace(a)
}

func getTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

func setTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(a)
}

func setList(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	if c > 0 {
		c--
	} else {
		c = Instruction(vm.Fetch()).Ax()
	}

	bIs0 := b == 0
	if bIs0 {
		b = int(vm.ToInteger(-1)) - a - 1
		vm.Pop(1)
	}
	vm.CheckStack(1)
	idx := int64(c * LFIELD_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j)
		vm.SetI(a, idx)
	}
	if bIs0 {
		for j := vm.RegisterCount() + 1; j <= vm.GetTop(); j++ {
			idx++
			vm.PushValue(j)
			vm.SetI(a, idx)
		}
		vm.SetTop(vm.RegisterCount())
	}
}
