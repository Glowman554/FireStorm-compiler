package llvm

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type CompiledFunction struct {
	variables       map[string]*ir.InstAlloca
	returnBlock     *ir.Block
	returnIncomings []*ir.Incoming
	returnType      types.Type
	name            string
}

func (cf *CompiledFunction) findVariable(name string, err func(string, *CompiledFunction)) (value.Value, types.Type) {
	if v, ok := cf.variables[name]; ok {
		return v, v.ElemType
	}
	err("Variable "+name+" not found!", cf)
	panic("?")
}
