package emitter

import . "compiler/ast"

func (fi *funcInfo) evalBlock(block *Block) {
	for _, stat := range block.States {
		fi.evalStat(stat)
	}
	if block.RetExps != nil {
		fi.evalRetStat(block.RetExps, block.LastLine)
	}
}
