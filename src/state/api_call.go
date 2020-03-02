package state

import (
	"api"
	"binchunk"
	"compiler"
	"vm"
)

func (self *luaState) Load(chunk []byte, name, mode string) int {
	var proto *binchunk.Prototype
	if binchunk.IsBinaryChunk(chunk) {
		if mode != "b" {
			panic("Source mode is not 'b'")
		}
		proto = binchunk.Undump(chunk)
	} else {
		if mode != "t" {
			panic("Source mode is not 't'")
		}
		proto = compiler.Compile(string(chunk), name)
	}
	c := newLuaClosure(proto)
	self.stack.push(c)
	if len(proto.Upvalues) > 0 {
		env := self.registry.get(api.LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return 0
}

func (self *luaState) Call(nArgs, nResults int) {
	val := self.stack.get(-(nArgs + 1))
	c, ok := val.(*closure)
	if !ok {
		if mf := getMetaField(val, "__call", self); mf != nil {
			if c, ok = mf.(*closure); ok {
				self.stack.push(val)
				self.Insert(-(nArgs + 2))
				nArgs++
			}
		}
	}
	if ok {
		if c.proto != nil {
			self.callLuaClosure(nArgs, nResults, c)
			//fmt.Printf("call lua closure: %s<%d, %d>\n", c.proto.Source, c.proto.LineDefined, c.proto.LastLineDefined)
		} else {
			self.callGoClosure(nArgs, nResults, c)
		}
	}
}

func (self *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(self.Fetch())
		inst.Execute(self)
		if inst.OpCode() == vm.OP_RETURN {
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
