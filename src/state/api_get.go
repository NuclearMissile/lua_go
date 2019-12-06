package state

import "api"

func (self *luaState) CreateTable(nArr, nRec int) {
	self.stack.push(newLuaTable(nArr, nRec))
}

func (self *luaState) NewTable() {
	self.CreateTable(0, 0)
}

func (self *luaState) GetTable(idx int) api.LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, false)
}

func (self *luaState) getTable(t, k luaValue, raw bool) api.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			self.stack.push(v)
			return typeOf(v)
		}
	}
	if !raw {
		if mf := getMetaField(t, "__index", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return self.getTable(x, k, false)
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.Call(2, 1)
				v := self.stack.get(-1)
				return typeOf(v)
			}
		}
	}
	panic("not a table")
}

func (self *luaState) GetField(idx int, k string) api.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k, false)
}

func (self *luaState) GetI(idx int, i int64) api.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, false)
}

func (self *luaState) GetGlobal(name string) api.LuaType {
	t := self.registry.get(api.LUA_RIDX_GLOBALS)
	return self.getTable(t, name, false)
}

func (self *luaState) GetMetatable(idx int) bool {
	val := self.stack.get(idx)
	if mt := getMetatable(val, self); mt != nil {
		self.stack.push(mt)
		return true
	}
	return false
}

func (self *luaState) RawGet(idx int) api.LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, true)
}

func (self *luaState) RawGetI(idx int, i int64) api.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, true)
}
