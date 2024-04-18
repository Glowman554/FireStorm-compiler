package llvm

import (
	"flc/firestorm/parser"
	"flc/firestorm/utils"
	"fmt"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type LLVM struct {
	global          *parser.Node
	globalVariables map[string]*ir.Global
	functions       map[string]*ir.Func
	module          *ir.Module
	globalId        int
	ptrType         types.Type
	target          string
}

func NewLLVM(global *parser.Node, target string) *LLVM {
	return &LLVM{
		global:          global,
		globalVariables: make(map[string]*ir.Global),
		functions:       make(map[string]*ir.Func),
		globalId:        0,
		ptrType:         types.I64,
		target:          target,
	}
}

func (l *LLVM) findFunction(name string) *ir.Func {
	if f, ok := l.functions[name]; ok {
		return f
	}
	panic("Function " + name + " not found!")
}

func (l *LLVM) findVariable(name string, cf *CompiledFunction) (value.Value, types.Type) {
	if v, ok := l.globalVariables[name]; ok {
		return v, v.ContentType
	}
	return cf.findVariable(name)
}

func (b *LLVM) newGlobalString(v string) value.Value {
	id := "str." + strconv.Itoa(b.globalId)
	b.globalId++

	str := constant.NewCharArrayFromString(v + "\x00")
	globalStr := b.module.NewGlobalDef(id, str)

	zero := constant.NewInt(types.I64, 0)
	return constant.NewGetElementPtr(str.Typ, globalStr, zero, zero)
}

func (b *LLVM) compareToLLVM(c parser.Compare) enum.IPred {
	switch c {
	case parser.More:
		return enum.IPredSGT
	case parser.Less:
		return enum.IPredSLT
	case parser.MoreEquals:
		return enum.IPredSGE
	case parser.LessEquals:
		return enum.IPredSLE
	case parser.Equals:
		return enum.IPredEQ
	case parser.NotEquals:
		return enum.IPredNE
	}
	panic("?")
}

func(b*LLVM) datatypeArraySelect(d parser.UnnamedDatatype, single types.Type, array types.Type) types.Type {
    if d.IsArray {
        return array
    } else {
        return single 
    }
}

func (b *LLVM) datatypeToLLVM(d parser.UnnamedDatatype) types.Type {
	switch d.Type {
	case parser.INT:
        return b.datatypeArraySelect(d, types.I64, types.I64Ptr)
	case parser.STR:
        return b.datatypeArraySelect(d, types.I8Ptr, types.NewPointer(types.I8Ptr))
	case parser.VOID:
		return types.Void	
	case parser.CHR:
        return b.datatypeArraySelect(d, types.I8, types.I8Ptr)
	case parser.PTR:
		return b.ptrType
    case parser.INT_32:
        return b.datatypeArraySelect(d, types.I32, types.I32Ptr)
    case parser.INT_16:
        return b.datatypeArraySelect(d, types.I16, types.I16Ptr)
    default:
		panic("Invalid datatype")
	}
}

func (b *LLVM) generateExpressionRaw(exp *parser.Node, block *ir.Block, cf *CompiledFunction) value.Value {

	switch exp.Type {
	case parser.NUMBER:
		return constant.NewInt(types.I64, int64(exp.Value.(int)))
	case parser.STRING:
		return b.newGlobalString(exp.Value.(string))
	case parser.COMPARE:
		cmp := block.NewICmp(b.compareToLLVM(exp.Value.(parser.Compare)), b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
		return block.NewZExt(cmp, types.I64)
	case parser.NOT:
		cmp := block.NewICmp(enum.IPredEQ, b.generateExpression(exp.A, block, cf), constant.NewInt(types.I64, 0))
		return block.NewZExt(cmp, types.I64)
	case parser.ADD:
		return block.NewAdd(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.SUBTRACT:
		return block.NewSub(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.MULTIPLY:
		return block.NewMul(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.DIVIDE:
		return block.NewSDiv(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.MODULO:
		return block.NewSRem(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.OR:
		return block.NewOr(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.AND:
		return block.NewAnd(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.XOR:
		return block.NewXor(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.BIT_NOT:
		return block.NewXor(b.generateExpression(exp.A, block, cf), constant.NewInt(types.I64, -1))
	case parser.SHIFT_LEFT:
		return block.NewShl(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.SHIFT_RIGHT:
		return block.NewLShr(b.generateExpression(exp.A, block, cf), b.generateExpression(exp.B, block, cf))
	case parser.FUNCTION_CALL:
		fc := exp.Value.(parser.FunctionCall)
		return b.generateFunctionCall(fc, block, cf)
	case parser.VARIABLE_LOOKUP:
		v, t := b.findVariable(exp.Value.(string), cf)
		// if _, ok := v.ElemType.(*types.PointerType); ok {
		// 	l := block.NewLoad(v.ElemType, v)
		// 	return b.autoTypeCast(l, b.ptrType, block)
		// }
		return block.NewLoad(t, v)

	case parser.VARIABLE_LOOKUP_ARRAY:
		v, t := b.findVariable(exp.Value.(string), cf)
		ptr := block.NewLoad(t, v)
		i := b.generateExpression(exp.A, block, cf)
		indexed := block.NewGetElementPtr(ptr.ElemType.(*types.PointerType).ElemType, ptr, i)
		return block.NewLoad(ptr.ElemType.(*types.PointerType).ElemType, indexed)
	default:
		panic("Unknown " + strconv.Itoa(int(exp.Type)))
	}

}

func (b *LLVM) generateExpression(exp *parser.Node, block *ir.Block, cf *CompiledFunction) value.Value {
	return b.autoTypeCast(b.generateExpressionRaw(exp, block, cf), types.I64, block)
}

func (b *LLVM) generateFunctionCall(fc parser.FunctionCall, block *ir.Block, cf *CompiledFunction) *ir.InstCall {
	f := b.findFunction(fc.Name)

	arguments := []value.Value{}
	for i := range fc.Arguments {
		arguments = append(arguments, b.autoTypeCast(b.generateExpression(fc.Arguments[i], block, cf), f.Sig.Params[i], block))
	}

	return block.NewCall(f, arguments...)
}

func (b *LLVM) autoTypeCast(source value.Value, target types.Type, block *ir.Block) value.Value {
	if source.Type().Equal(target) {
		return source
	}

	if _, ok := source.Type().(*types.PointerType); ok {
		return block.NewPtrToInt(source, target)
	} else {
		if _, ok := target.(*types.PointerType); ok {
			return block.NewIntToPtr(source, target)
		} else {
			if target.(*types.IntType).BitSize > source.Type().(*types.IntType).BitSize {
				return block.NewZExt(source, target)
			}
			return block.NewTrunc(source, target)
		}
	}
}

func (b *LLVM) newBlock(block *ir.Block) *ir.Block {
	new := block.Parent.NewBlock("")
	return new
}

func (b *LLVM) generateVariableSelfModify(name string, operation parser.NodeType) []*parser.Node {
	return []*parser.Node{
		{
			Type: parser.VARIABLE_ASSIGN,
			A: &parser.Node{
				Type: operation,
				A: &parser.Node{
					Type:  parser.VARIABLE_LOOKUP,
					Value: name,
				},
				B: &parser.Node{
					Type:  parser.NUMBER,
					Value: 1,
				},
			},
			Value: name,
		},
	}
}

func (b *LLVM) generateIf(block *ir.Block, node *parser.Node, iff parser.If, cf *CompiledFunction) *ir.Block {
	ifTrue := b.newBlock(block)
	ifFalse := b.newBlock(block)
	ifAfter := b.newBlock(block)

	x := b.generateExpression(node.A, block, cf)
	cmp := block.NewICmp(enum.IPredNE, x, constant.NewInt(types.I64, 0))
	block.NewCondBr(cmp, ifTrue, ifFalse)

	ifTrue = b.generateCodeBlock(ifTrue, iff.TrueBlock, cf)
	if ifTrue.Term == nil {
		ifTrue.NewBr(ifAfter)
	}

	ifFalse = b.generateCodeBlock(ifFalse, iff.FalseBlock, cf)
	if ifFalse.Term == nil {
		ifFalse.NewBr(ifAfter)
	}

	return ifAfter
}

func (b *LLVM) generateConditionalLoop(block *ir.Block, node *parser.Node, cf *CompiledFunction) *ir.Block {
	loopCompare := b.newBlock(block)
	loopBody := b.newBlock(block)
	loopEnd := b.newBlock(block)

	block.NewBr(loopCompare)

	x := b.generateExpression(node.A, loopCompare, cf)
	cmp := loopCompare.NewICmp(enum.IPredNE, x, constant.NewInt(types.I64, 0))
	loopCompare.NewCondBr(cmp, loopBody, loopEnd)

	loopBody = b.generateCodeBlock(loopBody, node.Value.([]*parser.Node), cf)
	if loopBody.Term == nil {
		loopBody.NewBr(loopCompare)
	}

	return loopEnd
}

func (b *LLVM) generatePostConditionalLoop(block *ir.Block, node *parser.Node, cf *CompiledFunction) *ir.Block {
	loopBody := b.newBlock(block)
	loopEnd := b.newBlock(block)

	block.NewBr(loopBody)

	loopBody = b.generateCodeBlock(loopBody, node.Value.([]*parser.Node), cf)

	x := b.generateExpression(node.A, loopBody, cf)
	cmp := loopBody.NewICmp(enum.IPredNE, x, constant.NewInt(types.I64, 0))
	if loopBody.Term == nil {
		loopBody.NewCondBr(cmp, loopBody, loopEnd)
	}

	return loopEnd
}

func (b *LLVM) generateCodeBlock(block *ir.Block, body []*parser.Node, cf *CompiledFunction) *ir.Block {

	for i := range body {
		node := body[i]

		switch node.Type {
		case parser.VARIABLE_DECLARATION:
			datatype := node.Value.(parser.NamedDatatype)
			v := block.NewAlloca(b.datatypeToLLVM(datatype.UnnamedDatatype))
			v.SetName("local_" + datatype.Name)
			cf.variables[datatype.Name] = v

			if node.A != nil {
				x := b.generateExpression(node.A, block, cf)
				c := b.autoTypeCast(x, v.ElemType, block)
				block.NewStore(c, v)
			}
		case parser.VARIABLE_ASSIGN:
			v, t := b.findVariable(node.Value.(string), cf)
			x := b.generateExpression(node.A, block, cf)
			c := b.autoTypeCast(x, t, block)
			block.NewStore(c, v)
		case parser.VARIABLE_ASSIGN_ARRAY:
			v, t := b.findVariable(node.Value.(string), cf)
			ptr := block.NewLoad(t, v)
			i := b.generateExpression(node.A, block, cf)
			indexed := block.NewGetElementPtr(ptr.ElemType.(*types.PointerType).ElemType, ptr, i)
			x := b.generateExpression(node.B, block, cf)
			c := b.autoTypeCast(x, ptr.ElemType.(*types.PointerType).ElemType, block)
			block.NewStore(c, indexed)
		case parser.VARIABLE_INCREASE:
			block = b.generateCodeBlock(block, b.generateVariableSelfModify(node.Value.(string), parser.ADD), cf)
		case parser.VARIABLE_DECREASE:
			block = b.generateCodeBlock(block, b.generateVariableSelfModify(node.Value.(string), parser.SUBTRACT), cf)
		case parser.FUNCTION_CALL:
			fc := node.Value.(parser.FunctionCall)
			b.generateFunctionCall(fc, block, cf)
		case parser.RETURN:
			if block.Term != nil {
				panic("Block already terminated")
			}

			if node.A != nil {
				x := b.generateExpression(node.A, block, cf)
				c := b.autoTypeCast(x, cf.returnType, block)
				cf.returnIncomings = append(cf.returnIncomings, ir.NewIncoming(c, block))
			}
			block.NewBr(cf.returnBlock)

		case parser.IF:
			block = b.generateIf(block, node, node.Value.(parser.If), cf)
		case parser.CONDITIONAL_LOOP:
			block = b.generateConditionalLoop(block, node, cf)
		case parser.POST_CONDITIONAL_LOOP:
			block = b.generatePostConditionalLoop(block, node, cf)
		case parser.LOOP:
			loopBody := b.newBlock(block)
			block.NewBr(loopBody)

			loopEnd := b.generateCodeBlock(loopBody, node.Value.([]*parser.Node), cf)
			if loopEnd.Term == nil {
				loopEnd.NewBr(loopBody)
			}

			block = b.newBlock(block)
		default:
			panic("Unknown " + strconv.Itoa(int(node.Type)))
		}
	}
	return block
}

func (b *LLVM) generateFunction(f *ir.Func, af parser.Function) *CompiledFunction {
	cf := CompiledFunction{
		variables:       make(map[string]*ir.InstAlloca),
		returnBlock:     nil,
		returnIncomings: []*ir.Incoming{},
		returnType:      f.Sig.RetType,
	}

	declareOnly := false
	noReturn := false

	if utils.IndexOf(af.Attributes, parser.Assembly) >= 0 {
		panic("Unsupported attribute assembly")
	} else if utils.IndexOf(af.Attributes, parser.NoReturn) >= 0 {
		noReturn = true
	} else if utils.IndexOf(af.Attributes, parser.External) >= 0 {
		declareOnly = true
	}

	if declareOnly {
		return &cf
	} else {
		entry := f.NewBlock("entry")
		main := f.NewBlock("body")

		for i := range af.Arguments {
			argument := af.Arguments[i]
			v := entry.NewAlloca(b.datatypeToLLVM(argument.UnnamedDatatype))
			v.SetName("arg_" + argument.Name)
			cf.variables[argument.Name] = v
			entry.NewStore(f.Params[i], v)
		}

		entry.NewBr(main)

		ret := f.NewBlock("return")
		ret.NewUnreachable()

		cf.returnBlock = ret

		main = b.generateCodeBlock(main, af.Body, &cf)
		if main.Term == nil {
			if f.Sig.RetType.Equal(types.Void) {
				main.NewRet(nil)
			} else {
				fmt.Println("[WARNING] no return in non void function")
				main.NewUnreachable()
			}
		}
		if noReturn {
			b.generateFunctionCall(parser.FunctionCall{
				Name:      "unreachable",
				Arguments: []*parser.Node{},
			}, ret, &cf)
		} else {
			if len(cf.returnIncomings) > 0 {
				phi := ret.NewPhi(cf.returnIncomings...)
				ret.NewRet(phi)
			}
		}
	}

	fmt.Println("[DEBUG]", f.Name(), "compiled with", len(f.Sig.Params), "arguments and", len(cf.variables), "local variables")

	return &cf
}

func (b *LLVM) generateFunctionDeclaration(f parser.Function, module *ir.Module) {

	if utils.IndexOf(f.Attributes, parser.Assembly) >= 0 {
		panic("Unsupported attribute assembly")
	}

	parameters := []*ir.Param{}
	for i := range f.Arguments {
		parameters = append(parameters, ir.NewParam(f.Arguments[i].Name, b.datatypeToLLVM(f.Arguments[i].UnnamedDatatype)))
	}

	b.functions[f.Name] = module.NewFunc(f.Name, b.datatypeToLLVM(f.ReturnDatatype), parameters...)

}

func (b *LLVM) Compile() string {
	tmp := b.global.Value.([]*parser.Node)

	b.module = ir.NewModule()
	b.module.TargetTriple = b.target

	for i := range tmp {
		switch tmp[i].Type {
		case parser.VARIABLE_DECLARATION:
			datatype := tmp[i].Value.(parser.NamedDatatype)
			d := b.datatypeToLLVM(datatype.UnnamedDatatype)

			var global *ir.Global

			if tmp[i].A != nil {
				if datatype.IsArray {
					panic("Global array initializers not supported!")
				}
				if tmp[i].A.Type == parser.STRING {
					s := b.module.NewGlobalDef(datatype.Name+".init", constant.NewCharArrayFromString(tmp[i].A.Value.(string)+"\x00"))
					global = b.module.NewGlobalDef(datatype.Name, constant.NewIntToPtr(constant.NewPtrToInt(s, types.I64), d))
				} else if tmp[i].A.Type == parser.NUMBER {
					global = b.module.NewGlobalDef(datatype.Name, constant.NewInt(d.(*types.IntType), int64(tmp[i].A.Value.(int))))
				} else {
					panic("Only string and number are supported for globals!")
				}
			} else {
				switch d := d.(type) {
				case *types.PointerType:
					global = b.module.NewGlobalDef(datatype.Name, constant.NewIntToPtr(constant.NewInt(types.I64, 0), d))
				case *types.IntType:
					global = b.module.NewGlobalDef(datatype.Name, constant.NewInt(d, 0))
				default:
					panic("?")
				}
			}

			b.globalVariables[datatype.Name] = global
		}
	}

	for i := range tmp {
		switch tmp[i].Type {
		case parser.FUNCTION:
			b.generateFunctionDeclaration(tmp[i].Value.(parser.Function), b.module)
		}
	}

	for i := range tmp {
		switch tmp[i].Type {
		case parser.FUNCTION:
			b.generateFunction(b.findFunction(tmp[i].Value.(parser.Function).Name), tmp[i].Value.(parser.Function))
		}
	}

	return b.module.String()
}
