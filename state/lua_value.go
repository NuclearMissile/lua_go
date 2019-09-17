package state

import (
	"lua_go/api"
	"lua_go/number"
)

type luaValue interface{}

func typeOf(val luaValue) api.LuaType {
	switch val.(type) {
	case nil:
		return api.LUA_TNIL
	case bool:
		return api.LUA_TBOOLEAN
	case int64:
		return api.LUA_TNUMBER
	case float64:
		return api.LUA_TNUMBER
	case string:
		return api.LUA_TSTRING
	case *luaTable:
		return api.LUA_TTABLE
	case luaClosure:
		return api.LUA_TFUNCTION
	default:
		panic("unknown type of luaValue")
	}
}

func convertToFloat(val luaValue) (float64, bool) {
	switch x := val.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case string:
		return number.ParseFloat(x)
	default:
		return 0, false
	}
}

func convertToInteger(val luaValue) (int64, bool) {
	switch x := val.(type) {
	case int64:
		return x, true
	case float64:
		return number.FloatToInteger(x)
	case string:
		return stringToInteger(x)
	default:
		return 0, false
	}
}

func stringToInteger(str string) (int64, bool) {
	if i, ok := number.ParseInteger(str); ok {
		return i, true
	}
	if f, ok := number.ParseFloat(str); ok {
		return number.FloatToInteger(f)
	}
	return 0, false
}

func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}
