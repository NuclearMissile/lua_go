package vm

import (
	"api"
)

func getUpVal(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Copy(vm.UpvalueIndex(b), a)
}

func setUpVal(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a++
	b++
	vm.Copy(a, vm.UpvalueIndex(b))
}

func getTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	vm.GetRK(c)
	vm.GetTable(vm.UpvalueIndex(b))
	vm.Replace(a)
}

func setTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(vm.UpvalueIndex(a))
}
