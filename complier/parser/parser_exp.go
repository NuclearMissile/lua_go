package parser

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
	"lua_go/number"
)

func parseExpList(lexer *Lexer) []Exp {
	exps := make([]Exp, 0, 4)
	exps = append(exps, parseExp(lexer))
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		exps = append(exps, parseExp(lexer))
	}
	return exps
}

func parseExp(lexer *Lexer) Exp {
	exp := parseExp11(lexer)
	for lexer.LookAhead() == TOKEN_OP_OR {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp11(lexer),
		}
	}
	return exp
}

func parseExp11(lexer *Lexer) Exp {
	exp := parseExp10(lexer)
	for lexer.LookAhead() == TOKEN_OP_AND {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp10(lexer),
		}
	}
	return exp
}

func parseExp10(lexer *Lexer) Exp {
	exp := parseExp9(lexer)
	for {
		switch lexer.LookAhead() {
		case TOKEN_OP_LT, TOKEN_OP_GT, TOKEN_OP_GE, TOKEN_OP_LE, TOKEN_OP_NE, TOKEN_OP_EQ:
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{
				Line: line,
				Op:   op,
				Exp1: exp,
				Exp2: parseExp9(lexer),
			}
		default:
			return exp
		}
	}
}

func parseExp9(lexer *Lexer) Exp {
	exp := parseExp8(lexer)
	for lexer.LookAhead() == TOKEN_OP_BOR {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp8(lexer),
		}
	}
	return exp
}

func parseExp8(lexer *Lexer) Exp {
	exp := parseExp7(lexer)
	for lexer.LookAhead() == TOKEN_OP_BXOR {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp7(lexer),
		}
	}
	return exp
}

func parseExp7(lexer *Lexer) Exp {
	exp := parseExp6(lexer)
	for lexer.LookAhead() == TOKEN_OP_BAND {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp6(lexer),
		}
	}
	return exp
}

func parseExp6(lexer *Lexer) Exp {
	exp := parseExp5(lexer)
	for lexer.LookAhead() == TOKEN_OP_SHL || lexer.LookAhead() == TOKEN_OP_SHR {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp5(lexer),
		}
	}
	return exp
}

func parseExp5(lexer *Lexer) Exp {
	exp := parseExp4(lexer)
	if lexer.LookAhead() != TOKEN_OP_CONCAT {
		return exp
	}
	line := 0
	exps := []Exp{exp}
	for lexer.LookAhead() == TOKEN_OP_CONCAT {
		line, _, _ = lexer.NextToken()
		exps = append(exps, parseExp4(lexer))
	}
	return &ConcatExp{
		Line:    line,
		ExpList: exps,
	}
}

func parseExp4(lexer *Lexer) Exp {
	exp := parseExp3(lexer)
	for lexer.LookAhead() == TOKEN_OP_ADD || lexer.LookAhead() == TOKEN_OP_SUB {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp3(lexer),
		}
	}
	return exp
}

func parseExp3(lexer *Lexer) Exp {
	exp := parseExp2(lexer)
	for {
		switch lexer.LookAhead() {
		case TOKEN_OP_MUL, TOKEN_OP_DIV, TOKEN_OP_IDIV, TOKEN_OP_MOD:
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{
				Line: line,
				Op:   op,
				Exp1: exp,
				Exp2: parseExp2(lexer),
			}
		default:
			return exp
		}
	}
}

func parseExp2(lexer *Lexer) Exp {
	switch lexer.LookAhead() {
	case TOKEN_OP_UNM, TOKEN_OP_BNOT, TOKEN_OP_LEN, TOKEN_OP_NOT:
		line, op, _ := lexer.NextToken()
		return optimizeUnaryOp(&UnopExp{
			Line: line,
			Op:   op,
			Exp:  parseExp2(lexer),
		})
	}
	return parseExp1(lexer)
}

func parseExp1(lexer *Lexer) Exp {
	exp := parseExp0(lexer)
	if lexer.LookAhead() == TOKEN_OP_POW {
		line, op, _ := lexer.NextToken()
		exp = &BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp2(lexer),
		}
	}
	return exp
}

func parseExp0(lexer *Lexer) Exp {
	switch lexer.LookAhead() {
	case TOKEN_VARARG:
		line, _, _ := lexer.NextToken()
		return &VarargExp{Line: line}
	case TOKEN_KW_NIL:
		line, _, _ := lexer.NextToken()
		return &NilExp{Line: line}
	case TOKEN_KW_TRUE:
		line, _, _ := lexer.NextToken()
		return &TrueExp{Line: line}
	case TOKEN_KW_FALSE:
		line, _, _ := lexer.NextToken()
		return &FalseExp{Line: line}
	case TOKEN_STRING:
		line, _, token := lexer.NextToken()
		return &StringExp{
			Line: line,
			Str:  token,
		}
	case TOKEN_NUMBER:
		return parseNumberExp(lexer)
	case TOKEN_SEP_LCURLY:
		return parseTableCtorExp(lexer)
	case TOKEN_KW_FUNCTION:
		lexer.NextToken()
		return parseFuncDefExp(lexer)
	default:
		return parsePrefixExp(lexer)
	}
}

func parseNumberExp(lexer *Lexer) Exp {
	line, _, token := lexer.NextToken()
	if i, ok := number.ParseInteger(token); ok {
		return &IntegerExp{
			Line: line,
			Val:  i,
		}
	} else if f, ok := number.ParseFloat(token); ok {
		return &FloatExp{
			Line: line,
			Val:  f,
		}
	}
	panic("not a number:" + token)
}

func parseFuncDefExp(lexer *Lexer) *FuncDefExp {
	line := lexer.Line()
	lexer.NextTokenOfKind(TOKEN_SEP_LPAREN)
	paraList, isVararg := parseParaList(lexer)
	lexer.NextTokenOfKind(TOKEN_SEP_RPAREN)
	block := parseBlock(lexer)
	lastLine, _ := lexer.NextTokenOfKind(TOKEN_KW_END)
	return &FuncDefExp{
		Line:     line,
		LastLine: lastLine,
		ParaList: paraList,
		IsVararg: isVararg,
		Block:    block,
	}
}

func parsePrefixExp(lexer *Lexer) Exp {
	var exp Exp
	if lexer.LookAhead() == TOKEN_IDENTIFIER {
		line, _, name := lexer.NextToken()
		exp = &NameExp{
			Line: line,
			Name: name,
		}
	} else {
		exp = parseParensExp(lexer)
	}
	return finishPrefixExp(lexer, exp)
}

func finishPrefixExp(lexer *Lexer, exp Exp) Exp {
	for {
		switch lexer.LookAhead() {
		case TOKEN_SEP_LBRACK:
			lexer.NextToken()
			keyExp := parseExp(lexer)
			lexer.NextTokenOfKind(TOKEN_SEP_RBRACK)
			exp = &TableAccessExp{
				LastLine:  lexer.Line(),
				PrefixExp: exp,
				KeyExp:    keyExp,
			}
		case TOKEN_SEP_DOT:
			lexer.NextToken()
			line, name := lexer.NextIdentifier()
			keyExp := &StringExp{
				Line: line,
				Str:  name,
			}
			exp = &TableAccessExp{
				LastLine:  line,
				PrefixExp: exp,
				KeyExp:    keyExp,
			}
		case TOKEN_SEP_COLON, TOKEN_SEP_LPAREN, TOKEN_SEP_LCURLY, TOKEN_STRING:
			exp = finishFuncCall(lexer, exp)
		default:
			return exp
		}
	}
}

func finishFuncCall(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	parseArgs := func(lexer *Lexer) (args []Exp) {
		switch lexer.LookAhead() {
		case TOKEN_SEP_LPAREN:
			lexer.NextToken()
			if lexer.LookAhead() != TOKEN_SEP_RPAREN {
				args = parseExpList(lexer)
			}
			lexer.NextTokenOfKind(TOKEN_SEP_RPAREN)
		case TOKEN_SEP_LCURLY:
			args = []Exp{parseTableCtorExp(lexer)}
		default:
			line, s := lexer.NextTokenOfKind(TOKEN_STRING)
			args = []Exp{&StringExp{
				Line: line,
				Str:  s,
			}}
		}
		return
	}

	parseNameExp := func(lexer *Lexer) *StringExp {
		if lexer.LookAhead() == TOKEN_SEP_COLON {
			lexer.NextToken()
			line, name := lexer.NextIdentifier()
			return &StringExp{
				Line: line,
				Str:  name,
			}
		}
		return nil
	}

	nameExp := parseNameExp(lexer)
	line := lexer.Line()
	args := parseArgs(lexer)
	lastLine := lexer.Line()
	return &FuncCallExp{
		Line:      line,
		LastLine:  lastLine,
		PrefixExp: prefixExp,
		NameExp:   nameExp,
		ArgList:   args,
	}
}

func parseParensExp(lexer *Lexer) Exp {
	lexer.NextTokenOfKind(TOKEN_SEP_LPAREN)
	exp := parseExp(lexer)
	lexer.NextTokenOfKind(TOKEN_SEP_RPAREN)
	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{Exp: exp}
	}
	return exp
}

func parseParaList(lexer *Lexer) (names []string, isVararg bool) {
	switch lexer.LookAhead() {
	case TOKEN_SEP_RPAREN:
		return nil, false
	case TOKEN_VARARG:
		lexer.NextToken()
		return nil, true
	}
	_, name := lexer.NextIdentifier()
	names = append(names, name)
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		if lexer.LookAhead() == TOKEN_IDENTIFIER {
			_, name := lexer.NextIdentifier()
			names = append(names, name)
		} else {
			lexer.NextTokenOfKind(TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return
}

func parseTableCtorExp(lexer *Lexer) *TableCtorExp {
	line := lexer.Line()
	lexer.NextTokenOfKind(TOKEN_SEP_LCURLY)
	keyExps, valExps := parseFieldList(lexer)
	lexer.NextTokenOfKind(TOKEN_SEP_RCURLY)
	lastLine := lexer.Line()
	return &TableCtorExp{
		Line:       line,
		LastLine:   lastLine,
		KeyExpList: keyExps,
		ValExpList: valExps,
	}
}

func parseFieldList(lexer *Lexer) (keyExps, valExps []Exp) {
	if lexer.LookAhead() != TOKEN_SEP_RCURLY {
		k, v := parseField(lexer)
		keyExps = append(keyExps, k)
		valExps = append(valExps, v)
		for isFieldSep(lexer.LookAhead()) {
			lexer.NextToken()
			if lexer.LookAhead() != TOKEN_SEP_RCURLY {
				k, v := parseField(lexer)
				keyExps = append(keyExps, k)
				valExps = append(valExps, v)
			} else {
				break
			}
		}
	}
	return
}

func isFieldSep(tokenKind int) bool {
	return tokenKind == TOKEN_SEP_COMMA || tokenKind == TOKEN_SEP_SEMI
}

func parseField(lexer *Lexer) (key, val Exp) {
	if lexer.LookAhead() == TOKEN_SEP_LBRACK {
		lexer.NextToken()
		key = parseExp(lexer)
		lexer.NextTokenOfKind(TOKEN_SEP_RBRACK)
		lexer.NextTokenOfKind(TOKEN_OP_ASSIGN)
		val = parseExp(lexer)
		return
	}

	exp := parseExp(lexer)
	if nameExp, ok := exp.(*NameExp); ok {
		if lexer.LookAhead() == TOKEN_OP_ASSIGN {
			lexer.NextToken()
			key = &StringExp{
				Line: nameExp.Line,
				Str:  nameExp.Name,
			}
			val = parseExp(lexer)
		}
	}
	return nil, exp
}
