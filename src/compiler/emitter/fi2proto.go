package emitter

import . "binchunk"

func (fi *funcInfo) toProto() *Prototype {
	proto := &Prototype{
		LineDefined:     uint32(fi.line),
		LastLineDefined: uint32(fi.lastLine),
		NumParams:       byte(fi.numParams),
		MaxStackSize:    byte(fi.maxRegs),
		Code:            fi.insts,
		Constants:       fi.getConstants(),
		Upvalues:        fi.getUpvalues(),
		Protos:          toProtos(fi.subFuncs),
		LineInfo:        fi.lineNums,
		LocVars:         fi.getLocVars(),
		UpvalueNames:    fi.getUpvalueNames(),
	}
	if fi.line == 0 {
		proto.LastLineDefined = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2
	}
	if fi.isVararg {
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

func (fi *funcInfo) getConstants() []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func (fi *funcInfo) getUpvalues() []Upvalue {
	upvals := make([]Upvalue, len(fi.upvalues))
	for _, uv := range fi.upvalues {
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

func (fi *funcInfo) getLocVars() []LocVar {
	locVars := make([]LocVar, len(fi.locVars))
	for i, locVar := range fi.locVars {
		locVars[i] = LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func (fi *funcInfo) getUpvalueNames() []string {
	names := make([]string, len(fi.upvalues))
	for name, uv := range fi.upvalues {
		names[uv.index] = name
	}
	return names
}
