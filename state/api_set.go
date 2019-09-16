package state

func (self *luaState) SetTable(idx int) {
	self.setTable(self.stack.get(idx), self.stack.pop(), self.stack.pop())
}

func (self *luaState) setTable(t, k, v luaValue) {
	if tbl, ok := t.(*luaTable); ok {
		tbl.set(k, v)
		return
	}
	panic("not a table")
}

func (self *luaState) SetField(idx int, k string) {
	self.setTable(self.stack.get(idx), k, self.stack.pop())
}

func (self *luaState) SetI(idx int, i int64) {
	self.setTable(self.stack.get(idx), i, self.stack.pop())
}
