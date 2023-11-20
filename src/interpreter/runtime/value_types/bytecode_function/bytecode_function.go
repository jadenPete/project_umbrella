package bytecode_function

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/built_in_definitions"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_util"
)

type BytecodeFunctionEvaluator struct {
	ContainingScope *scope
	BlockGraph      *runtime.BytecodeFunctionBlockGraph
}

// TODO: Make this concurrent
func (evaluator *BytecodeFunctionEvaluator) Evaluator(
	runtime_ *runtime.Runtime,
	arguments ...value.Value,
) value.Value {
	firstValueID := 0

	if evaluator.ContainingScope != nil {
		firstValueID = evaluator.BlockGraph.FirstValueID
	}

	scope := &scope{
		parent:       evaluator.ContainingScope,
		firstValueID: firstValueID,
		values:       map[int]value.Value{},
	}

	for i, argument := range arguments {
		scope.values[scope.firstValueID+i] = argument
	}

	isAcyclic := evaluator.BlockGraph.Evaluate(func(i int) {
		callArguments := []value.Value{}

		switch node := evaluator.BlockGraph.Nodes[i].(type) {
		case *runtime.BytecodeFunctionBlockGraph:
			scope.values[node.ValueID] = NewBytecodeFunction(
				node.ParameterCount,
				&BytecodeFunctionEvaluator{
					ContainingScope: scope,
					BlockGraph:      node,
				},
			)

		case runtime.InstructionList:
			for _, element := range node {
				switch element.Instruction.Type {
				case bytecode_generator.PushArgumentInstruction:
					callArguments =
						append(callArguments, scope.getValue(element.Instruction.Arguments[0]))

				case bytecode_generator.ValueCopyInstruction:
					scope.values[element.InstructionValueID] =
						scope.getValue(element.Instruction.Arguments[0])

				case bytecode_generator.ValueFromCallInstruction:
					function_, ok := scope.getValue(element.Instruction.Arguments[0]).(*function.Function)

					if !ok {
						errors.RaiseError(runtime_errors.NonFunctionCalled)
					}

					scope.values[element.InstructionValueID] =
						function_.Evaluate(runtime_, callArguments...)

				case bytecode_generator.ValueFromConstantInstruction:
					scope.values[element.InstructionValueID] =
						runtime_.Constants[element.Instruction.Arguments[0]]

				case bytecode_generator.ValueFromStructValueInstruction:
					value_ := scope.getValue(element.Instruction.Arguments[0])
					fieldNameConstant := runtime_.Constants[element.Instruction.Arguments[1]]
					fieldNameValue, ok := fieldNameConstant.(value_types.StringValue)

					if !ok {
						errors.RaiseError(runtime_errors.NonStringFieldName)
					}

					fieldName := fieldNameValue.Content
					selectType := parser_types.SelectType(element.Instruction.Arguments[2])

					scope.values[element.InstructionValueID] =
						value_util.LookupField(runtime_, value_, fieldName, selectType)
				}
			}
		}
	})

	if !isAcyclic {
		errors.RaiseError(runtime_errors.ValueCycle)
	}

	if len(scope.values) == 0 {
		errors.RaiseError(runtime_errors.EmptyFunctionBlockGraph)
	}

	lastValueID := 0

	for valueID := range scope.values {
		if valueID > lastValueID {
			lastValueID = valueID
		}
	}

	return scope.values[lastValueID]
}

type scope struct {
	parent       *scope
	firstValueID int
	values       map[int]value.Value
}

func (scope_ *scope) getValue(valueID int) value.Value {
	if builtInValue, ok :=
		built_in_definitions.BuiltInValues[built_in_declarations.BuiltInValueID(valueID)]; ok {
		return builtInValue
	}

	currentScope := scope_

	for currentScope.firstValueID > valueID {
		currentScope = currentScope.parent
	}

	return currentScope.values[valueID]
}

func NewBytecodeFunction(
	parameterCount int,
	evaluator *BytecodeFunctionEvaluator,
) *function.Function {
	name := "(function)"

	return &function.Function{
		FunctionEvaluator: evaluator,
		ArgumentValidator: function.NewFixedFunctionArgumentValidator(
			name,
			common.Repeat[reflect.Type](nil, parameterCount)...,
		),

		Name:  name,
		Type_: parser_types.NormalFunction,
	}
}
