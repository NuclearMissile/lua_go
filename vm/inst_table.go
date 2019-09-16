package vm

import "lua_go/api"

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
	idx := int64(c * LFIELD_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j)
		vm.SetI(a, idx)
	}
}
