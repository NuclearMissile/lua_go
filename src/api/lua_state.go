package api

type LuaType = int
type ArithOp = int
type CompareOp = int
type ThreadStatus = int

type GoFunction func(LuaState) int

type LuaState interface {
	// AuxLib
	BasicAPI
}

type BasicAPI interface {
	// stack manipulations
	GetTop() int
	AbsIndex(idx int) int
	CheckStack(n int) bool
	Pop(n int)
	Copy(fromIdx, toIdx int)
	PushValue(idx int)
	Replace(idx int)
	Insert(idx int)
	Remove(idx int)
	// circular shift
	Rotate(idx, n int)
	SetTop(idx int)
	// access functions (stack -> go)
	TypeName(tp LuaType) string
	Type(idx int) LuaType
	IsNone(idx int) bool
	IsNil(idx int) bool
	IsNoneOrNil(idx int) bool
	IsBoolean(idx int) bool
	IsInteger(idx int) bool
	IsNumber(idx int) bool
	IsString(idx int) bool
	ToBoolean(idx int) bool
	ToInteger(idx int) int64
	ToIntegerX(idx int) (int64, bool)
	ToNumber(idx int) float64
	ToNumberX(idx int) (float64, bool)
	ToString(idx int) string
	ToStringX(idx int) (string, bool)
	// push functions (go -> stack)
	PushNil()
	PushBoolean(b bool)
	PushInteger(i int64)
	PushNumber(f float64)
	PushString(s string)
	PushFString(fmt string, a ...interface{}) string
	// arithmetic functions
	Arith(op ArithOp)
	Compare(idx1, idx2 int, op CompareOp) bool
	Len(idx int)
	Concat(n int)
	// get functions (Lua -> stack)
	NewTable()
	CreateTable(nArr, nRec int)
	GetTable(idx int) LuaType
	GetField(idx int, k string) LuaType
	GetI(idx int, i int64) LuaType
	// set functions (stack - > Lua)
	SetTable(idx int)
	SetField(idx int, k string)
	SetI(idx int, i int64)
	// lua function api
	Load(chunk []byte, name, mode string) int
	Call(nArgs, nResults int)
	// go function api
	PushGoFunction(f GoFunction)
	IsGoFunction(idx int) bool
	ToGoFunction(idx int) GoFunction
	// global table api
	PushGlobalTable()
	GetGlobal(name string) LuaType
	SetGlobal(name string)
	Register(name string, f GoFunction)
	// go upvalue support
	PushGoClosure(f GoFunction, n int)
	UpvalueIndex(idx int) int
	// Metatable, metamethod
	GetMetatable(idx int) bool
	SetMetatable(idx int)
	RawLen(idx int) uint
	RawEqual(idx1, idx2 int) bool
	RawGet(idx int) LuaType
	RawSet(idx int)
	RawGetI(idx int, i int64) LuaType
	RawSetI(idx int, i int64)
	// iterator
	Next(idx int) bool
	// error handling
	Error() int
	PCall(nArgs, nRes, msgh int) int
}
