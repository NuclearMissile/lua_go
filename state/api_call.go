package state

import (
	"lua_go/api"
	"lua_go/binchunk"
	"lua_go/vm"
)

func (self *luaState) Load(chunk []byte, name, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	self.stack.push(c)
	return 0
}

func (self *luaState) Call(nArgs, nResults int) {
	val := self.stack.get(-(nArgs + 1))
	if c, ok := val.(*closure); ok {
		if c.proto != nil {
			self.callLuaClosure(nArgs, nResults, c)
		} else {
			self.callGoClosure(nArgs, nResults, c)
		}
		// fmt.Printf("call %s<%d, %d>\n", c.proto.Source, c.proto.LineDefined, c.proto.LastLineDefined)
	} else {
		panic("not a function")
	}
}

func (self *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(self.Fetch())
		inst.Execute(self)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}

func (self *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+api.LUA_MINSATCK, self)
	newStack.closure = c
	args := self.stack.popN(nArgs)
	newStack.pushN(args, nArgs)
	self.stack.pop()
	self.pushLuaStack(newStack)
	goResNum := c.goFunc(self)
	self.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(goResNum)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1
	newStack := newLuaStack(nRegs+api.LUA_MINSATCK, self)
	newStack.closure = c

	funcAndArgs := self.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	self.pushLuaStack(newStack)
	self.runLuaClosure()
	self.popLuaStack()
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}
