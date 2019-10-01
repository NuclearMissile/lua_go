package parser

import (
	. "lua_go/complier/ast"
	. "lua_go/complier/lexer"
)

func parseBlock(lexer *Lexer) *Block {
	return &Block{
		LastLine: lexer.Line(),
		States:   parseStates(lexer),
		RetExps:  parseRetExps(lexer),
	}
}

func parseRetExps(lexer *Lexer) []Exp {
	if lexer.LookAhead() != TOKEN_KW_RETURN {
		return nil
	}
	lexer.NextToken()
	switch lexer.LookAhead() {
	case TOKEN_EOF, TOKEN_KW_END, TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return []Exp{}
	case TOKEN_SEP_SEMI:
		lexer.NextToken()
		return []Exp{}
	default:
		exps := parseExpList(lexer)
		if lexer.LookAhead() == TOKEN_SEP_SEMI {
			lexer.NextToken()
		}
		return exps
	}
}

func parseExpList(lexer *Lexer) []Exp {
	exps := make([]Exp, 0, 4)
	exps = append(exps, parseExp(lexer))
	for lexer.LookAhead() == TOKEN_SEP_COMMA {
		lexer.NextToken()
		exps = append(exps, parseRetExps(lexer))
	}
	return exps
}

func parseExp(lexer *Lexer) Exp {

}

func parseStat(lexer *Lexer) Stat {

}

func parseStates(lexer *Lexer) []Stat {
	stats := make([]Stat, 0, 8)
	for !isReturnOrBlockEnd(lexer.LookAhead()) {
		stat := parseStat(lexer)
		if _, ok := stat.(*EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

func isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case TOKEN_KW_RETURN, TOKEN_EOF, TOKEN_KW_END, TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return true
	}
	return false
}
