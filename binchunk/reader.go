package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte
}

func (self *reader) readByte() byte {
	b := self.data[0]
	self.data = self.data[1:]
	return b
}
func (self *reader) readBytes(n uint) []byte {
	bytes := self.data[:n]
	self.data = self.data[n:]
	return bytes
}

func (self *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(self.data)
	self.data = self.data[4:]
	return i
}

func (self *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(self.data)
	self.data = self.data[8:]
	return i
}

func (self *reader) readLuaInteger() int64 {
	return int64(self.readUint64())
}

func (self *reader) readLuaNumber() float64 {
	return math.Float64frombits(self.readUint64())
}
func (self *reader) readString() string {
	size := uint(self.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(self.readUint64())
	}
	bytes := self.readBytes(size - 1)
	return string(bytes)
}

func (self *reader) checkHeader() {
	if string(self.readBytes(4)) != LUA_SIGNATURE {
		panic("note a precompiled chunk")
	}
	if self.readByte() != LUAC_VERSION {
		panic("version mismatch")
	}
	if self.readByte() != LUAC_FORMAT {
		panic("format mismatch")
	}
	if string(self.readBytes(6)) != LUAC_DATA {
		panic("corrupted")
	}
	if self.readByte() != CINT_SIZE {
		panic("int size mismatch")
	}
	if self.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch")
	}
	if self.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch")
	}
	if self.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch")
	}
	if self.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch")
	}
	if self.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch")
	}
	if self.readLuaNumber() != LUAC_NUM {
		panic("float format size mismatch")
	}
}

func (self *reader) readProto(parentSource string) *Prototype {
	source := self.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     self.readUint32(),
		LastLineDefined: self.readUint32(),
		NumParams:       self.readByte(),
		IsVararg:        self.readByte(),
		MaxStackSize:    self.readByte(),
		Code:            self.readCode(),
		Constants:       self.readConstants(),
		Upvalues:        self.readUpvalues(),
		Protos:          self.readProtos(source),
		LineInfo:        self.readLineInfo(),
		LocVars:         self.readLocVars(),
		UpvalueNames:    self.readUpValueNames(),
	}
}

func (self *reader) readCode() []uint32 {
	code := make([]uint32, self.readUint32())
	for i := range code {
		code[i] = self.readUint32()
	}
	return code
}

func (self *reader) readConstants() []interface{} {
	constants := make([]interface{}, self.readUint32())
	for i := range constants {
		constants[i] = self.readUint32()
	}
	return constants
}

func (self *reader) readUpvalues() []Upvalue {

}

func (self *reader) readProtos(s string) []*Prototype {

}

func (self *reader) readLocVars() []LocVar {

}

func (self *reader) readLineInfo() []uint32 {

}

func (self *reader) readUpValueNames() []string {

}
