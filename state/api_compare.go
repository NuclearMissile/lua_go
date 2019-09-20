package state

import "lua_go/api"

func (self *luaState) RawEqual(idx1, idx2 int) bool {
	if !self.stack.isValid(idx1) || !self.stack.isValid(idx2) {
		return false
	}
	a := self.stack.get(idx1)
	b := self.stack.get(idx2)
	return _eq(a, b, nil)
}

func (self *luaState) Compare(idx1, idx2 int, op api.CompareOp) bool {
	if !self.stack.isValid(idx1) || !self.stack.isValid(idx2) {
		return false
	}
	a := self.stack.get(idx1)
	b := self.stack.get(idx2)
	switch op {
	case api.LUA_OPEQ:
		return _eq(a, b, self)
	case api.LUA_OPLE:
		return _le(a, b, self)
	case api.LUA_OPLT:
		return _lt(a, b, self)
	default:
		panic("invalid op")
	}
}

func _le(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return x <= float64(y)
		case float64:
			return x <= y
		}
	}
	if result, ok := callMetamethod(a, b, "__le", ls); ok {
		return convertToBoolean(result)
	} else if result, ok := callMetamethod(a, b, "__lt", ls); ok {
		return convertToBoolean(result)
	}

	panic("comparision error")
}

func _eq(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case *luaTable:
		if y, ok := b.(*luaTable); ok && x != y && ls != nil {
			if result, ok := callMetamethod(x, y, "__eq", ls); ok {
				return convertToBoolean(result)
			}
		}
		return a == b
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case int64:
		switch y := b.(type) {
		case int64:
			return x == y
		case float64:
			return float64(x) == y
		default:
			return false
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x == y
		case int64:
			return x == float64(y)
		default:
			return false
		}
	default:
		return a == b
	}
}

func _lt(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x < y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x < y
		case float64:
			return float64(x) < y
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return x < float64(y)
		case float64:
			return x < y
		}
	}
	if result, ok := callMetamethod(a, b, "__lt", ls); ok {
		return convertToBoolean(result)
	}
	panic("comparision error")
}
