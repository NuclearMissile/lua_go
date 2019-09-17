package state

import (
	"lua_go/api"
	"lua_go/binchunk"
)

type closure struct {
	proto  *binchunk.Prototype
	goFunc api.GoFunction
}

func newGoClosure(f api.GoFunction) *closure {
	return &closure{goFunc: f}
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	return &closure{proto: proto}
}
