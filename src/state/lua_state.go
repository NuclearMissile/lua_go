package state

import "api"

type luaState struct {
	registry *luaTable
	stack    *luaStack
}

func New() *luaState {
	registry := newLuaTable(0, 0)
	registry.set(api.LUA_RIDX_GLOBALS, newLuaTable(0, 0))
	ls := &luaState{registry: registry}
	ls.pushLuaStack(newLuaStack(api.LUA_MINSATCK, ls))
	return ls
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

//func New(stackSize int, proto *binchunk.Prototype) *luaState {
//	return &luaState{
//		stack: newLuaStack(stackSize),
//		proto: proto,
//		pc:    0,
//	}
//}
