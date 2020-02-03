package ast

type Stat interface{}

type EmptyStat struct{}

type BreakStat struct {
	Line int
}

type LabelStat struct {
	Line int
	Name string
}

type GotoStat struct {
	Line int
	Name string
}

type DoStat struct {
	Block *Block
}

type FuncCallStat = FuncCallExp

type WhileStat struct {
	Exp   Exp
	Block *Block
}

type RepeatStat struct {
	Block *Block
	Exp   Exp
}

type IfStat struct {
	Exps   []Exp
	Blocks []*Block
}

type ForNumStat struct {
	LineOfFor int
	LineOfDo  int
	VarName   string
	InitExp   Exp
	LimitExp  Exp
	StepExp   Exp
	Block     *Block
}

type ForInStat struct {
	LineOfFor, LineOfDo int
	Names               []string
	Exps                []Exp
	Block               *Block
}

type LocalVarDeclStat struct {
	LastLine int
	Names    []string
	Exps     []Exp
}

type AssignStat struct {
	LastLine int
	Vars     []Exp
	Exps     []Exp
}

type LocalFuncDefStat struct {
	Name string
	Exp  *FuncDefExp
}
