package main

import (
	"fmt"
	"lua_go/api"
	"lua_go/binchunk"
	"lua_go/state"
	"lua_go/vm"
)

func main() {
	ls := state.New()
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)
	printStack(ls)

	ls.Arith(api.LUA_OPADD)
	printStack(ls)
	ls.Arith(api.LUA_OPBNOT)
	printStack(ls)
	ls.Len(2)
	printStack(ls)
	ls.Concat(3)
	printStack(ls)
	ls.PushBoolean(ls.Compare(1, 2, api.LUA_OPEQ))
	printStack(ls)
}

//func main() {
//	ls := state.New()
//	printStack(ls)
//	ls.PushBoolean(true)
//	printStack(ls)
//	ls.PushInteger(10)
//	printStack(ls)
//	ls.PushNil()
//	printStack(ls)
//	ls.PushString("hello")
//	printStack(ls)
//	ls.PushValue(-4)
//	printStack(ls)
//	ls.Replace(3)
//	printStack(ls)
//	ls.SetTop(6)
//	printStack(ls)
//	ls.Remove(-3)
//	printStack(ls)
//	ls.SetTop(-5)
//	printStack(ls)
//}

func printStack(ls api.LuaState) {
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case api.LUA_TBOOLEAN:
			fmt.Printf("[%t]", ls.ToBoolean(i))
		case api.LUA_TNUMBER:
			fmt.Printf("[%g]", ls.ToNumber(i))
		case api.LUA_TSTRING:
			fmt.Printf("[%q]", ls.ToString(i))
		default:
			fmt.Printf("[%s]", ls.TypeName(t))
		}
	}
	fmt.Println()
}

// decompiler main()
//func main() {
//	if len(os.Args) > 1 {
//		data, err := ioutil.ReadFile(os.Args[1])
//		if err != nil {
//			panic(err)
//		}
//		proto := binchunk.Undump(data)
//		printProto(proto)
//	}
//}

func printProto(proto *binchunk.Prototype) {
	printHeader(proto)
	printCode(proto)
	printDetail(proto)
	for _, p := range proto.Protos {
		printProto(p)
	}
}

func printHeader(prototype *binchunk.Prototype) {
	funcType := "main"
	if prototype.LineDefined > 0 {
		funcType = "function"
	}
	varargFlag := ""
	if prototype.IsVararg > 0 {
		varargFlag = "+"
	}
	fmt.Printf("\n%s <%s:%d,%d>\n", funcType,
		prototype.Source, prototype.LineDefined, prototype.LastLineDefined)
	fmt.Printf("%d%s params, %d slots, %d upvalues, ",
		prototype.NumParams, varargFlag, prototype.MaxStackSize, len(prototype.Upvalues))
	fmt.Printf("%d locals, %d constants, %d functions\n", len(prototype.LocVars),
		len(prototype.Constants), len(prototype.Protos))
	fmt.Printf("Instructions (%d)\n", len(prototype.Code))
	fmt.Println("\ti no.\tline no.name\t\toperands")
}

func printDetail(prototype *binchunk.Prototype) {
	fmt.Printf("constants (%d):\n", len(prototype.Constants))
	for i, k := range prototype.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(k))
	}
	fmt.Printf("locals (%d):\n", len(prototype.LocVars))
	for i, locVar := range prototype.LocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, locVar.VarName, locVar.StartPC+1, locVar.EndPC+1)
	}
	fmt.Printf("upvalues (%d):\n", len(prototype.Upvalues))
	for i, upVal := range prototype.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, upValName(prototype, i), upVal.Instack, upVal.Idx)
	}
}

func upValName(prototype *binchunk.Prototype, idx int) string {
	if len(prototype.UpvalueNames) > 0 {
		return prototype.UpvalueNames[idx]
	}
	return "-"
}

func constantToString(k interface{}) string {
	switch k.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case string:
		return fmt.Sprintf("%q", k)
	default:
		return "?"
	}
}

func printCode(prototype *binchunk.Prototype) {
	for pc, c := range prototype.Code {
		line := ""
		if len(prototype.LineInfo) > 0 {
			line = fmt.Sprintf("%d", prototype.LineInfo[pc])
		}
		// fmt.Printf("\t%d\t[%s]\t0x%08X\n", pc+1, line, c)
		i := vm.Instruction(c)
		fmt.Printf("\t%d\t[%s]\t%s \t", pc+1, line, i.OpName())
		printOperands(i)
		fmt.Printf("\n")
	}
}

func printOperands(i vm.Instruction) {
	switch i.OpMode() {
	case vm.IABC:
		a, b, c := i.ABC()
		fmt.Printf("%d", a)
		if i.BMode() != vm.OpArgN {
			if b > 0xFF {
				fmt.Printf(" %d", -1-b&0xFF)
			} else {
				fmt.Printf(" %d", b)
			}
		}
		if i.CMode() != vm.OpArgN {
			if c > 0xFF {
				fmt.Printf(" %d", -1-c&0xFF)
			} else {
				fmt.Printf(" %d", c)
			}
		}
	case vm.IABx:
		a, bx := i.ABx()
		fmt.Printf("%d", a)
		if i.BMode() == vm.OpArgK {
			fmt.Printf(" %d", -1-bx)
		} else if i.BMode() == vm.OpArgU {
			fmt.Printf(" %d", bx)
		}
	case vm.IAsBx:
		a, sbx := i.AsBx()
		fmt.Printf("%d %d", a, sbx)
	case vm.IAx:
		ax := i.Ax()
		fmt.Printf("%d", -1-ax)
	default:
		panic("unknown instruction")
	}
}
