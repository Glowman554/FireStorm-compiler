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
}

func (cf *CompiledFunction) findVariable(name string) (value.Value, types.Type) {
	if v, ok := cf.variables[name]; ok {
		return v, v.ElemType
	}
	panic("Variable " + name + " not found!")
}
