package parser

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
)

// MUX
func parseStat(lexer *Lexer) Stat {
	switch lexer.LookAhead() {
	case TOKEN_SEP_SEMI:
		return parseEmptyStat(lexer)
	case TOKEN_KW_BREAK:
		return parseBreakStat(lexer)
	case TOKEN_KW_GOTO:
		return parseGotoStat(lexer)
	case TOKEN_KW_DO:
		return parseDoStat(lexer)
	case TOKEN_SEP_LABEL:
		return parseLabelStat(lexer)
	case TOKEN_KW_WHILE:
		return parseWhileStat(lexer)
	case TOKEN_KW_REPEAT:
		return parseRepeatStat(lexer)
	case TOKEN_KW_IF:
		return parseIfStat(lexer)
	case TOKEN_KW_FOR:
		return parseForStat(lexer)
	case TOKEN_KW_FUNCTION:
		return parseFuncDefStat(lexer)
	case TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(lexer)
	default:
		return parseAssignOrFuncCallStat(lexer)
	}
}

func parseEmptyStat(lexer *Lexer) *EmptyStat {
	lexer.NextTokenOfKind(TOKEN_SEP_SEMI)
	return &EmptyStat{}
}

func parseBreakStat(lexer *Lexer) *BreakStat {
	lexer.NextTokenOfKind(TOKEN_KW_BREAK)
	return &BreakStat{Line: lexer.Line()}
}

func parseGotoStat(lexer *Lexer) *GotoStat {
	lexer.NextTokenOfKind(TOKEN_KW_GOTO)
	_, name := lexer.NextIdentifier()
	return &GotoStat{Name: name}
}

func parseDoStat(lexer *Lexer) *DoStat {
	lexer.NextTokenOfKind(TOKEN_KW_DO)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &DoStat{Block: block}
}

func parseLabelStat(lexer *Lexer) *LabelStat {
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL)
	_, name := lexer.NextIdentifier()
	lexer.NextTokenOfKind(TOKEN_SEP_LABEL)
	return &LabelStat{Name: name}
}

func parseWhileStat(lexer *Lexer) *WhileStat {
	lexer.NextTokenOfKind(TOKEN_KW_WHILE)
	exp := parseExp(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_DO)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &WhileStat{
		Exp:   exp,
		Block: block,
	}
}

func parseRepeatStat(lexer *Lexer) *RepeatStat {
	lexer.NextTokenOfKind(TOKEN_KW_REPEAT)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_UNTIL)
	exp := parseExp(lexer)
	return &RepeatStat{
		Block: block,
		Exp:   exp,
	}
}

func parseIfStat(lexer *Lexer) *IfStat {
	exps := make([]Exp, 0, 4)
	blocks := make([]*Block, 0, 4)
	lexer.NextTokenOfKind(TOKEN_KW_IF)
	exps = append(exps, parseExp(lexer))
	lexer.NextTokenOfKind(TOKEN_KW_THEN)
	blocks = append(blocks, parseBlock(lexer))
	for lexer.LookAhead() == TOKEN_KW_ELSEIF {
		lexer.NextToken()
		exps = append(exps, parseExp(lexer))
		lexer.NextTokenOfKind(TOKEN_KW_THEN)
		blocks = append(blocks, parseBlock(lexer))
	}
	if lexer.LookAhead() == TOKEN_KW_ELSE {
		lexer.NextToken()
		exps = append(exps, parseExp(lexer))
		blocks = append(blocks, parseBlock(lexer))
	}
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &IfStat{
		ExpList:   exps,
		BlockList: blocks,
	}
}

func parseForStat(lexer *Lexer) Stat {
	lineOfFor, _ := lexer.NextTokenOfKind(TOKEN_KW_FOR)
	_, name := lexer.NextIdentifier()
	if lexer.LookAhead() == TOKEN_OP_ASSIGN {
		return finishForNumStat(lexer, lineOfFor, name)
	} else {
		return finishForInStat(lexer, name)
	}
}

func finishForInStat(lexer *Lexer, name0 string) *ForInStat {
	nameList := finishNameList(lexer, name0)
	lexer.NextTokenOfKind(TOKEN_KW_IN)
	expList := parseExpList(lexer)
	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &ForInStat{
		LineOfDo: lineOfDo,
		NameList: nameList,
		ExpList:  expList,
		Block:    block,
	}
}

func finishNameList(lexer *Lexer, name0 string) []string {
	names := []string{name0}
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		_, name := lexer.NextIdentifier()
		names = append(names, name)
	}
	return names
}

func finishForNumStat(lexer *Lexer, lineOfFor int, name string) *ForNumStat {
	lexer.NextTokenOfKind(TOKEN_OP_ASSIGN)
	initExp := parseExp(lexer)
	lexer.NextTokenOfKind(TOKEN_SEP_COMMA)
	limitExp := parseExp(lexer)

	var stepExp Exp
	if lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		stepExp = parseExp(lexer)
	} else {
		stepExp = &IntegerExp{
			Line: lexer.Line(),
			Val:  1,
		}
	}

	lineOfDo, _ := lexer.NextTokenOfKind(TOKEN_KW_DO)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_KW_END)
	return &ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   name,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

func parseLocalAssignOrFuncDefStat(lexer *Lexer) Stat {
	lexer.NextTokenOfKind(TOKEN_KW_LOCAL)
	if lexer.LookAhead() == TOKEN_KW_FUNCTION {
		return finishLocalFuncDefStat(lexer)
	} else {
		return finishLocalVarDeclStat(lexer)
	}
}

func finishLocalFuncDefStat(lexer *Lexer) *LocalFuncDefStat {
	lexer.NextTokenOfKind(TOKEN_KW_FUNCTION)
	_, name := lexer.NextIdentifier()
	fdExp := parseFuncDefExp(lexer)
	return &LocalFuncDefStat{
		Name: name,
		Exp:  fdExp,
	}
}

func finishLocalVarDeclStat(lexer *Lexer) *LocalVarDeclStat {
	_, name0 := lexer.NextIdentifier()
	nameList := finishNameList(lexer, name0)
	var expList []Exp
	if lexer.LookAhead() == TOKEN_OP_ASSIGN {
		lexer.NextToken()
		expList = parseExpList(lexer)
	}
	lastLine := lexer.Line()
	return &LocalVarDeclStat{
		LastLine: lastLine,
		NameList: nameList,
		ExpList:  expList,
	}
}

func parseAssignOrFuncCallStat(lexer *Lexer) Stat {
	prefixExp := parsePrefixExp(lexer)
	if fc, ok := prefixExp.(*FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(lexer, prefixExp)
	}
}

func parseAssignStat(lexer *Lexer, var0 Exp) *AssignStat {
	varList := finishVarList(lexer, var0)
	lexer.NextTokenOfKind(TOKEN_OP_ASSIGN)
	expList := parseExpList(lexer)
	lastLine := lexer.Line()
	return &AssignStat{
		LastLine: lastLine,
		VarList:  varList,
		ExpList:  expList,
	}
}

func finishVarList(lexer *Lexer, exp Exp) []Exp {
	varList := []Exp{checkVar(lexer, exp)}
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		exp := parsePrefixExp(lexer)
		varList = append(varList, checkVar(lexer, exp))
	}
	return varList
}

func checkVar(lexer *Lexer, exp Exp) Exp {
	switch exp.(type) {
	case *NameExp, *TableAccessExp:
		return exp
	}
	lexer.NextTokenOfKind(-1)
	panic("unreachable")
}

func parseFuncDefStat(lexer *Lexer) *AssignStat {
	lexer.NextTokenOfKind(TOKEN_KW_FUNCTION)
	fnExp, hasColon := parseFuncName(lexer)
	fdExp := parseFuncDefExp(lexer)
	if hasColon {
		fdExp.ParaList = append(fdExp.ParaList, "")
		copy(fdExp.ParaList[1:], fdExp.ParaList)
		fdExp.ParaList[0] = "self"
	}
	return &AssignStat{
		LastLine: fdExp.Line,
		VarList:  []Exp{fnExp},
		ExpList:  []Exp{fdExp},
	}
}

func parseFuncName(lexer *Lexer) (exp Exp, hasColon bool) {
	line, name := lexer.NextIdentifier()
	exp = &NameExp{
		Line: line,
		Name: name,
	}
	for lexer.LookAhead() == TOKEN_SEP_DOT {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{
			Line: line,
			Str:  name,
		}
		exp = &TableAccessExp{
			LastLine:  line,
			PrefixExp: exp,
			KeyExp:    idx,
		}
	}
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		idx := &StringExp{
			Line: line,
			Str:  name,
		}
		exp = &TableAccessExp{
			LastLine:  line,
			PrefixExp: exp,
			KeyExp:    idx,
		}
		hasColon = true
	}
	return
}
