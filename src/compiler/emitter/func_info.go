package emitter

import (
	"compiler/ast"
	"fmt"
	"vm"
)

type funcInfo struct {
	instBuf
	constants                  map[interface{}]int
	usedRegs, maxRegs, scopeLv int
	line, lastLine, numParams  int
	locVars                    []*localVarInfo
	locVarNames                map[string]*localVarInfo
	breaks                     [][]int
	parent                     *funcInfo
	upvalues                   map[string]upvalInfo
	subFuncs                   []*funcInfo
	isVararg                   bool
	labels                     map[string]labelInfo
	gotos                      []*gotoInfo
}

type localVarInfo struct {
	prev                          *localVarInfo
	scopeLv, slot, startPC, endPC int
	name                          string
	isCaptured                    bool
}

type upvalInfo struct {
	locValSlot, upvalIndex, index int
}

type labelInfo struct {
	line, pc, scopeLv int
}

type gotoInfo struct {
	jmpPC, scopeLv int
	label          string
	pending        bool
}

func newFuncInfo(parent *funcInfo, funcDefExp *ast.FuncDefExp) *funcInfo {
	return &funcInfo{
		instBuf: instBuf{
			insts:    make([]uint32, 0, 8),
			lineNums: make([]uint32, 0, 8),
		},
		constants:   map[interface{}]int{},
		locVars:     make([]*localVarInfo, 0, 8),
		locVarNames: map[string]*localVarInfo{},
		breaks:      make([][]int, 1),
		parent:      parent,
		upvalues:    map[string]upvalInfo{},
		subFuncs:    make([]*funcInfo, 0, 8),
		isVararg:    funcDefExp.IsVararg,
		line:        funcDefExp.Line,
		lastLine:    funcDefExp.LastLine,
		numParams:   len(funcDefExp.Params),
		labels:      map[string]labelInfo{},
		gotos:       nil,
	}
}

// Constants
func (fi *funcInfo) indexOfConstant(konst interface{}) int {
	if idx, ok := fi.constants[konst]; ok {
		return idx
	}
	idx := len(fi.constants)
	fi.constants[konst] = idx
	return idx
}

// Registers
func (fi *funcInfo) allocReg() int {
	fi.usedRegs++
	if fi.usedRegs >= 255 {
		panic("too many regs are needed.")
	}
	if fi.usedRegs > fi.maxRegs {
		fi.maxRegs = fi.usedRegs
	}
	return fi.usedRegs - 1
}

func (fi *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		fi.allocReg()
	}
	return fi.usedRegs - n
}

func (fi *funcInfo) freeReg() {
	fi.usedRegs--
}

func (fi *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		fi.freeReg()
	}
}

// Upvalues
func (fi *funcInfo) indexOfUpval(name string) int {
	if upval, ok := fi.upvalues[name]; ok {
		return upval.index
	}
	if fi.parent != nil {
		if localVar, ok := fi.parent.locVarNames[name]; ok {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locValSlot: localVar.slot,
				upvalIndex: -1,
				index:      idx,
			}
			return idx
		}
		if upvalIdx, ok := fi.upvalues[name]; ok {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locValSlot: -1,
				upvalIndex: upvalIdx.index,
				index:      idx,
			}
			return idx
		}
	}
	return -1
}

func (fi *funcInfo) closeOpenUpvals(line int) {
	a := fi.getJmpArgA()
	if a > 0 {
		fi.emitJmp(line, a, 0)
	}
}

func (fi *funcInfo) getJmpArgA() int {
	hasLocalVarCaptured := false
	minSlotOfLocalVars := fi.maxRegs
	for _, locVar := range fi.locVarNames {
		if locVar.scopeLv == fi.scopeLv {
			for v := locVar; v != nil && v.scopeLv == fi.scopeLv; v = v.prev {
				if v.isCaptured {
					hasLocalVarCaptured = true
				}
				if v.slot < minSlotOfLocalVars && v.name[0] != '(' {
					minSlotOfLocalVars = v.slot
				}
			}
		}
	}
	if hasLocalVarCaptured {
		return minSlotOfLocalVars + 1
	} else {
		return 0
	}
}

// scope
func (fi *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &localVarInfo{
		prev:    fi.locVarNames[name],
		scopeLv: fi.scopeLv,
		slot:    fi.allocReg(),
		name:    name,
		startPC: startPC,
		endPC:   0,
	}
	fi.locVars = append(fi.locVars, newVar)
	fi.locVarNames[name] = newVar
	return newVar.slot
}

func (fi *funcInfo) slotOfLocalVar(name string) int {
	if localVar, ok := fi.locVarNames[name]; ok {
		return localVar.slot
	}
	return -1
}

func (fi *funcInfo) enterScope(breakable bool) {
	fi.scopeLv++
	if breakable {
		fi.breaks = append(fi.breaks, []int{})
	} else {
		fi.breaks = append(fi.breaks, nil)
	}
}

func (fi *funcInfo) exitScope(endPC int) {
	tempBreakJmps := fi.breaks[len(fi.breaks)-1]
	fi.breaks = fi.breaks[:len(fi.breaks)-1]
	a := fi.getJmpArgA()
	for _, pc := range tempBreakJmps {
		sBx := fi.pc() - pc
		i := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
		fi.insts[pc] = uint32(i)
	}
	fi.setGotoJmp()
	fi.scopeLv--
	for _, localVar := range fi.locVars {
		if fi.scopeLv < localVar.scopeLv {
			localVar.endPC = endPC
			fi.removeLocalVar(localVar)
		}
	}
}

func (fi *funcInfo) addBreakJmp(pc int) {
	for i := fi.scopeLv; i >= 0; i-- {
		if fi.breaks[i] != nil {
			fi.breaks[i] = append(fi.breaks[i], pc)
			return
		}
	}
	panic("<break> is not inside a loop")
}

func (fi *funcInfo) removeLocalVar(localVar *localVarInfo) {
	fi.freeReg()
	if localVar.prev == nil {
		delete(fi.locVarNames, localVar.name)
	} else if localVar.prev.scopeLv == localVar.scopeLv {
		fi.removeLocalVar(localVar.prev)
	} else {
		fi.locVarNames[localVar.name] = localVar.prev
	}
}

// label and goto
func (fi *funcInfo) addLabel(label string, line int) {
	key := fmt.Sprintf("%s@%d", label, fi.scopeLv)
	if labelInfo, ok := fi.labels[key]; ok {
		panic(fmt.Sprintf("label '%s' already defined on line '%d'", label, labelInfo.line))
	}
	fi.labels[label] = labelInfo{
		line:    line,
		pc:      fi.pc() + 1,
		scopeLv: fi.scopeLv,
	}
}

func (fi *funcInfo) addGoto(jmpPC, scopeLv int, label string) {
	fi.gotos = append(fi.gotos, &gotoInfo{
		jmpPC:   jmpPC,
		scopeLv: scopeLv,
		label:   label,
		pending: false,
	})
}

func (fi *funcInfo) setGotoJmp() {
	for i, gotoInfo := range fi.gotos {
		if gotoInfo == nil || gotoInfo.scopeLv < fi.scopeLv {
			continue
		}
		if gotoInfo.scopeLv == fi.scopeLv && gotoInfo.pending {
			continue
		}

		dstPC := fi.getGotoDst(gotoInfo.label)
		if dstPC >= 0 {
			if dstPC > gotoInfo.jmpPC && dstPC < fi.pc() {
				for _, locVar := range fi.locVarNames {
					if locVar.startPC > gotoInfo.jmpPC && locVar.startPC <= dstPC {
						panic(fmt.Sprintf("<goto %s> at line %d jumps into the scope of local '%s'",
							gotoInfo.label, fi.lineNums[gotoInfo.jmpPC], locVar.name))
					}
				}
			}

			a := 0
			for _, locVar := range fi.locVars {
				if locVar.startPC > dstPC {
					a = locVar.slot + 1
					break
				}
			}

			sBx := dstPC - gotoInfo.jmpPC - 1
			inst := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
			fi.insts[gotoInfo.jmpPC] = uint32(inst)
			fi.gotos[i] = nil
		} else if fi.scopeLv == 0 {
			panic(fmt.Sprintf("no legal label '%s' for <goto> at line %d",
				gotoInfo.label, fi.lineNums[gotoInfo.jmpPC]))
		} else {
			gotoInfo.pending = true
		}
	}

	for key, labelInfo := range fi.labels {
		if labelInfo.scopeLv == fi.scopeLv {
			delete(fi.labels, key)
		}
	}
}

func (fi *funcInfo) getGotoDst(lable string) int {
	for i := fi.scopeLv; i >= 0; i-- {
		key := fmt.Sprintf("%s@%d", lable, i)
		if li, ok := fi.labels[key]; ok {
			return li.pc
		}
	}
	return -1
}

// debug info
func (fi *funcInfo) setEndPC(name string, delta int) {
	for i := len(fi.locVars) - 1; i <= 0; i-- {
		localVar := fi.locVars[i]
		if localVar.name == name {
			localVar.endPC += delta
			return
		}
	}
}
