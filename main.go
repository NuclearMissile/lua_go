package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lua_go/api"
	"lua_go/binchunk"
	"lua_go/compiler/parser"
	"lua_go/vm"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		testParser(string(data), os.Args[1])
	}
}

func testParser(chunk, chunkName string) {
	ast := parser.Parse(chunk, chunkName)
	d, err := json.MarshalIndent(ast, "", "\t")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("ast.json", d, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d))
}

//func testLexer(chunk, chunkName string) {
//	lexer := NewLexer(chunk, chunkName)
//	for {
//		line, kind, token := lexer.NextToken()
//		fmt.Printf("[%2d] [%-10s] %s\n", line, kindToCat(kind), token)
//		if kind == TOKEN_EOF {
//			break
//		}
//	}
//}
//
//func kindToCat(kind int) string {
//	switch {
//	case kind < TOKEN_SEP_SEMI:
//		return "other"
//	case kind <= TOKEN_SEP_RCURLY:
//		return "separator"
//	case kind <= TOKEN_OP_NOT:
//		return "operator"
//	case kind <= TOKEN_KW_WHILE:
//		return "keyword"
//	case kind <= TOKEN_IDENTIFIER:
//		return "identifier"
//	case kind <= TOKEN_NUMBER:
//		return "number"
//	case kind <= TOKEN_STRING:
//		return "string"
//	default:
//		return "other"
//	}
//}

//func main() {
//	if len(os.Args) > 1 {
//		data, err := ioutil.ReadFile(os.Args[1])
//		if err != nil {
//			panic(err)
//		}
//		ls := state.New()
//		ls.Register("print", print)
//		ls.Register("getmetatable", getMetatable)
//		ls.Register("setmetatable", setMetatable)
//		ls.Register("next", next)
//		ls.Register("pairs", pairs)
//		ls.Register("ipairs", iPairs)
//		ls.Register("pcall", pCall)
//		ls.Register("error", _error)
//		ls.Load(data, os.Args[1], "b")
//		ls.Call(0, 0)
//	}
//}

func pCall(ls api.LuaState) int {
	nArgs := ls.GetTop() - 1
	status := ls.PCall(nArgs, -1, 0)
	ls.PushBoolean(status == api.LUA_OK)
	ls.Insert(1)
	return ls.GetTop()
}

func _error(ls api.LuaState) int {
	return ls.Error()
}

func iPairs(ls api.LuaState) int {
	aux := func(ls api.LuaState) int {
		i := ls.ToInteger(2) + 1
		ls.PushInteger(i)
		if ls.GetI(1, i) == api.LUA_TNIL {
			return 1
		} else {
			return 2
		}
	}
	ls.PushGoFunction(aux)
	ls.PushValue(1)
	ls.PushInteger(0)
	return 3
}

func pairs(ls api.LuaState) int {
	ls.PushGoFunction(next)
	ls.PushValue(1)
	ls.PushNil()
	return 3
}

func next(ls api.LuaState) int {
	ls.SetTop(2)
	if ls.Next(1) {
		return 2
	} else {
		ls.PushNil()
		return 1
	}
}

func getMetatable(ls api.LuaState) int {
	if !ls.GetMetatable(1) {
		ls.PushNil()
	}
	return 1
}

func setMetatable(ls api.LuaState) int {
	ls.SetMetatable(1)
	return 1
}

func print(ls api.LuaState) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Print(ls.ToString(i))
		} else {
			fmt.Print(ls.TypeName(ls.Type(i)))
		}
		if i < nArgs {
			fmt.Print("\t")
		}
	}
	fmt.Println()
	return 0
}

//func luaMain(proto *binchunk.Prototype) {
//	nRegs := int(proto.MaxStackSize)
//	ls := state.New(nRegs+8, proto)
//	ls.SetTop(nRegs)
//	for {
//		pc := ls.PC()
//		inst := vm.Instruction(ls.Fetch())
//		if inst.Opcode() != vm.OP_RETURN {
//			inst.Execute(ls)
//			fmt.Printf("[%02d] %s ", pc+1, inst.OpName())
//			printStack(ls)
//		} else {
//			break
//		}
//	}
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
