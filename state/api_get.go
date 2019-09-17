package state

import "lua_go/api"

func (self *luaState) CreateTable(nArr, nRec int) {
	self.stack.push(newLuaTable(nArr, nRec))
}

func (self *luaState) NewTable() {
	self.CreateTable(0, 0)
}

func (self *luaState) GetTable(idx int) api.LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k)
}

func (self *luaState) getTable(t, k luaValue) api.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		self.stack.push(v)
		return typeOf(v)
	}
	panic("not a table")
}

func (self *luaState) GetField(idx int, k string) api.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k)
}

func (self *luaState) GetI(idx int, i int64) api.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i)
}

func (self *luaState) GetGlobal(name string) api.LuaType {
	t := self.registry.get(api.LUA_RIDX_GLOBALS)
	return self.getTable(t, name)
}
