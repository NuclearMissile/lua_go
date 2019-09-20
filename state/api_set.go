package state

import "lua_go/api"

func (self *luaState) SetTable(idx int) {
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(self.stack.get(idx), k, v, false)
}

func (self *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.set(k, v)
			return
		}
	}

	if !raw {
		if mf := getMetaField(t, "__newindex", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				self.setTable(x, k, v, false)
				return
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.stack.push(v)
				self.Call(3, 0)
				return
			}
		}
	}
	panic("not a table")
}

func (self *luaState) SetField(idx int, k string) {
	self.setTable(self.stack.get(idx), k, self.stack.pop(), false)
}

func (self *luaState) SetI(idx int, i int64) {
	self.setTable(self.stack.get(idx), i, self.stack.pop(), false)
}

func (self *luaState) SetGlobal(name string) {
	t := self.registry.get(api.LUA_RIDX_GLOBALS)
	v := self.stack.pop()
	self.setTable(t, name, v, false)
}

func (self *luaState) Register(name string, f api.GoFunction) {
	self.PushGoFunction(f)
	self.SetGlobal(name)
}

func (self *luaState) SetMetatable(idx int) {
	val := self.stack.get(idx)
	mtVal := self.stack.pop()
	if mtVal == nil {
		setMetatable(val, nil, self)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, self)
	} else {
		panic("table expected")
	}
}

func (self *luaState) RawSet(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, v, k, true)
}

func (self *luaState) RawSetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, true)
}
