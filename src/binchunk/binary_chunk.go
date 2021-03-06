package binchunk

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

type Upvalue struct {
	Instack, Idx byte
}

type LocVar struct {
	VarName        string
	StartPC, EndPC uint32
}

type Prototype struct {
	Source                            string
	LineDefined, LastLineDefined      uint32
	NumParams, IsVararg, MaxStackSize byte
	Code                              []uint32
	Constants                         []interface{}
	Upvalues                          []Upvalue
	Protos                            []*Prototype
	LineInfo                          []uint32
	LocVars                           []LocVar
	UpvalueNames                      []string
}

type header struct {
	signature       [4]byte
	version, format byte
	luacData        [6]byte
	cintSize, sizetSize, instructionSize,
	luaIntegerSize, luaNumberSize byte
	luacInt int64
	luacNum float64
}

func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()
	reader.readByte()
	return reader.readProto("")
}

func Dump(proto *Prototype) []byte {
	w := &writer{}
	w.writeHeader()
	w.writeByte(byte(len(proto.Upvalues)))
	w.writeProto(proto, "")
	return w.data()
}

func List(proto *Prototype, full bool) string {
	p := &printer{buf: make([]string, 0, 64)}
	return p.printFunc(proto, full)
}

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == LUA_SIGNATURE
}

func StripDebug(proto *Prototype) {
	proto.Source = ""
	proto.LineInfo = nil
	proto.LocVars = nil
	proto.UpvalueNames = nil
	for _, p := range proto.Protos {
		StripDebug(p)
	}
}
