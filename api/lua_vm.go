package api

type LuaVM interface {
	LuaState
	PC() int
	AddPC(n int)
	Fetch() uint32
	GetConst(idx int)
	GetRK(rk int)
	LoadProto(idx int)
	RegisterCount() int
	LoadVararg(n int)
}
