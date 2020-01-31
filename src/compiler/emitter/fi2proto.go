package emitter

import . "binchunk"

func (self *funcInfo) toProto() *Prototype {
	proto := &Prototype{
		LineDefined:     uint32(self.line),
		LastLineDefined: uint32(self.lastLine),
		NumParams:       byte(self.numParams),
		Code:            self.insts,
		Constants:       self.getConstants(),
		Upvalues:        self.getUpvalues(),
		Protos:          toProtos(self.subFuncs),
		LineInfo:        self.lineNums,
		LocVars:         self.getLocVars(),
		UpvalueNames:    self.getUpvalueNames(),
	}
	if self.line == 0 {
		proto.LastLineDefined = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2
	}
	if self.isVararg {
		proto.IsVararg = 1
	}
	return proto
}

func toProtos(fis []*funcInfo) []*Prototype {
	protos := make([]*Prototype, len(fis))
	for i, fi := range fis {
		protos[i] = fi.toProto()
	}
	return protos
}

func (self *funcInfo) getConstants() []interface{} {
	consts := make([]interface{}, len(self.constants))
	for k, idx := range self.constants {
		consts[idx] = k
	}
	return consts
}

func (self *funcInfo) getUpvalues() []Upvalue {
	upvals := make([]Upvalue, len(self.upvalues))
	for _, uv := range self.upvalues {
		if uv.locValSlot >= 0 {
			upvals[uv.index] = Upvalue{
				Instack: 1,
				Idx:     byte(uv.locValSlot),
			}
		} else {
			upvals[uv.index] = Upvalue{
				Instack: 0,
				Idx:     byte(uv.upvalIndex),
			}
		}
	}
	return upvals
}

func (self *funcInfo) getLocVars() []LocVar {
	locVars := make([]LocVar, len(self.locVars))
	for i, locVar := range self.locVars {
		locVars[i] = LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func (self *funcInfo) getUpvalueNames() []string {
	names := make([]string, len(self.upvalues))
	for name, uv := range self.upvalues {
		names[uv.index] = name
	}
	return names
}
