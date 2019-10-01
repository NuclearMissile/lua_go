package ast

type Block struct {
	LastLine int
	States   []Stat
	RetExps  []Exp
}
