package compiler

import (
	"binchunk"
	"codegen"
	"compiler/parser"
)

func Compile(chunk, chunkName string) *binchunk.Prototype {
	ast := parser.Parse(chunk, chunkName)
	proto := codegen.GenProto(ast)
	setChunkName(proto, chunkName)
	return proto
}

func setChunkName(proto *binchunk.Prototype, chunkName string) {
	proto.Source = chunkName
	for _, p := range proto.Protos {
		setChunkName(p, chunkName)
	}
}
