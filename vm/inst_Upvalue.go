package vm

import "lua_go/api"

func getTabUp(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a++
	vm.PushGlobalTable()
	vm.GetRK(c)
	vm.GetTable(-2)
	vm.Replace(a)
	vm.Pop(1)
}