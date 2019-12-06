package state

import "api"

// stack frame of lua
type luaStack struct {
	slots   []luaValue
	top     int
	pc      int
	prev    *luaStack
	closure *closure
	varargs []luaValue
	state   *luaState
	openuvs map[int]*upvalue
}

func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
		state: state,
	}
}

func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

func (self *luaStack) set(idx int, val luaValue) {
	if idx < api.LUA_REGISTRYINDEX {
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c != nil && uvIdx < len(c.upvals) {
			*(c.upvals[uvIdx].val) = val
		}
		return
	}

	if idx == api.LUA_REGISTRYINDEX {
		self.state.registry = val.(*luaTable)
		return
	}
	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		self.slots[absIdx-1] = val
		return
	}
	panic("illegal index")
}

func (self *luaStack) get(idx int) luaValue {
	if idx < api.LUA_REGISTRYINDEX {
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c == nil || uvIdx >= len(c.upvals) {
			return nil
		}
		return *(c.upvals[uvIdx].val)
	}

	if idx == api.LUA_REGISTRYINDEX {
		return self.state.registry
	}
	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		return self.slots[absIdx-1]
	}
	return nil
}

func (self *luaStack) isValid(idx int) bool {
	if idx < api.LUA_REGISTRYINDEX {
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		return c != nil && uvIdx < len(c.upvals)
	}

	if idx == api.LUA_REGISTRYINDEX {
		return true
	}
	absIdx := self.absIndex(idx)
	return absIdx > 0 && absIdx <= self.top
}

func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 || idx <= api.LUA_REGISTRYINDEX {
		return idx
	}
	return idx + self.top + 1
}

func (self *luaStack) pop() luaValue {
	if self.top < 1 {
		panic("stack underflow")
	}
	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = nil
	return val
}

func (self *luaStack) push(val luaValue) {
	if self.top == len(self.slots) {
		panic("stack overflow")
	}
	self.slots[self.top] = val
	self.top++
}

func (self *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = self.pop()
	}
	return vals
}

func (self *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}
	for i := 0; i < n; i++ {
		if i < nVals {
			self.push(vals[i])
		} else {
			self.push(nil)
		}
	}
}

func (self *luaStack) check(n int) {
	free := len(self.slots)
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}
