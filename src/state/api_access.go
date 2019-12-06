package state

import (
	"api"
	"fmt"
)

func (self *luaState) Type(idx int) api.LuaType {
	if self.stack.isValid(idx) {
		val := self.stack.get(idx)
		return typeOf(val)
	}
	return api.LUA_TNONE
}

func (self *luaState) ToBoolean(idx int) bool {
	val := self.stack.get(idx)
	return convertToBoolean(val)
}

func (self *luaState) IsInteger(idx int) bool {
	val := self.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

func (self *luaState) ToInteger(idx int) int64 {
	n, _ := self.ToIntegerX(idx)
	return n
}

func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	return convertToInteger(self.stack.get(idx))
}

func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

func (self *luaState) ToNumberX(idx int) (float64, bool) {
	return convertToFloat(self.stack.get(idx))
}

func (self *luaState) ToString(idx int) string {
	s, _ := self.ToStringX(idx)
	return s
}

func (self *luaState) ToStringX(idx int) (string, bool) {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		self.stack.set(idx, s)
		return s, true
	default:
		return "", false
	}
}

func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return t == api.LUA_TSTRING || t == api.LUA_TNUMBER
}

func (self *luaState) IsNone(idx int) bool {
	return self.Type(idx) == api.LUA_TNONE
}

func (self *luaState) IsNil(idx int) bool {
	return self.Type(idx) == api.LUA_TNIL
}

func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= api.LUA_TNIL
}

func (self *luaState) IsBoolean(idx int) bool {
	return self.Type(idx) == api.LUA_TBOOLEAN
}

func (self *luaState) ToGoFunction(idx int) api.GoFunction {
	val := self.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}
	return nil
}

func (self *luaState) IsGoFunction(idx int) bool {
	return self.ToGoFunction(idx) != nil
}

func (self *luaState) TypeName(tp api.LuaType) string {
	switch tp {
	case api.LUA_TNONE:
		return "no value"
	case api.LUA_TNIL:
		return "nil"
	case api.LUA_TBOOLEAN:
		return "boolean"
	case api.LUA_TNUMBER:
		return "number"
	case api.LUA_TSTRING:
		return "string"
	case api.LUA_TTABLE:
		return "table"
	case api.LUA_TFUNCTION:
		return "function"
	case api.LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}
