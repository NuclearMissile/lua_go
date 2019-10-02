package parser

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
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

}

func parseFuncDefExp(lexer *Lexer) *FuncDefExp {

}

func parsePrefixExp(lexer *Lexer) Exp {

}
