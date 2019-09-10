package main

import (
	"fmt"
	"io/ioutil"
	"lua_go/binchunk"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		proto := binchunk.Undump(data)
		list(proto)
	}
}

func list(proto *binchunk.Prototype) {
	printHeader(proto)
	printCode(proto)
	printDetail(proto)
	for _, p := range proto.Protos {
		list(p)
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
	fmt.Printf("\n%s <%s:%d,%d> (%d instructions)\n", funcType,
		prototype.Source, prototype.LineDefined, prototype.LastLineDefined, len(prototype.Code))
	fmt.Printf("%d%s params, %d slots, %d upvalues, ",
		prototype.NumParams, varargFlag, prototype.MaxStackSize, len(prototype.Upvalues))
	fmt.Printf("%d locals, %d constants, %d functions\n", len(prototype.LocVars),
		len(prototype.Constants), len(prototype.Protos))
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
		line := "-"
		if len(prototype.LineInfo) > 0 {
			line = fmt.Sprintf("%d", prototype.LineInfo[pc])
		}
		fmt.Printf("\t%d\t[%s]\t0x%08X\n", pc+1, line, c)
	}
}
