package state

import (
	"api"
	"fmt"
)

func (self *luaState) PushNil() {
	self.stack.push(nil)
}

func (self *luaState) PushBoolean(b bool) {
	self.stack.push(b)
}

func (self *luaState) PushInteger(n int64) {
	self.stack.push(n)
}

func (self *luaState) PushNumber(n float64) {
	self.stack.push(n)
}

func (self *luaState) PushFString(fmtS string, a ...interface{}) string {
	str := fmt.Sprintf(fmtS, a...)
	self.stack.push(str)
	return str
}

func (self *luaState) PushString(s string) {
	self.stack.push(s)
}

func (self *luaState) PushGoFunction(f api.GoFunction) {
	self.stack.push(newGoClosure(f, 0))
}

func (self *luaState) PushGlobalTable() {
	self.stack.push(self.registry.get(api.LUA_RIDX_GLOBALS))
}

func (self *luaState) PushGoClosure(f api.GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := self.stack.pop()
		closure.upvals[n-1] = &upvalue{&val}
	}
	self.stack.push(closure)
}
