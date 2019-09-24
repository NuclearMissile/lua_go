package vm

import "lua_go/api"

func vararg(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a++
	if b != 1 {
		vm.LoadVararg(b - 1)
		popResults(a, b, vm)
	}
}

func self(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	b++
	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

func tForCall(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a++
	pushFuncAndArgs(a, 3, vm)
	vm.Call(2, c)
	popResults(a+3, c+1, vm)
}

func tForLoop(i Instruction, vm api.LuaVM) {
	a, sBx := i.AsBx()
	a++
	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}

// todo tail call optimization?
func tailCall(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	c = 0
	nArgs := pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	popResults(a, c, vm)
}

func closure(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a++
	vm.LoadProto(bx)
	vm.Replace(a)
}

func call(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a++
	nArgs := pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	popResults(a, c, vm)
}

func _return(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a++
	if b == 1 {
		return
	}
	if b > 1 {
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		fixStack(a, vm)
	}
}

func pushFuncAndArgs(a, b int, vm api.LuaVM) (nArgs int) {
	if b >= 1 {
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1
	} else {
		fixStack(a, vm)
		return vm.GetTop() - vm.RegisterCount() - 1
	}
}

func fixStack(a int, vm api.LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)
	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

func popResults(a, c int, vm api.LuaVM) {
	if c == 1 {
		return
	}
	if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}
